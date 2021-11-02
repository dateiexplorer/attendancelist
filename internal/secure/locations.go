// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package secure

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sync"

	"io/ioutil"
	"os"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/skip2/go-qrcode"
)

// Sets how often a token can be reused after it expired.
// A value of 0 means that a token is invalid after it expires once.
const LastValidTokens = 1

type ValidTokenResponse struct {
	Valid bool         `json:"valid"`
	Token *AccessToken `json:"token"`
}

// A TokenAction represents what can do with a token internally.
type TokenAction int

const (
	// Add a token to a map
	Add TokenAction = iota
	// Invalidate a token
	Invalidate
)

// ValidTokens is a map which holds all valid tokens.
// The sync.Map is embedded in this type to allow concurrent reads and writes.
type ValidTokens struct {
	sync.Map
}

func (m *ValidTokens) GetAll() []*AccessToken {
	tokens := make([]*AccessToken, 0)
	m.Range(func(key, value interface{}) bool {
		token := value.(*AccessToken)
		tokens = append(tokens, token)
		return true
	})

	return tokens
}

// GetAccessTokenForLocation searches in ValidTokens for the newest AccessToken for
// the specific location loc.
// It is guaranteed that if a token for the location exists this function returns
// always a token with the maximum valid value. However, this token can be already
// expired if the read happens before a refreshed token is stored in the map.
//
// This function returns a pointer to the newest AccessToken and an ok value which
// is true if a token was found.
func (m *ValidTokens) GetAccessTokenForLocation(loc journal.Location) (token *AccessToken, ok bool) {
	m.Range(func(key, value interface{}) bool {
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

// run listens to the tokenQueue channel and processes the action
// for a token defined in the TokenQueueItem.
func (m *ValidTokens) run(tokenQueue <-chan TokenQueueItem) {
	go func() {
		for item := range tokenQueue {
			switch item.action {
			case Add:
				m.Store(item.token.ID, item.token)
			case Invalidate:
				m.Delete(item.token.ID)
			}
		}
	}()
}

// A TokenQueueItem represents an item which can be consumed
// by the ValidTokens map.
// The action describes what to do with the Token hold by
// a AccessToken pointer.
type TokenQueueItem struct {
	action TokenAction
	token  *AccessToken
}

// Locations is a collection of places that people can access.
type Locations struct {
	Locations []journal.Location `xml:"Location"`
}

// GenerateAccessTokens generates AccessTokens for each location in the Locations
// struct. It returns a pointer to a ValidTokens map which holds all valid tokens
// If the function was executed once tokens are generated automatically with a
// specific idLength, expire duration, a baseUrl and a port.
//
// ValidTokens is updated in background.
func (l *Locations) GenerateAccessTokens(idLength int, exp time.Duration, baseUrl string, port int) *ValidTokens {
	validTokens := new(ValidTokens)
	tokenQueue := make(chan TokenQueueItem, len(l.Locations)*LastValidTokens)

	validTokens.run(tokenQueue)

	tokenIds := RandIDGenerator(idLength, len(l.Locations)*LastValidTokens)
	for _, loc := range l.Locations {
		generateAccessToken(tokenQueue, tokenIds, <-tokenIds, time.Now().UTC(), exp, LastValidTokens, loc, baseUrl, port)
	}

	return validTokens
}

// Contains returns true if a provided Location loc is in the Locations data
// structure, otherwise false.
func (l *Locations) Contains(loc journal.Location) bool {
	for _, location := range l.Locations {
		if loc == location {
			return true
		}
	}

	return false
}

// UnmarshalJSON takes the JSON representation of a Locations collection and parses
// it to the collection.
// If the data cannot be parsed into the data structure this function returns and error.
func (l *Locations) UnmarshalJSON(data []byte) error {
	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("error while decoding: %w", err)
	}

	for _, item := range v {
		location, ok := item.(string)
		if !ok {
			return fmt.Errorf("cannot convert data to string: %v", item)
		}

		l.Locations = append(l.Locations, journal.Location(location))
	}

	return nil
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

// An AccessToken represents a token for a location.
// The ID is a temporary unique identifier for this token,
// Exp is the timestamp where this token expires,
// Iat is the 'issued at' time where to token was created,
// Valid indicates wheter an AccessToken refreshed itself. If Valid equals 0 the
// token will not refresh and invalidates after the expire time.
// Location is the Location which is associated with this token
// and QR is a byte slice which defines a QR-Code for this Accesssecure.
type AccessToken struct {
	ID       string
	Exp      time.Time
	Iat      time.Time
	Valid    int
	Location journal.Location
	QR       []byte
}

// newAccessToken returns a new AccessToken with the given attributes.
func newAccessToken(id string, iat time.Time, exp time.Duration, valid int, loc journal.Location, baseUrl string, port int) AccessToken {
	token := AccessToken{ID: id, Exp: iat.Add(exp), Iat: iat, Valid: valid, Location: loc}

	qr, err := qrcode.Encode(fmt.Sprintf("%v:%v?token=%v", baseUrl, port, token.ID), qrcode.Medium, 256)
	if err != nil {
		panic(fmt.Errorf("cannot create qr code: %w", err))
	}

	token.QR = qr
	return token
}

type jsonAccessToken struct {
	ID       string           `json:"id"`
	Exp      int64            `json:"exp"`
	Iat      int64            `json:"iat"`
	Valid    int              `json:"valid"`
	Location journal.Location `json:"loc"`
	QR       []byte           `json:"qr"`
}

// MarshalJSON returns the JSON representation of an Accesssecure.
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

func (t *AccessToken) UnmarshalJSON(data []byte) error {
	var tmp jsonAccessToken

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	t.ID = tmp.ID
	t.Exp = time.Unix(tmp.Exp, 0)
	t.Iat = time.Unix(tmp.Iat, 0)
	t.Valid = tmp.Valid
	t.Location = tmp.Location
	t.QR = tmp.QR

	return nil
}

// generateAccessToken generates and manages the AccessToken for a specific Location loc.
// It pushes a store request in the tokenQueue and start a goroutine which handles the
// lifetime of a secure.
// This function is recursive and refreshes a invalid token automatically or generates
// a new token for the specific location if necessary.
//
// It is recommended to use this function instead of the newAccessToken function if new
// access tokens should generated automatically in an concurrent way.
func generateAccessToken(tokenQueue chan<- TokenQueueItem, idGenerator <-chan string, id string, iat time.Time, exp time.Duration, valid int, loc journal.Location, baseUrl string, port int) *AccessToken {
	token := newAccessToken(id, iat, exp, valid, loc, baseUrl, port)

	// Add AccessToken to map
	// Ensure that this happens before invalidation
	tokenQueue <- TokenQueueItem{Add, &token}

	go func() {
		// Wait for token expire
		timestamp := <-time.After(time.Until(token.Exp))

		// Generate new access token
		if token.Valid == LastValidTokens {
			generateAccessToken(tokenQueue, idGenerator, <-idGenerator, timestamp, exp, valid, loc, baseUrl, port)
		}

		// This AccessToken is invalid
		tokenQueue <- TokenQueueItem{Invalidate, &token}

		// Refresh this token
		if token.Valid > 0 {
			generateAccessToken(tokenQueue, idGenerator, id, timestamp, exp, valid-1, loc, baseUrl, port)
		}
	}()

	return &token
}
