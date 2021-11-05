package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dateiexplorer/attendancelist/internal/convert"
	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/timeutil"
)

type Action int

const (
	GetVisitedLocationsForPerson Action = iota
	GetContactsForPerson
	GetAttendanceListForLocation
)

var person, location, filePath string

func main() {
	// Subcommands
	locationsCommand := flag.NewFlagSet("locations", flag.ExitOnError)
	locationsCommand.StringVar(&person, "person", "", "person for whom the locations are determined")

	contactsCommand := flag.NewFlagSet("contacts", flag.ExitOnError)
	contactsCommand.StringVar(&person, "person", "", "person for whom the locations are determined")

	attendancesCommand := flag.NewFlagSet("attendances", flag.ExitOnError)
	attendancesCommand.StringVar(&location, "location", "", "location for which an attendance list is created")
	attendancesCommand.StringVar(&filePath, "w", "", "filename")

	// Command must contain:
	// analyzer [command] <date>
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(1)
	}

	lastArg := os.Args[len(os.Args)-1]
	date, err := timeutil.ParseDate(lastArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(1)
	}

	args := os.Args[2 : len(os.Args)-1]

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

		doAction(GetVisitedLocationsForPerson, "testdata", date)
		return
	}

	if contactsCommand.Parsed() {
		if len(person) == 0 {
			contactsCommand.Usage()
			os.Exit(1)
		}

		doAction(GetContactsForPerson, "testdata", date)
		return
	}

	if attendancesCommand.Parsed() {
		if len(location) == 0 {
			attendancesCommand.Usage()
			os.Exit(1)
		}

		doAction(GetAttendanceListForLocation, "testdata", date)
		return
	}
}

func usage() string {
	return "Usage..."
}

func doAction(action Action, journalDir string, journalDate timeutil.Date) error {
	j, err := journal.ReadJournal(journalDir, journalDate)
	if err != nil {
		return err
	}

	switch action {
	case GetVisitedLocationsForPerson:
		attr := strings.Split(person, ",")

		persons := make([]journal.Person, 0, 1)

	loop:
		for _, e := range j.Entries() {
			p := e.Person

			for _, per := range persons {
				if p == per {
					continue loop
				}
			}

			for _, a := range attr {
				if p.FirstName != a && p.LastName != a && p.Address.Street != a && p.Address.Number != a && p.Address.ZipCode != a && p.Address.City != a {
					continue loop
				}
			}

			persons = append(persons, p)
		}

		if len(persons) < 1 {
			fmt.Println("No person found matches this attributes")
			return nil
		}

		if len(persons) > 1 {
			fmt.Println("There are more than one persons matching this attributes")
			fmt.Println(persons)
			return nil
		}

		locs := j.GetVisitedLocationsForPerson(&persons[0])
		fmt.Println(locs)
	case GetAttendanceListForLocation:
		list := j.GetAttendanceListForLocation(journal.Location(location))

		if len(filePath) == 0 {
			fmt.Println(list)
			return nil
		}

		f, err := os.Create(filePath)
		if err != nil {
			return err
		}

		defer f.Close()

		convert.ToCSV(f, list)
	}
	return nil
}
