// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"fmt"
	"testing"

	"github.com/dateiexplorer/dhbw-attendancelist/internal/journal"
	"github.com/dateiexplorer/dhbw-attendancelist/internal/timeutil"
	"github.com/stretchr/testify/assert"
)

func TestPrintVisitedLocationsForPerson(t *testing.T) {
	j, err := journal.ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	msg, err := printVisitedLocationsForPerson(j, "Max")
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintln("DHBW Mosbach"), msg)

	msg, err = printVisitedLocationsForPerson(j, "Hans,Müller")
	assert.NoError(t, err)

	expected := []string{"DHBW Mosbach", "Alte Mälzerei"}

	for _, location := range expected {
		assert.Contains(t, msg, location)
	}

}

func TestPrintVisitedLocationsForPersonMultiplePersons(t *testing.T) {
	j, err := journal.ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	msg, err := printVisitedLocationsForPerson(j, "Müller")
	assert.Equal(t, "", msg)
	assert.Error(t, err)
}

func TestPrintVisitedLocationsForPersonNoPersonFound(t *testing.T) {
	j, err := journal.ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	msg, err := printVisitedLocationsForPerson(j, "")
	assert.Equal(t, "", msg)
	assert.Error(t, err)
}

func TestPrintContactsForPerson(t *testing.T) {
	j, err := journal.ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	msg, err := printContactsForPerson(j, "Max,Mustermann", "")
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintln("Output successfully written."), msg)
}

func TestCreateAttendanceListForLocation(t *testing.T) {
	j, err := journal.ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	msg, err := createAttendanceListForLocation(j, "DHBW Mosbach", "")
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintln("Output successfully written."), msg)
}

func TestGetMatchingPersonsFromJournal(t *testing.T) {
	j, err := journal.ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	persons := getMatchingPersonsFromJournal(j, "Max")
	expected := []journal.Person{
		journal.NewPerson("Max", "Mustermann", "Musterstraße", "20", "74821", "Mosbach"),
	}

	assert.Equal(t, expected, persons)
}

func TestGetMatchingPersonsFromJournalMultiplePersons(t *testing.T) {
	j, err := journal.ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	persons := getMatchingPersonsFromJournal(j, "Müller")
	expected := []journal.Person{
		journal.NewPerson("Hans", "Müller", "Feldweg", "12", "74722", "Buchen"),
		journal.NewPerson("Lieschen", "Müller", "Lindenstraße", "15", "10115", "Berlin"),
	}

	assert.Equal(t, expected, persons)
}

func TestGetMatchingPersonsFromJournalMutlipleAttributesRandomOrder(t *testing.T) {
	j, err := journal.ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	persons := getMatchingPersonsFromJournal(j, "Hans,Müller")
	expected := []journal.Person{
		journal.NewPerson("Hans", "Müller", "Feldweg", "12", "74722", "Buchen"),
	}

	assert.Equal(t, expected, persons)

	persons = getMatchingPersonsFromJournal(j, "Müller,Hans")
	assert.Equal(t, expected, persons)

	persons = getMatchingPersonsFromJournal(j, "Müller,Feldweg,12,Buchen")
	assert.Equal(t, expected, persons)
}
