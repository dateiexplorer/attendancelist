// Package timeutil provides functionality for working with timestamps.
//
// All types are based on the time package from the standard library.
package timeutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Represents the format of an Timestamp.
const TimestampFormat = "2006-01-02 15:04:05"

var (
	// An invalid Timestamp
	InvalidTimestamp = NewTimestamp(1, 1, 1, 0, 0, 0)

	// An invalid Date
	InvalidDate = NewDate(1, 1, 1)
)

// A Timestamp represents an instant in time with second precision.
//
// It embedds a Time type from the time package.
type Timestamp struct {
	time.Time
}

// NewTimestamp returns a new Timestamp for timezone time.UTC.
func NewTimestamp(year int, month time.Month, day, hour, min, sec int) Timestamp {
	date := time.Date(year, month, day, hour, min, sec, 0, time.UTC)
	return Timestamp{date}
}

// ParseTimestamp parses a formatted string and returns the timestamp value it
// represents.
//
// The format of the input value must be in form "yyyy-MM-dd hh:mm:ss".
// If the given string cannot be parsed into a timestamp an error and an
// InvalidTimestamp will be returned.
func ParseTimestamp(value string) (Timestamp, error) {
	time, err := time.Parse(TimestampFormat, value)
	if err != nil {
		return Timestamp{time}, err
	}

	return Timestamp{time}, nil
}

// Clock returns the time part of a Timestamp as a string formatted as "hh:mm:ss".
// Each value is separeted with a colon.
//
// If you want the values instead of a string use the Clock function from the
// time package like t.Time.Clock().
func (t Timestamp) Clock() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
}

// Date returns the date part of a Timestamp as a Date structure.
func (t Timestamp) Date() Date {
	return NewDate(t.Time.Date())
}

// String returns the internal string representation of a Timestamp formatted as
// "yyyy-MM-dd hh:mm:ss"
func (t Timestamp) String() string {
	return t.Format(TimestampFormat)
}

// A Date represents a calendar date.
type Date struct {
	year  int
	month time.Month
	date  int
}

// NewDate returns a new Date.
func NewDate(year int, month time.Month, date int) Date {
	return Date{year, month, date}
}

// ParseDate parses a formatted string of the layout "dd.MM.yyyy" into a date.
//
// If the string doesn't apply the layout an error an a InvalidDate will be
// returned. The InvalidDate is the 1st of January 1 to be convenient to the
// time.Time struct from the standard library.
func ParseDate(value string) (Date, error) {
	values := strings.Split(value, ".")
	if len(values) != 3 {
		return Date{1, 1, 1}, errors.New("ParseDate: invalid length")
	}

	day, err := strconv.Atoi(values[0])
	if err != nil {
		return Date{1, 1, 1}, fmt.Errorf("error parsing date \"%v\": cannot convert \"%v\" to int: %w", value, values[0], err)
	}

	month, err := strconv.Atoi(values[1])
	if err != nil {
		return Date{1, 1, 1}, fmt.Errorf("error parsing date \"%v\": cannot convert \"%v\" to int: %w", value, values[1], err)
	}

	year, err := strconv.Atoi(values[2])
	if err != nil {
		return Date{1, 1, 1}, fmt.Errorf("error parsing date \"%v\": cannot convert \"%v\" to int: %w", value, values[2], err)
	}

	return Date{year, time.Month(month), day}, nil
}

// String returns the internal representation of a Date formatted as "yyyy-MM-dd".
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.year, int(d.month), d.date)
}
