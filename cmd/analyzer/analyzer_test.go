package analyzer

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Data

const locations = map[string]Location{
	"DHBW Mosbach":  Location{"DHBW Mosbach"},
	"Alte Mälzerei": Location{"Alte Mälzerei"},
}

const persons = map[string]Person{
	"Hans Müller":            Person{"Hans", "Müller", Address{"Feldweg", "12", "74722", "Buchen"}},
	"Gisela Musterfrau":      Person{"Gisela", "Musterfrau", Address{"Musterstraße", "10", "74821", "Mosbach"}},
	"Max Mustermann":         Person{"Max", "Mustermann", Address{"Musterstraße", "20", "74821", "Mosbach"}},
	"Anne Meier":             Person{"Anne", "Meier", Address{"Haupstraße", "18", "74821", "Mosbach"}},
	"Lieschen Müller":        Person{"Lieschen", "Müller", Address{"Lindenstraße", "15", "10115", "Berlin"}},
	"Otto Normalverbraucher": Person{"Otto", "Normalverbraucher", Address{"Dieselstraße", "52", "70376", "Stuttgart"}},
}

// Functions

func TestNewSimpleDate(t *testing.T) {
	year, month, day := time.Date(2021, time.October, 15, 0, 0, 0, 0, time.UTC).Date()
	expected := SimpleDate{year, month, day}

	actual := NewSimpleDate(2021, time.October, 15)

	assert.Equal(t, expected, actual)
}

func TestTime(t *testing.T) {
	expected := "15:30:25"

	date := time.Date(2021, 10, 15, 15, 30, 25, 20, time.UTC)
	actual := date.Time(date)

	assert.Equal(t, expected, actual)
}

func TestParseTimestamp(t *testing.T) {
	expected := time.Date(2021, 10, 15, 15, 30, 25, 20, time.UTC)

	actual, err := ParseTimestamp("2021-10-15 15:30:25")

	assert.NoError(err)
	assert.Equal(t, expected, actual)
}

func TestParseTimestampFailed(t *testing.T) {
	expected := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)

	actual, err := ParseTimestamp("2021.10.15 15:30:25")

	assert.Error(err)
	assert.Equal(t, expected, actual)
}

func TestParseDate(t *testing.T) {
	expected := SimpleDate{2021, time.October, 15}

	inputs := []string{"15.10.2021"}
	for input := range inputs {
		actual, err := ParseDate(input)

		assert.NoError(err)
		assert.Equal(t, expected, actual)
	}
}

func TestParseDateFailed(t *testing.T) {
	expected := SimpleDate{0, 0, 0}

	actual, err := ParseDate("some input")

	assert.Error(err)
	assert.Equal(actual, expected)
}

func TestNextEntry(t *testing.T) {
	expected := [][]string{
		[]string{"Hans", "Müller", "Feldweg", "12", "74722", "Buchen", "13:40:11", ""},
		[]string{"Otto", "Normalverbraucher", "Dieselstraße", "52", "70376", "Stuttgart", "17:32:45", "19:15:12"},
	}

	list := AttendanceList{
		AttendanceEntry{persons["Hans Müller"], ParseTimestamp("2021-10-15 13:40:11"), nil},
		AttendanceEntry{persons["Otto Normalverbraucher"], ParseTimestamp("2021-10-15 17:32:45"), ParseTimestamp("2021-10-15 19:15:12")},
	}

	counter := 0
	for i, actual := range list.NextEntry() {
		assert.Equal(expected[i], actual)
		counter++
	}

	assert.Equal(len(list), counter)
}

func TestHeader(t *testing.T) {
	expected := []string{"FirstName", "LastName", "Street", "Number", "ZipCode", "City", "Login", "Logout"}

	list := AttendanceList{}
	actual := list.Header()

	assert.Equal(t, expected, actual)
}

