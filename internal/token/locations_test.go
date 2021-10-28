// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package token

import (
	"strings"
	"testing"
	"time"

	"path"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/skip2/go-qrcode"
	"github.com/stretchr/testify/assert"
)

// Data

var locations = Locations{[]journal.Location{"DHBW Mosbach", "Alte Mältzerei"}}

// Functions

func TestReadLocationsFromXML(t *testing.T) {
	expected := locations
	actual, err := ReadLocationsFromXML(path.Join("testdata", "locations.xml"))

	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
}

func TestReadLocationFromXMLFailedRead(t *testing.T) {
	expected := Locations{}
	actual, err := ReadLocationsFromXML(path.Join("falsePath", " locations.xml"))

	assert.Error(t, err)
	assert.EqualValues(t, expected, actual)
}

func TestNewAccessToken(t *testing.T) {
	id := "aabbccddee"
	iat := time.Now()
	exp := iat.Add(time.Duration(10) * time.Second)
	loc := locations.Locations[0]
	qr, err := qrcode.Encode("localhost:8081?token=aabbccddee", qrcode.Medium, 256)
	assert.NoError(t, err)

	// Create an AccessToken which expires in 10 seconds
	expected := AccessToken{id, exp, iat, 2, loc, qr}
	actual := NewAccessToken(loc, id, iat, time.Duration(10)*time.Second, "localhost", 8081)

	assert.Equal(t, expected, actual)
}

func TestQRCodeCreationFailed(t *testing.T) {
	assert.Panics(t, func() {
		id := strings.Repeat("a", 2311) // not more than 2332 bytes
		NewAccessToken(locations.Locations[0], string(id), time.Now(), time.Duration(10)*time.Second, "localhost", 8081)
	})
}

func TestGenerateAccessTokens(t *testing.T) {
	loc := locations.Locations[0]

	// Setup channels
	tokenQueue := make(chan TokenQueueItem, 1)
	idGenerator := make(chan string, 5)
	idGenerator <- "aabbccddee"
	idGenerator <- "ffgghhiijj"

	expected := GenerateAccessTokens(loc, tokenQueue, idGenerator, time.Duration(1)*time.Second, "localhost", 8081)
	actual := <-tokenQueue

	// Token should be added to queue
	assert.Equal(t, Add, actual.action)
	assert.Equal(t, expected, actual.token)
}

func TestGenerateAccessTokenInvalidate(t *testing.T) {
	loc := locations.Locations[0]

	// Setup channels
	tokenQueue := make(chan TokenQueueItem, 1)
	idGenerator := make(chan string, 5)
	idGenerator <- "aabbccddee"
	idGenerator <- "ffgghhiijj"
	idGenerator <- "kkllmmnnoo"

	token := GenerateAccessTokens(loc, tokenQueue, idGenerator, time.Duration(1)*time.Second, "localhost", 8081)

	// After 4 Entries the token must be invalidate
	// 1.) Add the token
	item := <-tokenQueue
	assert.Equal(t, Add, item.action)
	assert.Equal(t, token, item.token)
	// 2.) Add a second token -> start another goroutine
	item = <-tokenQueue
	assert.Equal(t, Add, item.action)
	assert.NotEqual(t, token, item.token)
	// 3.) Can be either the Adding of a third token or the Invalidation of the first
	// 4.) same as 3.)
	for i := 0; i < 2; i++ {
		<-tokenQueue
	}

	// Token is Invalid
	assert.Equal(t, 0, token.Valid)
}

func TestRenewToken(t *testing.T) {
	id := "aabbccddee"
	iat := time.Now()
	loc := locations.Locations[0]

	// Setup channels
	tokenQueue := make(chan TokenQueueItem, 1)
	idGenerator := make(chan string, 1)
	idGenerator <- "ffgghhiijj"

	token := NewAccessToken(loc, id, iat, time.Duration(10)*time.Second, "localhost", 8081)

	timestamp := time.Now()
	assert.Equal(t, 2, token.Valid)

	// Renew token
	token.renew(timestamp, tokenQueue, idGenerator, time.Duration(10)*time.Second, "localhost", 8081)

	// Token should have new iat and exp
	assert.Equal(t, timestamp, token.Iat)
	assert.Equal(t, timestamp.Add(time.Duration(10)*time.Second), token.Exp)
	// Valid is decreased by 1
	assert.Equal(t, 1, token.Valid)
	// Token's Id hasn't changed
	assert.Equal(t, "aabbccddee", token.ID)
	// Token's location hasn't changed
	assert.Equal(t, loc, token.Location)

	// Get current action from queue
	// New token should be added to queue
	item := <-tokenQueue

	assert.Equal(t, Add, item.action)
	assert.Equal(t, 2, item.token.Valid)
	assert.Equal(t, "ffgghhiijj", item.token.ID)
	assert.Equal(t, time.Duration(10)*time.Second, item.token.Exp.Sub(item.token.Iat))
	assert.Equal(t, loc, item.token.Location)
}
