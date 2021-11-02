package secure

import (
	"sync"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
)

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

type OpenSessions struct {
	sync.Map
}

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

type Session struct {
	ID       journal.SessionIdentifier
	UserHash string
	Location journal.Location
}

func NewSession(id journal.SessionIdentifier, userHash string, loc journal.Location) Session {
	return Session{id, userHash, loc}
}

func (m *OpenSessions) run(sessionQueue <-chan sessionQueueItem, journalWriter chan<- journal.JournalEntry) {
	go func() {
		for item := range sessionQueue {
			switch item.action {
			case journal.Login:
				m.Store(item.session.UserHash, item.session)
			case journal.Logout:
				m.Delete(item.session.UserHash)
			}

			journalWriter <- journal.NewJournalEntry(item.timestamp, item.session.ID, item.action, item.session.Location, *item.person)
		}
	}()
}

type sessionQueueItem struct {
	action    journal.Event
	timestamp timeutil.Timestamp
	session   *Session
	person    *journal.Person
}

func OpenSession(sessionIDs <-chan journal.SessionIdentifier, person *journal.Person, loc journal.Location, privkey string) sessionQueueItem {
	hash, _ := Hash(*person, privkey)
	session := NewSession(<-sessionIDs, hash, loc)
	return sessionQueueItem{journal.Login, timeutil.Now(), &session, person}
}

func CloseSession(session *Session, person *journal.Person) sessionQueueItem {
	return sessionQueueItem{journal.Logout, timeutil.Now(), session, person}
}
