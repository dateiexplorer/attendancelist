package csv

import (
	"fmt"
	"os"
	"strings"

	"github.com/dateiexplorer/attendance-list/pkg/convert"
)

// Convert converts the data from a type which implements convert.Converter in a
// CSV file and stores it on the filesystem at path.
//
// An error returned if the file cannot be created or the data cannot be written.
func Convert(path string, c convert.Converter) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot convert to csv: %w", err)
	}

	defer f.Close()

	_, err = f.WriteString(strings.Join(c.Header(), ",") + "\n")
	if err != nil {
		return fmt.Errorf("cannot convert to csv: %w", err)
	}

	for entry := range c.NextEntry() {
		_, err = f.WriteString(strings.Join(entry, ",") + "\n")
		if err != nil {
			return fmt.Errorf("cannot convert to csv: %w", err)
		}
	}

	return nil
}
