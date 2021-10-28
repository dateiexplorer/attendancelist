// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package token

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

type TokenAction int

const (
	Add TokenAction = iota
	Invalidate
)

type ValidTokens struct {
	sync.Map
}

func (m *ValidTokens) GetAccessTokenForLocation(loc journal.Location) (token *AccessToken, ok bool) {
	m.Range(func(key interface{}, value interface{}) bool {
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

func (m *ValidTokens) run(tokenQueue <-chan TokenQueueItem) {
	for item := range tokenQueue {
		switch item.action {
		case Add:
			m.Store(item.token.ID, item.token)
		case Invalidate:
			m.Delete(item.token.ID)
		}
	}
}

type TokenQueueItem struct {
	action TokenAction
	token  *AccessToken
}

// Locations are a collection of places that people can access.
type Locations struct {
	Locations []journal.Location `xml:"Location"`
}

func (l Locations) GenerateAccessTokens(idLength int, exp time.Duration, baseUrl string, port int) *ValidTokens {
	validTokens := new(ValidTokens)
	tokenQueue := make(chan TokenQueueItem, len(l.Locations)*LastValidTokens)

	go validTokens.run(tokenQueue)

	tokenIds := RandIDGenerator(idLength, len(l.Locations)*LastValidTokens)
	for _, loc := range l.Locations {
		generateAccessToken(tokenQueue, tokenIds, <-tokenIds, time.Now().UTC(), exp, LastValidTokens, loc, baseUrl, port)
	}

	return validTokens
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
	ID       string
	Exp      time.Time
	Iat      time.Time
	Valid    int
	Location journal.Location
	QR       []byte
}

func newAccessToken(id string, iat time.Time, exp time.Duration, valid int, loc journal.Location, baseUrl string, port int) AccessToken {
	token := AccessToken{ID: id, Exp: iat.Add(exp), Iat: iat, Valid: valid, Location: loc}

	qr, err := qrcode.Encode(fmt.Sprintf("%v:%v?token=%v", baseUrl, port, token.ID), qrcode.Medium, 256)
	if err != nil {
		panic(fmt.Errorf("cannot create qr code: %w", err))
	}

	token.QR = qr
	return token
}

func (t *AccessToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID       string           `json:"id"`
		Exp      int64            `json:"exp"`
		Iat      int64            `json:"iat"`
		Valid    int              `json:"valid"`
		Location journal.Location `json:"loc"`
		QR       []byte           `json:"qr"`
	}{
		ID:       t.ID,
		Exp:      t.Exp.Unix(),
		Iat:      t.Iat.Unix(),
		Valid:    t.Valid,
		Location: t.Location,
		QR:       t.QR,
	})
}

func generateAccessToken(tokenQueue chan<- TokenQueueItem, idGenerator <-chan string, id string, iat time.Time, exp time.Duration, valid int, loc journal.Location, baseUrl string, port int) *AccessToken {
	token := newAccessToken(id, iat, exp, valid, loc, baseUrl, port)

	go func() {
		// Add AccessToken to map
		tokenQueue <- TokenQueueItem{Add, &token}

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
