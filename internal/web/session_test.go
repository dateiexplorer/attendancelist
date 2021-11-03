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
