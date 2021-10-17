// Package journal provides functionality for writing text based journal files.
package journal

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/dateiexplorer/attendance-list/pkg/timeutil"
)

// A Journal represents a journal file with a date an severeal journal entries
type Journal struct {
	date    timeutil.Date
	entries []journalEntry
}

// ReadJournal reads data from a file for a specific date on the filesystem and
// returns a Journal.
// The filename must be formatted as "yyyy-MM-dd.log", otherwise the ReadJournal
// function cannot find the journal file.
// The dir variable specifies the directory on the filesystem where all journal files
// are stored, the date specifies the date of the journal file which should be
// read.
//
// A journal file is a text file which holds data separated by a comma sign.
// The layout must be as follows to parse the file correctly:
//
// timestamp,sessionIdentifier,event,locationName,firstName,lastName,street,number,zipCode,city
//
// where timestamp is a string formatted as "yyyy-MM-dd hh:mm:ss",
// sessionIdentifier is a temporary unique token as string,
// event is an numeric value which represents an Event,
// locationName is the name of the visited location,
// firstName is the firstName attribute of a Person
// lastName is the lastName attribute of a Person
// street is the street attribute of an Address
// number is the number attribute of an Address
// zipCode is the zipCode attribute of an Address
// and city is the city attribute of an Address.
//
// An error returned if the specific journal file cannot be open or cannot be
// parsed. If an error occured the functions returns also an empty Journal which
// contains the date and an empty slice of journal entries.
func ReadJournal(dir string, date timeutil.Date) (Journal, error) {
	entries := []journalEntry{}

	// Open file
	f, err := os.Open(path.Join(dir, date.String()+".log"))
	if err != nil {
		return Journal{date, []journalEntry{}}, fmt.Errorf("cannot open journal file: %w", err)
	}

	defer f.Close()

	// Scan every line.
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for i := 0; scanner.Scan(); i++ {
		values := strings.Split(scanner.Text(), ",")

		timestamp, err := timeutil.ParseTimestamp(values[0])
		if err != nil {
			return Journal{date, []journalEntry{}}, fmt.Errorf("cannot parse timestamp of journal file on line %v: %w", i, err)
		}

		id := values[1]

		action, err := strconv.Atoi(values[2])
		if err != nil {
			return Journal{date, []journalEntry{}}, fmt.Errorf("cannot parse action of journal file on line %v: %w", i, err)
		}

		location := Location{values[3]}
		person := Person{values[4], values[5], Address{values[6], values[7], values[8], values[9]}}

		entry := journalEntry{timestamp, sessionIdentifier(id), Event(action), location, person}
		entries = append(entries, entry)
	}

	return Journal{date, entries}, nil
}

// GetVisitedLocationsForPerson returns a slice of Locations which Person p has visited.
//
// If the Person doesn't exist the function will return an empty slice.
// The order in which the locations are returned is not deterministic and can change for
// each call.
// If a Person visited a Location multiple times the Location appears only once in the
// slice.
func (j Journal) GetVisitedLocationsForPerson(p Person) []Location {
	// Use a map to guarantee that a Location appears only once in the slice.
	m := map[Location]Location{}
	for _, e := range j.entries {
		if e.person == p {
			m[e.location] = e.location
		}
	}

	// Convert map into equal lengthed slice.
	locations := make([]Location, 0, len(m))
	for _, l := range m {
		locations = append(locations, l)
	}

	return locations
}

// GetAttendanceListForLocation returns an AttendanceList for a Location l
// which holds all AttendanceEntries extracted from the Journal j.
//
// If the Location wasn't found in the Journal an empty AttendanceList will
// be returned.
// The AttendanceList is sorted by the Login timestamp. It is not guaranteed that
// two AttendanceEntries with the same Login timestamp are always appears in the
// same order.
func (j Journal) GetAttendanceListForLocation(l Location) AttendanceList {
	m := map[sessionIdentifier]AttendanceEntry{}
	for _, e := range j.entries {
		if e.location == l {
			switch e.event {
			case Login:
				m[e.session] = NewAttendanceEntry(e.person, e.timestamp, timeutil.InvalidTimestamp)
			case Logout:
				if entry, ok := m[e.session]; ok {
					entry.logout = e.timestamp
					m[e.session] = entry
				}
			}
		}
	}

	// Convert the map into an equal length slice.
	list := make(AttendanceList, 0, len(m))
	for _, e := range m {
		list = append(list, e)
	}

	// Sort by Login timestamp.
	sort.Slice(list, func(i, j int) bool {
		return list[i].login.Before(list[j].login.Time)
	})

	return list
}

// A sessionIdentifier identifies which journal entries match together.
// This is important since the journal file documents the login and logout for a
// user in two separate journal entries.
type sessionIdentifier string

// A journalEntry represents one row in the Journal.
type journalEntry struct {
	timestamp timeutil.Timestamp
	session   sessionIdentifier
	event     Event
	location  Location
	person    Person
}

// An Event represents the reason why a new journal entry was written into the
// journal file.
type Event int

const (
	Login Event = iota
	Logout
)

// A Location represents a place where a Person can associated with.
type Location struct {
	name string
}

// A Person represents a citizen with a name and address.
type Person struct {
	firstName string
	lastName  string
	address   Address
}

// An Address represents a place where a people live.
type Address struct {
	street  string
	number  string
	zipCode string
	city    string
}

// An AttendanceList is a collection of AttendanceEntries.
type AttendanceList []AttendanceEntry

// NextEntry returns a read-only channel that loops through the hole AttendanceList
// and returns data of the AttendanceEntry as a string slice.
//
// Used to convert an AttendanceList to any file format.
func (a AttendanceList) NextEntry() <-chan []string {
	entries := make(chan []string)
	go func() {
		for _, e := range a {
			login := ""
			if e.login != timeutil.InvalidTimestamp {
				login = e.login.Clock()
			}

			logout := ""
			if e.logout != timeutil.InvalidTimestamp {
				logout = e.logout.Clock()
			}

			entries <- []string{e.person.firstName, e.person.lastName,
				e.person.address.street, e.person.address.number, e.person.address.zipCode, e.person.address.city,
				login, logout}
		}

		close(entries)
	}()

	return entries
}

// Header returns a string slice which describes the data given by the NextEntry
// function.
//
// Used to convert an AttendanceList to any file format.
func (a AttendanceList) Header() []string {
	return []string{"FirstName", "LastName", "Street", "Number", "ZipCode", "City", "Login", "Logout"}
}

// An AttendanceEntry represents a row of a AttendanceList.
// It associates a Person with a login and a logout timestamp.
type AttendanceEntry struct {
	person Person
	login  timeutil.Timestamp
	logout timeutil.Timestamp
}

// NewAttendanceEntry returns an AttendanceEntry.
func NewAttendanceEntry(person Person, login, logout timeutil.Timestamp) AttendanceEntry {
	return AttendanceEntry{person, login, logout}
}
