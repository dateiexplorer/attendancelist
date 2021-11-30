// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

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
	"time"

	"github.com/dateiexplorer/attendancelist/internal/timeutil"
)

const journalFileExtension = ".journal"

// A Journal represents a journal file with a date an severeal JournalEntries
type Journal struct {
	Date    timeutil.Date
	Entries []JournalEntry
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
// contains the date and an empty slice of JournalEntries.
func ReadJournal(dir string, date timeutil.Date) (Journal, error) {
	entries := []JournalEntry{}
	// Open file
	f, err := os.Open(path.Join(dir, date.String()+journalFileExtension))
	if err != nil {
		return Journal{date, []JournalEntry{}}, fmt.Errorf("cannot open journal file: %w", err)
	}
	defer f.Close()

	// Scan every line.
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for i := 0; scanner.Scan(); i++ {
		values := strings.Split(scanner.Text(), ",")
		timestamp, err := timeutil.ParseTimestamp(values[0])
		if err != nil {
			return Journal{date, []JournalEntry{}}, fmt.Errorf("cannot parse timestamp of journal file on line %v: %w", i, err)
		}

		id := values[1]
		action, err := strconv.Atoi(values[2])
		if err != nil {
			return Journal{date, []JournalEntry{}}, fmt.Errorf("cannot parse action of journal file on line %v: %w", i, err)
		}

		location := Location(values[3])
		person := Person{values[4], values[5], Address{values[6], values[7], values[8], values[9]}}
		entry := JournalEntry{timestamp, id, Event(action), location, person}
		entries = append(entries, entry)
	}

	return Journal{date, entries}, nil
}

// WriteToJournalFile appends a JournalEntry e to the corresponding journal file in
// the dir directory.
// The JournalEntry will be written in the file named "yyyy-MM-dd.log". The
// date will be extracted from the journalEntries timestamp itself.
//
// The functions returns an error if the writing operations causes an error.
func WriteToJournalFile(dir string, e *JournalEntry) error {
	// Get wright journal file for this entry.
	// Every day has it's own journal file.
	date := e.Timestamp.Date()
	f, err := os.OpenFile(path.Join(dir, date.String()+journalFileExtension), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("cannot write to journal file: %w", err)
	}
	defer f.Close()

	// Write to journal file
	s := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\n", e.Timestamp, e.SessionID, e.Event, e.Location, e.Person.FirstName, e.Person.LastName,
		e.Person.Address.Street, e.Person.Address.Number, e.Person.Address.ZipCode, e.Person.Address.City)
	_, err = f.WriteString(s)
	if err != nil {
		return fmt.Errorf("cannot write to journal file: %w", err)
	}

	return nil
}

