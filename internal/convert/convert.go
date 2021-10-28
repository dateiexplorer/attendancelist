// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

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
