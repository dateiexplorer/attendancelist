// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

// Package timeutil provides functionality for working with timestamps.
//
// All types are based on the time package from the standard library.
package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvalidTimestamp(t *testing.T) {
	invalid := NewTimestamp(1, 1, 1, 0, 0, 0)
	assert.Equal(t, invalid, InvalidTimestamp, "constant InvalidTimestamp is wrong")
}

func TestParseTimestamp(t *testing.T) {
	expected := NewTimestamp(2021, 10, 15, 15, 30, 25)
	actual, err := ParseTimestamp("2021/10/15 15:30:25 UTC")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual, "wrong timestamp while parsing")
}

func TestParseTimestampWithInvalidData(t *testing.T) {
	// Invalid format
	ts, err := ParseTimestamp("2021.10.15 15.30.25 UTC")
	assert.Error(t, err)
	assert.Equal(t, InvalidTimestamp, ts)
}

func TestTimestampDate(t *testing.T) {
	expected := Date{2021, 10, 15}
	actual := NewTimestamp(2021, 10, 15, 15, 20, 10).Date()
	assert.Equal(t, expected, actual)
}

func TestTimestampString(t *testing.T) {
	expected := "2021/10/15 15:20:10 UTC"
	actual := NewTimestamp(2021, 10, 15, 15, 20, 10).String()
	assert.Equal(t, expected, actual)
}

func TestInvalidDate(t *testing.T) {
	invalid := Date{1, 1, 1}
	assert.Equal(t, InvalidDate, invalid)
}

func TestNewDate(t *testing.T) {
	year, month, day := time.Date(2021, time.October, 15, 0, 0, 0, 0, time.UTC).Date()
	expected := Date{year, month, day}
	actual := NewDate(2021, time.October, 15)
	assert.Equal(t, expected, actual)
}

func TestParseDate(t *testing.T) {
	expected := map[string]Date{
		"2021/10/15": {2021, time.October, 15},
		"2021/02/01": {2021, time.February, 1},
	}

	for k, v := range expected {
		actual, err := ParseDate(k)
		assert.NoError(t, err)
		assert.Equal(t, v, actual)
	}
}

func TestParseDateFailed(t *testing.T) {
	// Define invalid input data
	inputs := []string{
		// Some input
		"some other input",
		// Invalid day
		"2001/02/xx",
		// Invalid month
		"2001/xx/01",
		// Invalid year
		"xxxx/10/01",
	}

	for _, input := range inputs {
		actual, err := ParseDate(input)
		assert.Error(t, err)
		assert.Equal(t, InvalidDate, actual)
	}
}

func TestClock(t *testing.T) {
	inputs := map[string]Timestamp{
		"15:30:25": NewTimestamp(2021, 10, 15, 15, 30, 25),
		"06:05:09": NewTimestamp(2021, 10, 15, 6, 5, 9),
	}

	for k, v := range inputs {
		actual := v.Clock()
		assert.Equal(t, k, actual)
	}
}

func TestDateString(t *testing.T) {
	date := Date{2021, 10, 15}
	assert.Equal(t, "2021-10-15", date.String())
}
