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

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
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
func RunSessionManager(journalWriter chan<- journal.JournalEntry, idLength int) (*OpenSessions, chan<- sessionQueueItem, <-chan journal.SessionIdentifier) {
	maxConcurrentRequests := cap(journalWriter)

	openSessions := new(OpenSessions)
	sessionQueue := make(chan sessionQueueItem, maxConcurrentRequests)

	randomIDs := RandIDGenerator(idLength, maxConcurrentRequests)
	sessionIDs := make(chan journal.SessionIdentifier, maxConcurrentRequests)

	// Convert string to journal.SessionIdentifier
	go func() {
		for id := range randomIDs {
			sessionIDs <- journal.SessionIdentifier(id)
		}
	}()

	openSessions.run(sessionQueue, journalWriter)
	return openSessions, sessionQueue, sessionIDs
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

// run starts an internal session manager in a separate goroutine, which takes the
// sessionQueueItems from the sessinQueueItem channel and performs the specified action.
// Furthermore the function writes each event to the jorunalWriter channel, which can
// be used to log session events such as new logins or logouts.
func (m *OpenSessions) run(sessionQueue <-chan sessionQueueItem, journalWriter chan<- journal.JournalEntry) {
	go func() {
		for item := range sessionQueue {
			switch item.action {
			case journal.Login:
				m.Store(item.session.UserHash, item.session)
			case journal.Logout:
				m.Delete(item.session.UserHash)
			}

			// Write the event to the jorunalWriter.
			journalWriter <- journal.NewJournalEntry(item.timestamp, item.session.ID, item.action, item.session.Location, *item.person)
		}
	}()
}

// A Session represents a user session with a unique random identifier, a unique
// user hash value and an associated Location
type Session struct {
	ID       journal.SessionIdentifier
	UserHash string
	Location journal.Location
}

// NewSession returns a new Session struct.
func NewSession(id journal.SessionIdentifier, userHash string, loc journal.Location) Session {
	return Session{id, userHash, loc}
}

// A sessionQueueItem represents a Item which is consumed by the session manager.
// It is private and only accessible throw the public functions OpenSession and
// CloseSession.
type sessionQueueItem struct {
	action    journal.Event
	timestamp timeutil.Timestamp
	session   *Session
	person    *journal.Person
}

// OpenSession returns a sessionQueueItem which initiates to open a new Session
// for the specific person associated with the location loc.
func OpenSession(sessionIDs <-chan journal.SessionIdentifier, timestamp timeutil.Timestamp, person *journal.Person, loc journal.Location, privkey string) sessionQueueItem {
	hash, _ := Hash(*person, privkey)
	session := NewSession(<-sessionIDs, hash, loc)
	return sessionQueueItem{journal.Login, timestamp, &session, person}
}

// CloseSession returns a sessionQueueItem which initiates to close the
// given session.
func CloseSession(timestamp timeutil.Timestamp, session *Session, person *journal.Person) sessionQueueItem {
	return sessionQueueItem{journal.Logout, timestamp, session, person}
}
