package journal

import (
	"os"
	"testing"

	"github.com/dateiexplorer/attendancelist/internal/timeutil"
	"github.com/stretchr/testify/assert"
)

// Data

var locations = map[string]Location{
	"DHBW Mosbach":  {"DHBW Mosbach"},
	"Alte Mälzerei": {"Alte Mälzerei"},
}

var persons = map[string]Person{
	"Hans Müller":            {"Hans", "Müller", Address{"Feldweg", "12", "74722", "Buchen"}},
	"Gisela Musterfrau":      {"Gisela", "Musterfrau", Address{"Musterstraße", "10", "74821", "Mosbach"}},
	"Max Mustermann":         {"Max", "Mustermann", Address{"Musterstraße", "20", "74821", "Mosbach"}},
	"Anne Meier":             {"Anne", "Meier", Address{"Hauptstraße", "18", "74821", "Mosbach"}},
	"Lieschen Müller":        {"Lieschen", "Müller", Address{"Lindenstraße", "15", "10115", "Berlin"}},
	"Otto Normalverbraucher": {"Otto", "Normalverbraucher", Address{"Dieselstraße", "52", "70376", "Stuttgart"}},
}

// Functions

// 8.1) Read journal file for a specific date
func TestReadJournal(t *testing.T) {
	expected := Journal{timeutil.NewDate(2021, 10, 15), []journalEntry{
		{timeutil.NewTimestamp(2021, 10, 15, 6, 20, 13), "d61ec70b78628e15", Login, locations["DHBW Mosbach"], persons["Hans Müller"]},
		{timeutil.NewTimestamp(2021, 10, 15, 9, 15, 20), "989ce491d5df53c9", Login, locations["DHBW Mosbach"], persons["Gisela Musterfrau"]},
		{timeutil.NewTimestamp(2021, 10, 15, 12, 15, 30), "f797f342aebab436", Login, locations["DHBW Mosbach"], persons["Max Mustermann"]},
		{timeutil.NewTimestamp(2021, 10, 15, 12, 17, 20), "1ce7549a51133e9f", Login, locations["DHBW Mosbach"], persons["Anne Meier"]},
		{timeutil.NewTimestamp(2021, 10, 15, 13, 30, 0), "68e7faee906ffd4c", Login, locations["DHBW Mosbach"], persons["Lieschen Müller"]},
		{timeutil.NewTimestamp(2021, 10, 15, 13, 40, 10), "d61ec70b78628e15", Logout, locations["DHBW Mosbach"], persons["Hans Müller"]},
		{timeutil.NewTimestamp(2021, 10, 15, 13, 40, 11), "5faacdf0e6e7b44a", Login, locations["Alte Mälzerei"], persons["Hans Müller"]},
		{timeutil.NewTimestamp(2021, 10, 15, 15, 42, 23), "68e7faee906ffd4c", Logout, locations["DHBW Mosbach"], persons["Lieschen Müller"]},
		{timeutil.NewTimestamp(2021, 10, 15, 16, 48, 21), "f797f342aebab436", Logout, locations["DHBW Mosbach"], persons["Max Mustermann"]},
		{timeutil.NewTimestamp(2021, 10, 15, 16, 52, 0), "989ce491d5df53c9", Logout, locations["DHBW Mosbach"], persons["Gisela Musterfrau"]},
		{timeutil.NewTimestamp(2021, 10, 15, 17, 15, 22), "1ce7549a51133e9f", Logout, locations["DHBW Mosbach"], persons["Anne Meier"]},
		{timeutil.NewTimestamp(2021, 10, 15, 17, 32, 45), "848dc86c0b5e62a0", Login, locations["Alte Mälzerei"], persons["Otto Normalverbraucher"]},
		{timeutil.NewTimestamp(2021, 10, 15, 19, 15, 12), "848dc86c0b5e62a0", Logout, locations["Alte Mälzerei"], persons["Otto Normalverbraucher"]},
	}}

	actual, err := ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestReadNotExistingJournal(t *testing.T) {
	expected := Journal{timeutil.NewDate(2020, 9, 4), []journalEntry{}}

	actual, err := ReadJournal("testdata", timeutil.NewDate(2020, 9, 4))

	assert.Error(t, err)
	assert.Equal(t, expected, actual)
}

func TestReadMalformedJournal(t *testing.T) {
	expected := []Journal{
		{timeutil.NewDate(2020, 1, 1), []journalEntry{}},
		{timeutil.NewDate(2020, 1, 2), []journalEntry{}},
	}

	for _, journal := range expected {
		actual, err := ReadJournal("testdata", journal.date)

		assert.NotErrorIs(t, err, os.ErrNotExist)
		assert.Error(t, err)
		assert.Equal(t, journal, actual)
	}
}

// 8.2) Get all locations for a specific person
func TestGetVisitedLocationsForPerson(t *testing.T) {
	expected := []Location{
		locations["DHBW Mosbach"],
	}

	journal, err := ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	actual := journal.GetVisitedLocationsForPerson(persons["Max Mustermann"])

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func TestGetVisitedLocationForPersonMultipleLocations(t *testing.T) {
	expected := []Location{
		locations["DHBW Mosbach"],
		locations["Alte Mälzerei"],
	}

	journal, err := ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	actual := journal.GetVisitedLocationsForPerson(persons["Hans Müller"])

	assert.NotNil(t, actual)
	for _, location := range expected {
		assert.Contains(t, actual, location)
	}
}

func TestGetVisitedLocationsForPersonNotExistingPerson(t *testing.T) {
	expected := []Location{}

	person := Person{"Susi", "Sorglos", Address{"Musterstraße", "20", "72327", "Musterstadt"}}
	journal, err := ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	actual := journal.GetVisitedLocationsForPerson(person)

	assert.Equal(t, expected, actual)
}

// 8.3) Get attendance list for a specific location as CSV
func TestGetAttendanceEntriesForLocation(t *testing.T) {
	expected := AttendanceList{
		NewAttendanceEntry(persons["Hans Müller"], timeutil.NewTimestamp(2021, 10, 15, 13, 40, 11), timeutil.InvalidTimestamp),
		NewAttendanceEntry(persons["Otto Normalverbraucher"], timeutil.NewTimestamp(2021, 10, 15, 17, 32, 45), timeutil.NewTimestamp(2021, 10, 15, 19, 15, 12)),
	}

	journal, err := ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	actual := journal.GetAttendanceListForLocation(locations["Alte Mälzerei"])

	assert.Equal(t, expected, actual)
}

func TestGetAttendanceListForLocationNotExistingLocation(t *testing.T) {
	expected := AttendanceList{}

	journal, err := ReadJournal("testdata", timeutil.NewDate(2021, 10, 15))
	assert.NoError(t, err)

	actual := journal.GetAttendanceListForLocation(Location{"Night Club"})

	assert.Equal(t, expected, actual)
}

func TestNextEntry(t *testing.T) {
	expected := [][]string{
		{"Hans", "Müller", "Feldweg", "12", "74722", "Buchen", "13:40:11", ""},
		{"Otto", "Normalverbraucher", "Dieselstraße", "52", "70376", "Stuttgart", "17:32:45", "19:15:12"},
	}

	list := AttendanceList{
		NewAttendanceEntry(persons["Hans Müller"], timeutil.NewTimestamp(2021, 10, 15, 13, 40, 11), timeutil.InvalidTimestamp),
		NewAttendanceEntry(persons["Otto Normalverbraucher"], timeutil.NewTimestamp(2021, 10, 15, 17, 32, 45), timeutil.NewTimestamp(2021, 10, 15, 19, 15, 12)),
	}

	counter := 0
	for actual := range list.NextEntry() {
		assert.Equal(t, expected[counter], actual)
		counter++
	}

	assert.Equal(t, len(list), counter)
}

func TestHeader(t *testing.T) {
	expected := []string{"FirstName", "LastName", "Street", "Number", "ZipCode", "City", "Login", "Logout"}

	list := AttendanceList{}
	actual := list.Header()

	assert.Equal(t, expected, actual)
}

func TestNewPerson(t *testing.T) {
	expected := Person{"Max", "Mustermann", Address{"Musterstraße", "20", "74821", "Mosbach"}}

	actual := NewPerson("Max", "Mustermann", "Musterstraße", "20", "74821", "Mosbach")

	assert.Equal(t, expected, actual)
}
