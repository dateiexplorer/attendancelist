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
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/skip2/go-qrcode"
	"github.com/stretchr/testify/assert"
)

var locations = Locations{[]journal.Location{"DHBW Mosbach", "Alte Mälzerei"}}

func TestValidTokensAdd(t *testing.T) {
	validTokens := new(ValidTokens)

	// Map must be empty
	counter := 0
	validTokens.internal.Range(func(key, value interface{}) bool {
		counter++
		return true
	})

	assert.Equal(t, 0, counter)

	iat := time.Now()
	exp := iat.Add(time.Duration(5) * time.Second)
	qr, err := qrcode.Encode("http://login", qrcode.Medium, 256)
	assert.NoError(t, err)

	tokens := []*AccessToken{
		{"a", iat, exp, 1, "DHBW Mosbach", qr},
		{"b", iat, exp, 1, "Alte Mälzerei", qr},
	}

	validTokens.add(tokens)

	// Tokens should be in map
	v, ok := validTokens.internal.Load("a")
	assert.True(t, ok)
	token := v.(*AccessToken)
	assert.Equal(t, tokens[0], token)

	v, ok = validTokens.internal.Load("b")
	assert.True(t, ok)
	token = v.(*AccessToken)
	assert.Equal(t, tokens[1], token)

	// Map should contain exact 2 entries now
	counter = 0
	validTokens.internal.Range(func(key, value interface{}) bool {
		counter++
		return true
	})

	assert.Equal(t, 2, counter)
}

func TestValidTokensAddEmptyList(t *testing.T) {
	validTokens := new(ValidTokens)

	// Map must be empty
	counter := 0
	validTokens.internal.Range(func(key, value interface{}) bool {
		counter++
		return true
	})

	assert.Equal(t, 0, counter)

	// Add empty map
	validTokens.add([]*AccessToken{})

	// Map must be empty
	counter = 0
	validTokens.internal.Range(func(key, value interface{}) bool {
		counter++
		return true
	})

	assert.Equal(t, 0, counter)
}

func TestValidTokensUpdate(t *testing.T) {
	validTokens := new(ValidTokens)

	iat := time.Now()
	exp := iat.Add(time.Duration(5) * time.Second)
	qr, err := qrcode.Encode("http://login", qrcode.Medium, 256)
	assert.NoError(t, err)

	tokens := []*AccessToken{
		{"a", iat, exp, 1, "DHBW Mosbach", qr},
		{"b", iat, exp, 0, "Alte Mälzerei", qr},
	}

	for _, token := range tokens {
		validTokens.internal.Store(token.ID, token)
	}

	ts := time.Now()
	expDuration := time.Duration(5) * time.Second
	updated := AccessToken{"a", iat, ts.Add(expDuration), 0, "DHBW Mosbach", qr}

	validTokens.update(ts, expDuration)

	val, ok := validTokens.internal.Load("a")
	assert.True(t, ok)
	token := val.(*AccessToken)
	assert.Equal(t, updated, *token)

	val, ok = validTokens.internal.Load("b")
	assert.False(t, ok)
	assert.Nil(t, val)

	// Map contains only one element
	counter := 0
	validTokens.internal.Range(func(key, value interface{}) bool {
		counter++
		return true
	})

	assert.Equal(t, 1, counter)
}

func TestOnExpireInit(t *testing.T) {
	validTokens := new(ValidTokens)
	ids := make(chan string, len(locations.Slice))
	ids <- "aabbccddee"
	ids <- "ffgghhiijj"

	ts := time.Now()
	exp := time.Duration(5) * time.Second
	url, err := url.Parse("https://login")
	assert.NoError(t, err)

	onExpire(validTokens, ids, &locations, ts, exp, url)

	counter := 0
	validTokens.internal.Range(func(key, value interface{}) bool {
		id := key.(string)
		token := value.(*AccessToken)

		assert.Equal(t, id, token.ID)
		assert.Equal(t, 1, token.Valid)
		assert.Equal(t, ts, token.Iat)
		assert.Equal(t, ts.Add(exp), token.Exp)

		qr, err := qrcode.Encode(fmt.Sprintf("https://login?token=%v", key), qrcode.Medium, 256)
		assert.NoError(t, err)
		assert.Equal(t, qr, token.QR)

		if id == "aabbccddee" {
			assert.Equal(t, locations.Slice[0], token.Location)
		}

		if id == "ffgghhiijj" {
			assert.Equal(t, locations.Slice[1], token.Location)
		}

		counter++
		return true
	})

	assert.Equal(t, len(locations.Slice), counter)
}

