package convert

import (
	"encoding/csv"
	"io"
)

// A Converter provides the functions to convert data in any file format.
type Converter interface {
	Header() []string
	NextEntry() <-chan []string
}

// ToCSV converts the data from a type which implements convert.Converter in a
// CSV file format.
//
// An error returned if the data cannot be written.
func ToCSV(w io.Writer, c Converter) error {
	writer := csv.NewWriter(w)

	writer.Write(c.Header())

	for entry := range c.NextEntry() {
		writer.Write(entry)
	}

	writer.Flush()
	return writer.Error()
}
