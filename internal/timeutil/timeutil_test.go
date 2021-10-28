// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package timeutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvalidTimestamp(t *testing.T) {
	expected := NewTimestamp(1, 1, 1, 0, 0, 0)

	actual := InvalidTimestamp

	assert.Equal(t, expected, actual)
}

func TestParseTimestamp(t *testing.T) {
	expected := NewTimestamp(2021, 10, 15, 15, 30, 25)

	actual, err := ParseTimestamp("2021-10-15 15:30:25")

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestParseTimestampFailed(t *testing.T) {
	expected := InvalidTimestamp

	actual, err := ParseTimestamp("2021.10.15 15:30:25")

	assert.Error(t, err)
	assert.Equal(t, expected, actual)
}

func TestTimestampDate(t *testing.T) {
	expected := Date{2021, 10, 15}

	actual := NewTimestamp(2021, 10, 15, 15, 20, 10).Date()

	assert.Equal(t, expected, actual)
}

func TestTimestampString(t *testing.T) {
	expected := "2021-10-15 15:20:10"

	actual := NewTimestamp(2021, 10, 15, 15, 20, 10).String()

	assert.Equal(t, expected, actual)
}

func TestNewDate(t *testing.T) {
	year, month, day := time.Date(2021, time.October, 15, 0, 0, 0, 0, time.UTC).Date()
	expected := Date{year, month, day}

	actual := NewDate(2021, time.October, 15)

	assert.Equal(t, expected, actual)
}

func TestParseDate(t *testing.T) {
	expected := map[string]Date{
		"15.10.2021": {2021, time.October, 15},
		"01.02.2021": {2021, time.February, 1},
	}

	for k, v := range expected {
		actual, err := ParseDate(k)

		assert.NoError(t, err)
		assert.Equal(t, v, actual)
	}
}

func TestParseDateFailed(t *testing.T) {
	expected := InvalidDate

	// Define invalid input data
	inputs := []string{
		"some input",
		"x.09.2001",
		"01.x.2001",
		"01.09.x",
	}

	for _, input := range inputs {
		actual, err := ParseDate(input)

		assert.Error(t, err)
		assert.Equal(t, actual, expected)
	}
}

func TestClock(t *testing.T) {
	expected := []string{"15:30:25", "06:05:09"}

	timestamps := []Timestamp{
		NewTimestamp(2021, 10, 15, 15, 30, 25),
		NewTimestamp(2021, 10, 15, 6, 5, 9),
	}

	for i, timestamp := range timestamps {
		actual := timestamp.Clock()
		assert.Equal(t, expected[i], actual)
	}
}

func TestDateString(t *testing.T) {
	expected := "2021-10-15"

	date := Date{2021, 10, 15}
	actual := fmt.Sprint(date)

	assert.Equal(t, expected, actual)
}