func TestValidTokensGetAll(t *testing.T) {
	validTokens := new(ValidTokens)

	// Empty map returns an empty slice
	tokens := validTokens.GetAll()
	assert.Equal(t, 0, len(tokens))

	iat := time.Now()
	exp := iat.Add(time.Duration(5) * time.Second)
	qr, err := qrcode.Encode("https://login", qrcode.Medium, 256)
	assert.NoError(t, err)

	input := []*AccessToken{
		{"a", iat, exp, 1, "DHBW Mosbach", qr},
		{"b", iat, exp, 1, "Alte Mälzerei", qr},
	}

	for _, v := range input {
		validTokens.internal.Store(v.ID, v)
	}

	// All items must be returned.
	tokens = validTokens.GetAll()
	counter := 0
	for _, expected := range input {
		for _, actual := range tokens {
			if expected.ID == actual.ID {
				assert.Equal(t, expected, actual)
				counter++
			}
		}
	}

	assert.Equal(t, len(input), counter)
}

func TestValidTokensGetCurrentForLocation(t *testing.T) {
	validTokens := new(ValidTokens)

	// Prepare map
	iat := time.Now()
	exp := iat.Add(time.Duration(5) * time.Second)
	qr, err := qrcode.Encode("https://login", qrcode.Medium, 256)
	assert.NoError(t, err)

	input := []*AccessToken{
		{"a", iat, exp, 0, "DHBW Mosbach", qr},
		{"b", iat, exp, 1, "DHBW Mosbach", qr},
		{"c", iat, exp, 1, "Alte Mälzerei", qr},
	}

	for _, v := range input {
		validTokens.internal.Store(v.ID, v)
	}

	// Location exist
	token, ok := validTokens.GetCurrentForLocation("DHBW Mosbach")
	assert.True(t, ok)
	assert.Equal(t, input[1], token)

	// Location doesn't exist
	token, ok = validTokens.GetCurrentForLocation("Night Club")
	assert.False(t, ok)
	assert.Nil(t, token)
}

func TestValidTokensGetByID(t *testing.T) {
	validTokens := new(ValidTokens)

	// Prepare map
	iat := time.Now()
	exp := iat.Add(time.Duration(5) * time.Second)
	qr, err := qrcode.Encode("https://login", qrcode.Medium, 256)
	assert.NoError(t, err)

	input := []*AccessToken{
		{"a", iat, exp, 0, "DHBW Mosbach", qr},
		{"b", iat, exp, 1, "Alte Mälzerei", qr},
	}

	for _, v := range input {
		validTokens.internal.Store(v.ID, v)
	}

	// Token exists
	token, ok := validTokens.GetByID("a")
	assert.True(t, ok)
	assert.Equal(t, input[0], token)

	// Token doesn't exist
	token, ok = validTokens.GetByID("c")
	assert.False(t, ok)
	assert.Nil(t, token)
}

func TestNewAccessToken(t *testing.T) {
	iat := time.Now()
	expDuration := time.Duration(5) * time.Second
	urlStr := "https://login"
	url, err := url.Parse(urlStr)
	assert.NoError(t, err)
	qr, err := qrcode.Encode(urlStr+"?token=a", qrcode.Medium, 256)
	assert.NoError(t, err)

	token := AccessToken{"a", iat, iat.Add(expDuration), 1, "DHBW Mosbach", qr}
	actual := newAccessToken("a", iat, expDuration, 1, "DHBW Mosbach", url)
	assert.Equal(t, token, actual)
}

func TestQRCodeCreationFailed(t *testing.T) {
	url, err := url.Parse("https://localhost:4443")
	assert.NoError(t, err)
	assert.Panics(t, func() {
		id := strings.Repeat("a", 2311) // not more than 2332 bytes
		newAccessToken(string(id), time.Now(), time.Duration(10)*time.Second, 1, "DHBW Mosbach", url)
	})
}

func TestAccessTokenRefresh(t *testing.T) {
	iat := time.Now()
	expDuration := time.Duration(5) * time.Second
	qr, err := qrcode.Encode("https://login", qrcode.Medium, 256)
	assert.NoError(t, err)

	token := AccessToken{"a", iat, iat.Add(expDuration), 1, "DHBW Mosbach", qr}

	ts := time.Now()
	refresh := token.refresh(ts, expDuration)

	// Old token should be the same as before
	assert.Equal(t, "a", token.ID)
	assert.Equal(t, iat, token.Iat)
	assert.Equal(t, iat.Add(expDuration), token.Exp)
	assert.Equal(t, journal.Location("DHBW Mosbach"), token.Location)
	assert.Equal(t, 1, token.Valid)
	assert.Equal(t, qr, token.QR)

	// New token has updated exp and valid attribute.
	assert.Equal(t, token.ID, refresh.ID)
	assert.Equal(t, token.Iat, refresh.Iat)
	assert.Equal(t, token.Location, refresh.Location)
	assert.Equal(t, token.QR, refresh.QR)
	assert.Equal(t, 0, refresh.Valid)
	assert.Equal(t, ts.Add(expDuration), refresh.Exp)
}

