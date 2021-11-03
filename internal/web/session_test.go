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
	"testing"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
	"github.com/stretchr/testify/assert"
)

// Data

var sessions = []Session{
	{"aabbccddee", "userHash1", "DHBW Mosbach"},
	{"ffgghhiijj", "userHash2", "Alte Mälzerei"},
	{"kkllmmnnoo", "userHash3", "DHBW Mosbach"},
}

// Functions

func TestNewSession(t *testing.T) {
	expected := Session{"aabbccddee", "userHash", "DHBW Mosbach"}
	actual := NewSession("aabbccddee", "userHash", "DHBW Mosbach")

	assert.Equal(t, expected, actual)
}

func TestRunOpenSessions(t *testing.T) {
	openSessions := new(OpenSessions)

	sessionQueue := make(chan sessionQueueItem)
	journalWriter := make(chan journal.JournalEntry)

	// Run goroutine in background
	openSessions.run(sessionQueue, journalWriter)

	timestamp := timeutil.Now()
	location := journal.Location("DHBW Mosbach")
	person := journal.NewPerson("Max", "Mustermann", "Musterstaße", "20", "74821", "Mosbach")
	session := NewSession("aabbccddee", "hash", location)

	// Open Session
	sessionQueue <- sessionQueueItem{journal.Login, timestamp, &session, &person}

	// If journalWriter receives entry, value is stored
	entry := <-journalWriter
	journalEntryLogin := journal.NewJournalEntry(timestamp, session.ID, journal.Login, location, person)
	assert.Equal(t, journalEntryLogin, entry)

	value, ok := openSessions.Load(session.UserHash)
	assert.True(t, ok)

	actual, ok := value.(*Session)
	assert.True(t, ok)
	assert.Equal(t, session, *actual)

	// Close Session
	sessionQueue <- sessionQueueItem{journal.Logout, timestamp, &session, &person}
	entry = <-journalWriter
	journalEntryLogout := journal.NewJournalEntry(timestamp, session.ID, journal.Logout, location, person)

	value, ok = openSessions.Load(session.UserHash)
	assert.False(t, ok)
	assert.Nil(t, value)

	assert.Equal(t, journalEntryLogout, entry)
}

func TestOpenSession(t *testing.T) {
	timestamp := timeutil.Now()
	location := journal.Location("DHBW Mosbach")
	person := journal.NewPerson("Max", "Mustermann", "Musterstaße", "20", "74821", "Mosbach")
	hash, err := Hash(person, "privServerSecret")
	assert.NoError(t, err)

	session := NewSession("aabbccddee", hash, location)

	expected := sessionQueueItem{journal.Login, timestamp, &session, &person}

	sessionIDs := make(chan journal.SessionIdentifier, 1)
	sessionIDs <- "aabbccddee"

	actual := OpenSession(sessionIDs, timestamp, &person, location, "privServerSecret")

	assert.Equal(t, expected, actual)
}

func TestCloseSession(t *testing.T) {
	timestamp := timeutil.Now()
	location := journal.Location("DHBW Mosbach")
	person := journal.NewPerson("Max", "Mustermann", "Musterstaße", "20", "74821", "Mosbach")
	hash, err := Hash(person, "privServerSecret")
	assert.NoError(t, err)

	session := NewSession("aabbccddee", hash, location)

	expected := sessionQueueItem{journal.Logout, timestamp, &session, &person}

	actual := CloseSession(timestamp, &session, &person)

	assert.Equal(t, expected, actual)
}

func TestGetSessionForUser(t *testing.T) {
	openSessions := new(OpenSessions)

	for i := 0; i < len(sessions); i++ {
		openSessions.Store(sessions[i].UserHash, &sessions[i])
	}

	actual, ok := openSessions.GetSessionForUser("userHash1")

	assert.True(t, ok)
	assert.Equal(t, &sessions[0], actual)
}

func TestGetSessionForUserNotFound(t *testing.T) {
	openSessions := new(OpenSessions)

	for i := 0; i < len(sessions); i++ {
		openSessions.Store(sessions[i].UserHash, &sessions[i])
	}

	actual, ok := openSessions.GetSessionForUser("notFound")

	assert.False(t, ok)
	assert.Nil(t, actual)
}

func TestRunSessionManager(t *testing.T) {
	journalWriter := make(chan journal.JournalEntry)
	openSessions, sessionQueueItem, sessionIdentifier := RunSessionManager(journalWriter, 10)

	timestamp := timeutil.Now()
	location := journal.Location("DHBW Mosbach")
	person := journal.NewPerson("Max", "Mustermann", "Musterstaße", "20", "74821", "Mosbach")
	hash, err := Hash(person, "privServerSecret")
	assert.NoError(t, err)

	sessionQueueItem <- OpenSession(sessionIdentifier, timestamp, &person, location, "privServerSecret")
	entry := <-journalWriter

	value, ok := openSessions.Load(hash)
	assert.True(t, ok)

	actual := value.(*Session)
	assert.Equal(t, entry.Session, actual.ID)
	assert.Equal(t, entry.Location, actual.Location)
	assert.Equal(t, hash, actual.UserHash)
}