// 8.1) Read journal file for a specific date
func TestReadJournal(t *testing.T) {
	expected := Journal{SimpleDate{2021, 10, 15}, []JournalEntry{
		JournalEntry{ParseTimestamp("2021-10-15 06:20:13"), Login, locations["DHBW Mosbach"], persons["Hans Müller"]},
		JournalEntry{ParseTimestamp("2021-10-15 09:15:20"), Login, locations["DHBW Mosbach"], persons["Gisela Musterfrau"]},
		JournalEntry{ParseTimestamp("2021-10-15 12:15:30"), Login, locations["DHBW Mosbach"], persons["Max Mustermann"]},
		JournalEntry{ParseTimestamp("2021-10-15 12:17:20"), Login, locations["DHBW Mosbach"], persons["Anne Meier"]},
		JournalEntry{ParseTimestamp("2021-10-15 13:30:00"), Login, locations["DHBW Mosbach"], persons["Lieschen Müller"]},
		JournalEntry{ParseTimestamp("2021-10-15 13:40:10"), Logout, locations["DHBW Mosbach"], persons["Hans Müller"]},
		JournalEntry{ParseTimestamp("2021-10-15 13:40:11"), Login, locations["Alte Mälzerei"], persons["Hans Müller"]},
		JournalEntry{ParseTimestamp("2021-10-15 15:42:23"), Logout, locations["DHBW Mosbach"], persons["Lieschen Müller"]},
		JournalEntry{ParseTimestamp("2021-10-15 16:48:21"), Logout, locations["DHBW Mosbach"], persons["Max Mustermann"]},
		JournalEntry{ParseTimestamp("2021-10-15 16:52:00"), Logout, locations["DHBW Mosbach"], persons["Gisela Musterfrau"]},
		JournalEntry{ParseTimestamp("2021-10-15 17:15:22"), Logout, locations["DHBW Mosbach"], persons["Anne Meier"]},
		JournalEntry{ParseTimestamp("2021-10-15 17:32:45"), Login, locations["Alte Mälzerei"], persons["Otto Normalverbraucher"]},
		JournalEntry{ParseTimestamp("2021-10-15 19:15:12"), Logout, locations["Alte Mälzerei"], persons["Otto Normalverbraucher"]},
	}}

	actual, err := ReadJournal("analyzer_testdata", NewSimpleDate(2021, 10, 15))

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestReadNotExistingJournal(t *testing.T) {
	actual, err := ReadJournal("analyzer_testdata", NewSimpleDate(2020, 9, 4))

	assert.Error(t, err)
	assert.Nil(t, actual)
}

// 8.2) Get all locations for a specific person
func TestGetVisitedLocationsForPerson(t *testing.T) {
	expected := []Location{
		locations["DHBW Mosbach"],
	}

	journal := ReadJournal("analyzer_testdata", NewSimpleDate(2021, 10, 14))
	actual := journal.GetVisitedLocationsForPerson(persons["Max Mustermann"])

	assert.NotNil(actual)
	assert.Equal(t, expected, actual)
}

func TestGetVisitedLocationForPersonMultipleLocations(t *testing.T) {
	expected := []Location{
		locations["DHBW Mosbach"],
		locations["Alte Mälzerei"],
	}

	journal := ReadJournal("analyzer_testdata", NewSimpleDate(2021, 10, 14))
	actual := journal.GetVisitedLocationsForPerson(persons["Hans Müller"])

	assert.NotNil(actual)
	assert.Equal(t, expected, actual)
}

func TestGetVisitedLocationsForPersonNotExistingPerson(t *testing.T) {
	expected := []Location{}

	person := NewPerson("Susi", "Sorglos", Address{"Musterstraße", "20", "72327", "Musterstadt"})
	journal := ReadJournal("analyzer_testdata", GetSimpleDate(2021, 10, 14))
	actual := journal.GetVisitedLocationsForPerson(person)

	assert.Equal(t, expected, actual)
}

// 8.3) Get attendance list for a specific location as CSV
func TestGetAttendanceEntriesForLocation(t *testing.T) {
	expected := AttendanceList{
		AttendanceEntry{persons["Hans Müller"], ParseTimestamp("2021-10-15 13:40:11"), nil},
		AttendanceEntry{persons["Otto Normalverbraucher"], ParseTimestamp("2021-10-15 17:32:45"), ParseTimestamp("2021-10-15 19:15:12")},
	}

	journal := ReadJournal("analyzer_testdata", NewSimpleDate(2021, 10, 15))
	actual := journal.GetAttendanceListForLocation(locations["Alte Mälzerei"])

	assert.Equal(t, expected, actual)
}

func TestGetAttendanceListForLocationNotExistingLocation(t *testing.T) {
	expected := AttendanceList{}

	journal := ReadJournal("analyzer_testdata", NewSimpleDate(2021, 10, 15))

	actual := journal.GetAttendanceListAsCSVForLocation(NewLocation("Night Club"))

	assert.Equal(t, expected, actual)
}

func TestExtractToCSV(t *testing.T) {
	expected := `FirstName,LastName,Street,Number,ZipCode,City,Login,Logout
Hans,Müller,Feldweg,12,74722,Buchen,13:40:11,,
Otto,Normalverbraucher,Dieselstraße,52,70376,Stuttgart,Login,17:32:45,19:15:12`

	// Prepare list
	list := AttendanceList{
		AttendanceEntry{persons["Hans Müller"], ParseTimestamp("2021-10-15 13:40:11"), nil},
		AttendanceEntry{persons["Otto Normalverbraucher"], ParseTimestamp("2021-10-15 17:32:45"), ParseTimestamp("2021-10-15 19:15:12")},
	}

	err := ExtractToCSV("analyzer_testdata/2020-10-15_Alte-Mälzerei_Test.csv", list)
	assert.NoError(t, err)

	// Read file
	actual, err := os.ReadFile("analyzer_testdata/2020-10-15_Alte-Mälzerei_Test.csv")

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestExtractToCSVFailedToWriteFile(t *testing.T) {
	// Prepare list
	list := AttendanceList{
		AttendanceEntry{persons["Hans Müller"], ParseTimestamp("2021-10-15 13:40:11"), nil},
		AttendanceEntry{persons["Otto Normalverbraucher"], ParseTimestamp("2021-10-15 17:32:45"), ParseTimestamp("2021-10-15 19:15:12")},
	}

	err := ExtractToCSV("analyzer_testdata/2020-10-15_Alte-Mälzerei_Test.csv", list)
	assert.Error(t, err)
}

func TestExtractEmptyAttendanceListToCSV(t *testing.T) {
	expected := `FirstName,LastName,Street,Number,ZipCode,City,Login,Logout`

	// Prepare list
	list := AttendanceList{}

	err := ExtractToCSV("analyzer_testdata/2020-10-15_Empty_Test.csv", list)
	assert.NoError(t, err)

	// Read file
	actual, err := os.ReadFile("analyzer_testdata/2020-10-15_Empty_Test.csv")

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
