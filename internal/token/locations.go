package token

import (
	"encoding/xml"
	"fmt"

	"io/ioutil"
	"os"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/skip2/go-qrcode"
)

// Represents a second as nanosecond
const SecondInNanoseconds = 1_000_000_000

// Sets how many tokens for a location can be used to log in.
const LastValidTokens = 2

type TokenAction int

const (
	Add TokenAction = iota
	Invalidate
)

type ValidTokens map[string]*AccessToken

type TokenQueueItem struct {
	action TokenAction
	token  *AccessToken
}

// Locations are a collection of places that people can access.
type Locations struct {
	Locations []journal.Location `xml:"Location"`
}

func (l Locations) AccessTokenMap(idLength int, exp int64, baseUrl string, port int) (*ValidTokens, chan TokenQueueItem) {
	validTokens := make(ValidTokens)
	tokenQueue := make(chan TokenQueueItem, len(l.Locations)*LastValidTokens)

	go validTokens.run(tokenQueue)

	tokenIds := RandIDGenerator(idLength, len(l.Locations)*LastValidTokens)
	for _, loc := range l.Locations {
		GenerateAccessTokens(&loc, tokenQueue, tokenIds, exp, baseUrl, port)
	}

	return &validTokens, tokenQueue
}

func (m *ValidTokens) run(tokenQueue <-chan TokenQueueItem) {
	for item := range tokenQueue {
		switch item.action {
		case Add:
			(*m)[item.token.id] = item.token
		case Invalidate:
			delete(*m, item.token.id)
		}
	}
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
type AccessToken struct {
	id       string
	exp      time.Time
	iat      time.Time
	valid    int
	location *journal.Location
	qr       []byte
}

func NewAccessToken(loc *journal.Location, id string, iat time.Time, exp int64, baseUrl string, port int) AccessToken {
	token := AccessToken{id: id, location: loc, exp: iat.Add(time.Duration(exp * SecondInNanoseconds)), iat: iat, valid: LastValidTokens}

	qr, err := qrcode.Encode(fmt.Sprintf("%v:%v?token=%v", baseUrl, port, token.id), qrcode.Medium, 256)
	if err != nil {
		panic(fmt.Errorf("cannot create qr code: %w", err))
	}

	token.qr = qr
	return token
}

func GenerateAccessTokens(loc *journal.Location, tokenQueue chan<- TokenQueueItem, idGenerator <-chan string, exp int64, baseUrl string, port int) *AccessToken {
	token := NewAccessToken(loc, <-idGenerator, time.Now(), exp, baseUrl, port)

	go func() {
		// Create new AccessToken
		tokenQueue <- TokenQueueItem{Add, &token}

		for token.valid > 0 {
			// Wait for expire interval
			timestamp := <-time.After(time.Until(token.exp))

			token.renew(timestamp, tokenQueue, idGenerator, exp, baseUrl, port)
		}

		// This AccessToken is invalid
		tokenQueue <- TokenQueueItem{Invalidate, &token}
	}()

	return &token
}

func (t *AccessToken) renew(timestamp time.Time, tokenQueue chan<- TokenQueueItem, idGenerator <-chan string, exp int64, baseUrl string, port int) {
	// If token first expired, create a new token
	// This generates the token chain for a specific location
	if t.valid == LastValidTokens {
		GenerateAccessTokens(t.location, tokenQueue, idGenerator, exp, baseUrl, port)
	}

	// Update timestamps
	t.iat = timestamp
	t.exp = timestamp.Add(time.Duration(exp * SecondInNanoseconds))
	t.valid--
}
