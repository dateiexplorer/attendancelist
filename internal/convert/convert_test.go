// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

// Package convert provides functionality to convert types between multiple formats.
// A type must implement the Converter interface to prepare it for using this it
// with this package.
package convert

import (
	"bytes"
	"errors"
	"testing"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("Test")

// Create a buffer which cannot read any input
type errorWriter struct{}

func (e errorWriter) Write(b []byte) (int, error) {
	return 0, errTest
}

func TestToCSV(t *testing.T) {
	expected := `FirstName,LastName,Street,Number,ZipCode,City,Login,Logout
Hans,Müller,Feldweg,12,74722,Buchen,13:40:11,
Otto,Normalverbraucher,Dieselstraße,52,70376,Stuttgart,17:32:45,19:15:12
Max,Mustermann,Musterstraße,20,74722,Buchen,,23:59:59
`
	// Create attendance list
	list := journal.AttendanceList{
		journal.NewAttendanceEntry(journal.NewPerson("Hans", "Müller", "Feldweg", "12", "74722", "Buchen"), timeutil.NewTimestamp(2021, 10, 15, 13, 40, 11), timeutil.InvalidTimestamp),
		journal.NewAttendanceEntry(journal.NewPerson("Otto", "Normalverbraucher", "Dieselstraße", "52", "70376", "Stuttgart"), timeutil.NewTimestamp(2021, 10, 15, 17, 32, 45), timeutil.NewTimestamp(2021, 10, 15, 19, 15, 12)),
		journal.NewAttendanceEntry(journal.NewPerson("Max", "Mustermann", "Musterstraße", "20", "74722", "Buchen"), timeutil.InvalidTimestamp, timeutil.NewTimestamp(2021, 10, 15, 23, 59, 59)),
	}

	// Read into buffer
	actual := new(bytes.Buffer)
	err := ToCSV(actual, list)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual.String())
}

func TestEmptyAttendanceListToCSV(t *testing.T) {
	expected := `FirstName,LastName,Street,Number,ZipCode,City,Login,Logout
`
	// Empty attendance list
	list := journal.AttendanceList{}

	// Read into buffer
	actual := new(bytes.Buffer)
	err := ToCSV(actual, list)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual.String())
}

func TestToCSVFailedToWrite(t *testing.T) {
	actual := errorWriter{}
	err := ToCSV(actual, journal.AttendanceList{})
	assert.ErrorIs(t, err, errTest)
}
