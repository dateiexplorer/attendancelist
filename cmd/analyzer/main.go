// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dateiexplorer/attendancelist/internal/convert"
	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
)

func main() {
	var person, location, filePath string

	// Subcommands
	locationsCommand := flag.NewFlagSet("locations", flag.ExitOnError)
	locationsCommand.StringVar(&person, "person", "", "person for whom the locations are determined")

	contactsCommand := flag.NewFlagSet("contacts", flag.ExitOnError)
	contactsCommand.StringVar(&person, "person", "", "person for whom the locations are determined")
	contactsCommand.StringVar(&filePath, "w", "", "filename")

	attendancesCommand := flag.NewFlagSet("attendances", flag.ExitOnError)
	attendancesCommand.StringVar(&location, "location", "", "location for which an attendance list is created")
	attendancesCommand.StringVar(&filePath, "w", "", "filename")

	// Command must contain:
	// analyzer [command] <date>
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(1)
	}

	// Parse date
	lastArg := os.Args[len(os.Args)-1]
	date, err := timeutil.ParseDate(lastArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(1)
	}

	// Read journal file
	j, err := journal.ReadJournal("data", date)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read journal file for the specific date: %v\n", err)
	}

	args := os.Args[2 : len(os.Args)-1]

	// Decide which command should be executed.
	switch os.Args[1] {
	case locationsCommand.Name():
		locationsCommand.Parse(args)
	case contactsCommand.Name():
		contactsCommand.Parse(args)
	case attendancesCommand.Name():
		attendancesCommand.Parse(args)
	default:
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(1)
	}

	if locationsCommand.Parsed() {
		if len(person) == 0 {
			locationsCommand.Usage()
			os.Exit(1)
		}

		if msg, err := printVisitedLocationsForPerson(j, person); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			fmt.Print(msg)
		}

		return
	}

	if contactsCommand.Parsed() {
		if len(person) == 0 {
			contactsCommand.Usage()
			os.Exit(1)
		}

		if msg, err := printContactsForPerson(j, person, filePath); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			fmt.Print(msg)
		}

		return
	}

	if attendancesCommand.Parsed() {
		if len(location) == 0 {
			attendancesCommand.Usage()
			os.Exit(1)
		}

		if msg, err := createAttendanceListForLocation(j, location, filePath); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			fmt.Print(msg)
		}

		return
	}
}

func usage() string {
	return `Journal analyzer cli tool by 5703004 and 5736465.

Usage:
    analyzer [command] <date>

    <date> is of form YYYY/mm/dd and specifies for which date a
    journal file should be load.

Commands:
    locations    Print locations for a specific person.
    contacts     Print all contacts for a specific person.
    attendances  Create an attendance list for a specific location.

To get help for any command type -h after the command name.
`
}

func printVisitedLocationsForPerson(j journal.Journal, person string) (string, error) {
	persons := getMatchingPersonsFromJournal(j, person)

	if len(persons) < 1 {
		return "", fmt.Errorf("no person found matches this attributes")
	}

	if len(persons) > 1 {
		errMsg := "there are more than one person matching this attributes:\n"
		for _, p := range persons {
			errMsg += fmt.Sprintf("  %v\n", p.String())
		}
		errMsg += "add more search criterias"
		return "", fmt.Errorf(errMsg)
	}

	// Get Locations for this peson
	locs := j.GetVisitedLocationsForPerson(&persons[0])
	msg := ""
	for _, l := range locs {
		msg += fmt.Sprintln(l)
	}

	return msg, nil
}

func printContactsForPerson(j journal.Journal, person string, filePath string) (string, error) {
	persons := getMatchingPersonsFromJournal(j, person)

	if len(persons) < 1 {
		return "", fmt.Errorf("no person found matches this attributes")
	}

	if len(persons) > 1 {
		errMsg := "there are more than one person matching this attributes:\n"
		for _, p := range persons {
			errMsg += fmt.Sprintf("  %v\n", p.String())
		}
		errMsg += "add more search criterias"
		return "", fmt.Errorf(errMsg)
	}

	// Get Contacts for this peson
	contacts := j.GetContactsForPerson(&persons[0])
	return writeToCSV(contacts, filePath)
}

func createAttendanceListForLocation(j journal.Journal, location string, filePath string) (string, error) {
	list := j.GetAttendanceListForLocation(journal.Location(location))
	return writeToCSV(list, filePath)
}

func getMatchingPersonsFromJournal(j journal.Journal, person string) []journal.Person {
	attr := strings.Split(person, ",")
	persons := make([]journal.Person, 0)

loop:
	for _, e := range j.Entries {
		p := e.Person

		// If same person, skip this entry
		for _, per := range persons {
			if p == per {
				continue loop
			}
		}

		// Attributes can be in a random order.
		for _, a := range attr {
			if p.FirstName != a && p.LastName != a && p.Address.Street != a && p.Address.Number != a && p.Address.ZipCode != a && p.Address.City != a {
				continue loop
			}
		}

		persons = append(persons, p)
	}

	return persons
}

func writeToCSV(c convert.Converter, filePath string) (string, error) {
	var f io.Writer

	// If no file path set, write to console
	if len(filePath) == 0 {
		f = os.Stdout
	} else {
		file, err := os.Create(filePath)
		if err != nil {
			return "", fmt.Errorf("cannot create file: %w", err)
		}

		defer file.Close()
		f = file
	}

	if err := convert.ToCSV(f, c); err != nil {
		return "", fmt.Errorf("cannot convert to csv: %w", err)
	}

	return fmt.Sprintln("Output successfully written."), nil
}
