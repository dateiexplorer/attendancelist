// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

// Package web provides all functionality which is necessary for the
// service communication, such as cookies or user session management.
package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"sync"

	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/skip2/go-qrcode"
)

// Sets how often a token can be reused after it invalidates.
// A value of 0 means that a token is invalid after it reaches the expire time once.
const LastValidTokens = 1

func RunTokenManager(locs *Locations, exp time.Duration, url *url.URL, idLength int) *ValidTokens {
	validTokens := new(ValidTokens)

	ids := RandIDGenerator(idLength, len(locs.Slice))
	go func() {
		// Initially create tokens.
		onExpire(validTokens, ids, locs, time.Now(), exp, url)
		for {
			timestamp := <-time.After(exp)
			onExpire(validTokens, ids, locs, timestamp, exp, url)
		}

	}()

	return validTokens
}

func onExpire(validTokens *ValidTokens, ids <-chan string, locs *Locations, timestamp time.Time, exp time.Duration, url *url.URL) {
	// Update existing AccessTokens
	validTokens.update(timestamp, exp)

	// Generate new AccessTokens for each location
	tokens := locs.GenerateTokens(ids, timestamp, exp, url)
	validTokens.add(tokens)
}

// ValidTokens is a map which holds and manages all valid tokens for all locations.
// This type uses a sync.Map under the hood to allow concurrent read and writes.
type ValidTokens struct {
	internal sync.Map
}

// add stores a slice of new AccessTokens in the map.
func (m *ValidTokens) add(tokens []*AccessToken) {
	for _, t := range tokens {
		m.internal.Store(t.ID, t)
	}
}

// update performs a refresh for each AccessToken stored in the map. If an
// AccessToken invalidates through the refresh process, this token will be
// deleted automatically from the map.
// This guarantees that the map only holds valid tokens.
func (m *ValidTokens) update(timestamp time.Time, exp time.Duration) {
	m.internal.Range(func(key, value interface{}) bool {
		id := key.(string)
		t := value.(*AccessToken)

		m.internal.Delete(id)
		if t.Valid > 0 {
			m.internal.Store(id, t.refresh(timestamp, exp))
		}

		return true
	})
}

// GetAll returns a slice of all AccessTokens stored in this map.
func (m *ValidTokens) GetAll() []*AccessToken {
	tokens := make([]*AccessToken, 0)
	m.internal.Range(func(key, value interface{}) bool {
		token := value.(*AccessToken)
		tokens = append(tokens, token)
		return true
	})

	return tokens
}

// GetCurrentForLocation searches in ValidTokens for the newest AccessToken for
// the specific location loc.
// It is guaranteed that if a token for the location exists this function returns
// always a token with the maximum valid value. However, this token can be already
// expired if the read happens before a refreshed token is stored in the map.
//
// This function returns a pointer to the newest AccessToken and an ok value which
// is true if a token was found.
func (m *ValidTokens) GetCurrentForLocation(loc journal.Location) (token *AccessToken, ok bool) {
	m.internal.Range(func(key, value interface{}) bool {
		val := value.(*AccessToken)
		if val.Location == loc && val.Valid == LastValidTokens {
			token = val
			ok = true
			return false
		}

		ok = false
		return true
	})

	return token, ok
}

// GetByID returns an AccessToken with the given ID from the map.
// If no token found the function returns a nil value and the ok value is false.
// Note that either the token is a valid pointer or the ok value is false.
func (m *ValidTokens) GetByID(id string) (*AccessToken, bool) {
	if value, ok := m.internal.Load(id); ok {
		token := value.(*AccessToken)
		return token, true
	}

	return nil, false
}

// An AccessToken represents a token for a location.
// The ID is a temporary unique identifier for this token,
// Exp is the timestamp where this token expires,
// Iat is the 'issued at' time where to token was created,
// Valid indicates wheter an AccessToken refreshed itself. If Valid equals 0 the
// token will not refresh and invalidates after the expire time.
// Location is the Location which is associated with this token
// and QR is a byte slice which defines a QR-Code for this AccessToken.
type AccessToken struct {
	ID       string
	Iat      time.Time
	Exp      time.Time
	Valid    int
	Location journal.Location
	QR       []byte
}