func TestAccessTokenRefreshInvalid(t *testing.T) {
	iat := time.Now()
	expDuration := time.Duration(5) * time.Second
	qr, err := qrcode.Encode("https://login", qrcode.Medium, 256)
	assert.NoError(t, err)

	token := AccessToken{"a", iat, iat.Add(expDuration), 0, "DHBW Mosbach", qr}

	ts := time.Now()
	refresh := token.refresh(ts, expDuration)

	// Don't update expire time for an invalid token.
	assert.Equal(t, token.ID, refresh.ID)
	assert.Equal(t, token.Iat, refresh.Iat)
	assert.Equal(t, token.Exp, refresh.Exp)
	assert.Equal(t, token.Location, refresh.Location)
	assert.Equal(t, -1, refresh.Valid)
	assert.Equal(t, qr, refresh.QR)
}

func TestAccessTokenMarshalJSON(t *testing.T) {
	id := "a"
	iat := time.Now()
	exp := iat.Add(time.Duration(5) * time.Second)
	valid := 1
	loc := journal.Location("DHBW Mosbach")
	qr, err := qrcode.Encode("https://login", qrcode.Medium, 256)
	assert.NoError(t, err)

	iatUnix := iat.Unix()
	expUnix := exp.Unix()
	qrEncoded := base64.StdEncoding.EncodeToString(qr)

	marshal := fmt.Sprintf("{\"id\":\"%v\",\"iat\":%v,\"exp\":%v,\"valid\":%v,\"loc\":\"%v\",\"qr\":\"%v\"}", id, iatUnix, expUnix, valid, loc, qrEncoded)

	token := AccessToken{id, iat, exp, valid, loc, qr}
	actual, err := token.MarshalJSON()
	assert.NoError(t, err)

	assert.Equal(t, marshal, string(actual))
}

func TestAccesstokenUnmarshalJSON(t *testing.T) {
	id := "a"
	// Get iat and exp with precicion of seconds
	iat := time.Unix(time.Now().Unix(), 0)
	exp := iat.Add(time.Duration(5) * time.Second)
	valid := 1
	loc := journal.Location("DHBW Mosbach")
	qr, err := qrcode.Encode("https://login", qrcode.Medium, 256)
	assert.NoError(t, err)

	qrEncoded := base64.StdEncoding.EncodeToString(qr)

	// Expected token
	token := AccessToken{id, iat, exp, valid, loc, qr}

	// String to unmarshal
	marshal := fmt.Sprintf("{\"id\":\"%v\",\"iat\":%v,\"exp\":%v,\"valid\":%v,\"loc\":\"%v\",\"qr\":\"%v\"}", id, iat.Unix(), exp.Unix(), valid, loc, qrEncoded)

	var unmarshal AccessToken
	err = unmarshal.UnmarshalJSON([]byte(marshal))
	assert.NoError(t, err)

	assert.Equal(t, token, unmarshal)
}

func TestReadLocationsFromXML(t *testing.T) {
	actual, err := ReadLocationsFromXML(path.Join("testdata", "locations.xml"))

	assert.NoError(t, err)
	assert.EqualValues(t, locations, actual)
}

func TestReadLocationFromXMLFailedRead(t *testing.T) {
	expected := Locations{}
	actual, err := ReadLocationsFromXML(path.Join("falsePath", "locations.xml"))

	assert.Error(t, err)
	assert.EqualValues(t, expected, actual)
}

func TestGenerateTokens(t *testing.T) {
	ids := make(chan string, 2)
	ids <- "a"
	ids <- "b"

	ts := time.Now()
	exp := time.Duration(5) * time.Second
	url, err := url.Parse("https://login")
	assert.NoError(t, err)

	qrA, err := qrcode.Encode(url.String()+"?token=a", qrcode.Medium, 256)
	assert.NoError(t, err)
	qrB, err := qrcode.Encode(url.String()+"?token=b", qrcode.Medium, 256)
	assert.NoError(t, err)

	expected := []*AccessToken{
		{"a", ts, ts.Add(exp), LastValidTokens, locations.Slice[0], qrA},
		{"b", ts, ts.Add(exp), LastValidTokens, locations.Slice[1], qrB},
	}

	tokens := locations.GenerateTokens(ids, ts, exp, url)
	assert.Equal(t, expected, tokens)
}

func TestContains(t *testing.T) {
	actual := locations.Contains("DHBW Mosbach")
	assert.True(t, actual)
}

func TestContainsNotContains(t *testing.T) {
	actual := locations.Contains("Night Club")
	assert.False(t, actual)
}