// GetVisitedLocationsForPerson returns a slice of Locations which Person p has visited.
//
// If the Person doesn't exist the function will return an empty slice.
// The order in which the locations are returned is not deterministic and can change for
// each call.
// If a Person visited a Location multiple times the Location appears only once in the
// slice.
func (j Journal) GetVisitedLocationsForPerson(p *Person) []Location {
	// Use a map to guarantee that a Location appears only once in the slice.
	m := map[Location]Location{}
	for _, e := range j.Entries {
		if e.Person == *p {
			m[e.Location] = e.Location
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
	m := map[string]AttendanceEntry{}
	for _, e := range j.Entries {
		if e.Location == l {
			switch e.Event {
			case Login:
				m[e.SessionID] = NewAttendanceEntry(e.Person, e.Timestamp, timeutil.InvalidTimestamp)
			case Logout:
				if entry, ok := m[e.SessionID]; ok {
					entry.logout = e.Timestamp
					m[e.SessionID] = entry
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

// GetContactsForPerson returns a ContactList which holds all contacts for the
// Person p, which can be extracted from the Journal j.
func (j Journal) GetContactsForPerson(p *Person) ContactList {
	var contacts ContactList

	global := make(map[string]JournalEntry)
	var local map[string]JournalEntry

	startTimestamp := timeutil.InvalidTimestamp
	var loc Location

	// Defines what happens on logout of the searched person
	onLogout := func(end timeutil.Timestamp) {
		// Happends if searched person not logged in on this day, but logged out
		if startTimestamp == timeutil.InvalidTimestamp {
			// Choose beginning of this day
			startTimestamp = timeutil.NewTimestamp(j.Date.Year, j.Date.Month, j.Date.Day, 0, 0, 0)
		}

		// Get all entries from local map
		for key, value := range local {
			contacts = append(contacts, NewContact(value.Person, value.Location, value.Timestamp, end))
			delete(local, key)
		}

		// Get corresponding contacts from global map
		for _, value := range global {
			if value.Location == loc {
				contacts = append(contacts, NewContact(value.Person, value.Location, startTimestamp, end))
			}
		}

		startTimestamp = timeutil.InvalidTimestamp
	}

	for _, entry := range j.Entries {
		// Person is not searched
		if entry.Person != *p {
			switch entry.Event {
			case Login:
				if startTimestamp != timeutil.InvalidTimestamp && entry.Location == loc {
					// Store in local map
					local[entry.SessionID] = entry
				} else {
					global[entry.SessionID] = entry
				}
			case Logout:
				if startTimestamp != timeutil.InvalidTimestamp && entry.Location == loc {
					if value, ok := local[entry.SessionID]; ok {
						// Contact after persons login
						contacts = append(contacts, NewContact(entry.Person, entry.Location, value.Timestamp, entry.Timestamp))
						delete(local, entry.SessionID)
					} else {
						// Contact before persons login
						contacts = append(contacts, NewContact(entry.Person, entry.Location, startTimestamp, entry.Timestamp))
					}
				}

				delete(global, entry.SessionID)
			}
		} else {
			loc = entry.Location
			switch entry.Event {
			case Login:
				startTimestamp = entry.Timestamp
				// Begin new local map for entries
				local = make(map[string]JournalEntry)
			case Logout:
				onLogout(entry.Timestamp)
			}
		}
	}

	// There are contacts left
	if startTimestamp != timeutil.InvalidTimestamp {
		onLogout(timeutil.NewTimestamp(j.Date.Year, j.Date.Month, j.Date.Day, 23, 59, 59))
	}

	return contacts
}

// A Contact represents the meet with a person. It additionally stores the Location of the meet,
// the Start and End time and the Duration.
type Contact struct {
	Person     Person
	Location   Location
	Start, End timeutil.Timestamp
	Duration   time.Duration
}

// NewContact returns a new Contact with the given attributes.
// The Duration will be calculated as the difference between the Start and End Timestamp.
func NewContact(p Person, loc Location, start, end timeutil.Timestamp) Contact {
	return Contact{Person: p, Location: loc, Start: start, End: end, Duration: end.Sub(start.Time)}
}

// A ContactList is a slice of Contacts.
type ContactList []Contact

// NextEntry returns a read-only channel that loops through the hole ContactList
// and returns data of the Contact as a string slice.
//
// Used to convert an ContactList to any file format.
func (l ContactList) NextEntry() <-chan []string {
	entries := make(chan []string)
	go func() {
		for _, e := range l {
			entries <- []string{e.Person.FirstName, e.Person.LastName,
				e.Person.Address.Street, e.Person.Address.Number, e.Person.Address.ZipCode, e.Person.Address.City,
				string(e.Location), e.Start.String(), e.End.String(), e.Duration.String()}
		}

		close(entries)
	}()

	return entries
}

// Header returns a string slice which describes the data given by the NextEntry
// function.
//
// Used to convert an ContactList to any file format.
func (l ContactList) Header() []string {
	return []string{"FirstName", "LastName", "Street", "Number", "ZipCode", "City", "Location", "Start", "End", "Duration"}
}

// A JournalEntry represents one row in the Journal.
type JournalEntry struct {
	Timestamp timeutil.Timestamp `json:"timestamp"`
	SessionID string             `json:"sessionId"`
	Event     Event              `json:"event"`
	Location  Location           `json:"location"`
	Person    Person             `json:"person"`
}

// NewJournalEntry returns a new JournalEntry with the given parameters.
func NewJournalEntry(timestamp timeutil.Timestamp, sessionID string, event Event, location Location, person Person) JournalEntry {
	return JournalEntry{timestamp, sessionID, event, location, person}
}

// An Event represents the reason why a new JournalEntry was written into the
// journal file.
type Event int

const (
	Login Event = iota
	Logout
)

// A Location represents a place where a Person can associated with.
type Location string

// A Person represents a citizen with a name and address.
type Person struct {
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Address   Address `json:"address"`
}

func (p *Person) String() string {
	return fmt.Sprintf("%v,%v,%v,%v,%v,%v", p.FirstName, p.LastName, p.Address.Street, p.Address.Number, p.Address.ZipCode, p.Address.City)
}

// NewPerson returns a new Person with the given attributes.
func NewPerson(firstName, lastName, street, number, zipCode, city string) Person {
	return Person{firstName, lastName, Address{street, number, zipCode, city}}
}

// An Address represents a place where a people live.
type Address struct {
	Street  string `json:"street"`
	Number  string `json:"number"`
	ZipCode string `json:"zipCode"`
	City    string `json:"city"`
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
			entries <- []string{e.person.FirstName, e.person.LastName,
				e.person.Address.Street, e.person.Address.Number, e.person.Address.ZipCode, e.person.Address.City,
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
