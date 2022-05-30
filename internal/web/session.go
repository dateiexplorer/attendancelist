// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

// Package web provides all functionality which is necessary for the
// service communication, such as cookies or user session management.
package web

import (
	"sync"

	"github.com/dateiexplorer/dhbw-attendancelist/internal/journal"
	"github.com/dateiexplorer/dhbw-attendancelist/internal/timeutil"
)

// RunSessionManager starts a concurrent goroutine which handles sessions.
// The idLength defines the length of a SessionIdentifier, the journalWriter channel
// is a write only channel. The channel is used to write session events to it.
// It can be used to communicate the session events to another service.
//
// Note that the buffer of a the journalWriter defines how many concurrent
// session events can be processed, this means how many login and logouts can be
// performed concurrently without waiting.
//
// The function returns a OpenSession map a read only channel for sessionQueueItems
// and a random SessionIdentifier generator.
func RunSessionManager(journalWriter chan<- journal.JournalEntry, idLength int) (*OpenSessions, chan<- SessionQueueItem, <-chan string) {
	maxConcurrentRequests := cap(journalWriter)

	openSessions := new(OpenSessions)
	sessionQueue := make(chan SessionQueueItem, maxConcurrentRequests)
	ids := RandIDGenerator(idLength, maxConcurrentRequests)
	go func() {
		for item := range sessionQueue {
			switch item.Action {
			case journal.Login:
				openSessions.Store(item.Session.UserHash, item.Session)
			case journal.Logout:
				openSessions.Delete(item.Session.UserHash)
			}

			// Write the event to the jorunalWriter.
			journalWriter <- journal.NewJournalEntry(item.Timestamp, item.Session.ID, item.Action, item.Session.Location, *item.Person)
		}
	}()

	return openSessions, sessionQueue, ids
}

// An OpenSessions map holds all open Sessions.
type OpenSessions struct {
	sync.Map
}

// GetSessionForUser returns a pointer to the Session stored in the OpenSessions map
// for a specific user defined by a unique userHash.
// It returns the pointer to a session and an ok value if a session was found.
// Note that if no session found, session is nil.
func (m *OpenSessions) GetSessionForUser(userHash string) (session *Session, ok bool) {
	m.Range(func(key, value interface{}) bool {
		val := value.(*Session)
		if val.UserHash == userHash {
			session = val
			ok = true
			return false
		}

		ok = false
		return true
	})

	return session, ok
}

// A Session represents a user session with a unique random identifier, a unique
// user hash value and an associated Location
type Session struct {
	ID       string
	UserHash string
	Location journal.Location
}

// NewSession returns a new Session struct.
func NewSession(id string, userHash string, loc journal.Location) Session {
	return Session{id, userHash, loc}
}

// A SessionQueueItem represents a Item which is consumed by the session manager.
// It is private and only accessible throw the public functions OpenSession and
// CloseSession.
type SessionQueueItem struct {
	Action    journal.Event
	Timestamp timeutil.Timestamp
	Session   *Session
	Person    *journal.Person
}

// OpenSession returns a SessionQueueItem which initiates to open a new Session
// for the specific person associated with the location loc.
func OpenSession(sessionIDs <-chan string, timestamp timeutil.Timestamp, person *journal.Person, loc journal.Location, privkey string) SessionQueueItem {
	hash, _ := Hash(*person, privkey)
	session := NewSession(<-sessionIDs, hash, loc)
	return SessionQueueItem{journal.Login, timestamp, &session, person}
}

// CloseSession returns a sessionQueueItem which initiates to close the
// given session.
func CloseSession(timestamp timeutil.Timestamp, session *Session, person *journal.Person) SessionQueueItem {
	return SessionQueueItem{journal.Logout, timestamp, session, person}
}
