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

	"github.com/dateiexplorer/dhbw-attendancelist/internal/journal"
	"github.com/dateiexplorer/dhbw-attendancelist/internal/timeutil"
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
	session := NewSession("aabbccddee", "userHash", "DHBW Mosbach")

	assert.Equal(t, expected, session)
}

func TestOpenSession(t *testing.T) {
	ts := timeutil.Now()
	loc := journal.Location("DHBW Mosbach")
	p := journal.NewPerson("Max", "Mustermann", "Musterstaße", "20", "74821", "Mosbach")
	hash, err := Hash(p, "privServerSecret")
	assert.NoError(t, err)

	session := NewSession("aabbccddee", hash, loc)
	expected := SessionQueueItem{journal.Login, ts, &session, &p}

	sessionIDs := make(chan string, 1)
	sessionIDs <- "aabbccddee"

	actual := OpenSession(sessionIDs, ts, &p, loc, "privServerSecret")
	assert.Equal(t, expected, actual)
}

func TestCloseSession(t *testing.T) {
	ts := timeutil.Now()
	loc := journal.Location("DHBW Mosbach")
	p := journal.NewPerson("Max", "Mustermann", "Musterstaße", "20", "74821", "Mosbach")
	hash, err := Hash(p, "privServerSecret")
	assert.NoError(t, err)

	session := NewSession("aabbccddee", hash, loc)

	expected := SessionQueueItem{journal.Logout, ts, &session, &p}

	actual := CloseSession(ts, &session, &p)
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

	actual, ok := openSessions.GetSessionForUser("noUser")
	assert.False(t, ok)
	assert.Nil(t, actual)
}

func TestRunSessionManager(t *testing.T) {
	journalWriter := make(chan journal.JournalEntry)
	openSessions, sessionQueue, sessionIdentifier := RunSessionManager(journalWriter, 10)

	ts := timeutil.Now()
	loc := journal.Location("DHBW Mosbach")
	p := journal.NewPerson("Max", "Mustermann", "Musterstaße", "20", "74821", "Mosbach")
	hash, err := Hash(p, "privServerSecret")
	assert.NoError(t, err)

	// OpenSession
	sessionQueue <- OpenSession(sessionIdentifier, ts, &p, loc, "privServerSecret")
	entry := <-journalWriter
	assert.Equal(t, journal.Login, entry.Event)
	assert.Equal(t, entry.Timestamp, ts)
	assert.Equal(t, entry.Person, p)

	value, ok := openSessions.Load(hash)
	assert.True(t, ok)

	actual := value.(*Session)
	assert.Equal(t, entry.SessionID, actual.ID)
	assert.Equal(t, entry.Location, actual.Location)
	assert.Equal(t, hash, actual.UserHash)

	// CloseSession
	sessionQueue <- CloseSession(ts, actual, &p)
	entry = <-journalWriter
	assert.Equal(t, journal.Logout, entry.Event)
	assert.Equal(t, entry.Timestamp, ts)
	assert.Equal(t, entry.Person, p)

	value, ok = openSessions.Load(hash)
	assert.False(t, ok)
	assert.Nil(t, value)
}