// newAccessToken returns a new AccessToken with the given attributes.
func newAccessToken(id string, iat time.Time, exp time.Duration, valid int, loc journal.Location, url *url.URL) AccessToken {
	token := AccessToken{ID: id, Iat: iat, Exp: iat.Add(exp), Valid: valid, Location: loc}

	qr, err := qrcode.Encode(fmt.Sprintf("%v?token=%v", url.String(), token.ID), qrcode.Medium, 256)
	if err != nil {
		panic(fmt.Errorf("cannot create qr code: %w", err))
	}

	token.QR = qr
	return token
}

// refresh Refreshes the AccessToken by decrease the Valid value and
// update the expire time with the given timestamp and exp duration.
func (t AccessToken) refresh(timestamp time.Time, exp time.Duration) *AccessToken {
	t.Valid--
	if t.Valid >= 0 {
		t.Exp = timestamp.Add(exp)
	}

	return &t
}

// A jsonAccessToken is used for marshalling and unmarshalling an AccessToken.
// It is necessary to convert an AccessToken to a JSON object and back to an
// AccessToken.
type jsonAccessToken struct {
	ID       string           `json:"id"`
	Iat      int64            `json:"iat"`
	Exp      int64            `json:"exp"`
	Valid    int              `json:"valid"`
	Location journal.Location `json:"loc"`
	QR       []byte           `json:"qr"`
}

// MarshalJSON returns the JSON representation of an AccessToken.
func (t *AccessToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAccessToken{
		ID:       t.ID,
		Exp:      t.Exp.Unix(),
		Iat:      t.Iat.Unix(),
		Valid:    t.Valid,
		Location: t.Location,
		QR:       t.QR,
	})
}

// UnmarshalJSON returns an AccessToken from a JSON object.
// If the data cannot parse into an AccessToken this function returns an error.
func (t *AccessToken) UnmarshalJSON(data []byte) error {
	var tmp jsonAccessToken

	json.Unmarshal(data, &tmp)

	t.ID = tmp.ID
	t.Exp = time.Unix(tmp.Exp, 0)
	t.Iat = time.Unix(tmp.Iat, 0)
	t.Valid = tmp.Valid
	t.Location = tmp.Location
	t.QR = tmp.QR

	return nil
}

// Locations is a collection of places that Persons can be.
type Locations struct {
	Slice []journal.Location `xml:"Location"`
}

// ReadLocationsFormXML reads Locations from a XML file to the Locations struct.
// The path parameter defines the path to the XML file in the filesystem.
//
// Returns an error if the given XML file cannot be parsed into the Locations struct.
func ReadLocationsFromXML(path string) (Locations, error) {
	var locations Locations
	file, err := os.Open(path)
	if err != nil {
		return locations, fmt.Errorf("cannot open xml file: %w", err)
	}

	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return locations, fmt.Errorf("cannot read xml file: %w", err)
	}

	err = xml.Unmarshal(bytes, &locations)
	if err != nil {
		return locations, fmt.Errorf("cannot parse xml file: %w", err)
	}

	return locations, nil
}

// GenerateTokens generates AccessTokens for each Location in the collection with
// the given parameters.
func (l *Locations) GenerateTokens(ids <-chan string, timestamp time.Time, exp time.Duration, url *url.URL) []*AccessToken {
	tokens := make([]*AccessToken, 0, len(l.Slice))
	for _, loc := range l.Slice {
		t := newAccessToken(<-ids, timestamp, exp, LastValidTokens, loc, url)
		tokens = append(tokens, &t)
	}

	return tokens
}

// Contains returns true if a provided Location loc is in the Locations data
// structure, otherwise false.
func (l *Locations) Contains(loc journal.Location) bool {
	for _, location := range l.Slice {
		if loc == location {
			return true
		}
	}

	return false
}
