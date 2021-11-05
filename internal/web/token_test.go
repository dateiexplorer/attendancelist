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
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
)

// Data

var locations = Locations{[]journal.Location{"DHBW Mosbach", "Alte Mälzerei"}}

// var tokens = []AccessToken{
// 	newAccessToken("aabbccddee", time.Date(2021, 10, 28, 20, 0, 0, 0, time.UTC), time.Duration(1)*time.Second, 0, "DHBW Mosbach", "localhost", 8081),
// 	newAccessToken("ffgghhiijj", time.Date(2021, 10, 28, 20, 0, 20, 0, time.UTC), time.Duration(1)*time.Second, 1, "DHBW Mosbach", "localhost", 8081),
// 	newAccessToken("kkllmmnnoo", time.Date(2021, 10, 28, 20, 0, 0, 0, time.UTC), time.Duration(1)*time.Second, 1, "Alte Mälzerei", "localhost", 8081),
// }

// Functions

func TestGenerateTokens(t *testing.T) {
	ids := make(chan string, 20)
	ids <- "a"
	ids <- "b"
	ids <- "c"
	ids <- "d"
	ids <- "e"
	ids <- "f"
	ids <- "g"
	ids <- "h"
	ids <- "i"
	ids <- "j"
	ids <- "k"
	ids <- "l"
	ids <- "m"
	ids <- "n"
	ids <- "o"
	ids <- "p"
	ids <- "q"

	url, _ := url.Parse("https://localhost/test")

	validTokens := locations.GenerateTokens(ids, time.Duration(2)*time.Second, *url)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(2) * time.Second)
		validTokens.Range(func(key, value interface{}) bool {
			t := value.(*AccessToken)
			fmt.Println(t.Location, t.Valid)
			return true
		})
		fmt.Println("------")
	}
}

// func TestReadLocationsFromXML(t *testing.T) {
// 	expected := locations
// 	actual, err := ReadLocationsFromXML(path.Join("testdata", "locations.xml"))

// 	assert.NoError(t, err)
// 	assert.EqualValues(t, expected, actual)
// }

// func TestReadLocationFromXMLFailedRead(t *testing.T) {
// 	expected := Locations{}
// 	actual, err := ReadLocationsFromXML(path.Join("falsePath", " locations.xml"))

// 	assert.Error(t, err)
// 	assert.EqualValues(t, expected, actual)
// }

// func TestUnmarshalJSON(t *testing.T) {
// 	expected := Locations{[]journal.Location{"DHBW Mosbach", "Alte Mälzerei"}}

// 	var actual Locations
// 	actual.UnmarshalJSON([]byte(`["DHBW Mosbach", "Alte Mälzerei"]`))

// 	assert.Equal(t, expected, actual)
// }

// func TestUnmarshalJSONFail(t *testing.T) {
// 	var actual Locations
// 	err := actual.UnmarshalJSON([]byte(`["DHBW Mosbach", "Alte Mälzerei", 0]`))

// 	assert.Error(t, err)
// }

// func TestNewAccessToken(t *testing.T) {
// 	id := "aabbccddee"
// 	iat := time.Now()
// 	exp := iat.Add(time.Duration(10) * time.Second)
// 	loc := locations.Locations[0]
// 	qr, err := qrcode.Encode("localhost:8081?token=aabbccddee", qrcode.Medium, 256)
// 	assert.NoError(t, err)

// 	// Create an AccessToken which expires in 10 seconds
// 	expected := AccessToken{id, exp, iat, 1, loc, qr}
// 	actual := newAccessToken(id, iat, time.Duration(10)*time.Second, 1, loc, "localhost", 8081)

// 	assert.Equal(t, expected, actual)
// }

// func TestQRCodeCreationFailed(t *testing.T) {
// 	assert.Panics(t, func() {
// 		id := strings.Repeat("a", 2311) // not more than 2332 bytes
// 		newAccessToken(string(id), time.Now(), time.Duration(10)*time.Second, 1, locations.Locations[0], "localhost", 8081)
// 	})
// }

// func TestMarhsalJSONAccessToken(t *testing.T) {
// 	id := "112f073f4c"
// 	iat := time.Now()
// 	exp := iat.Add(time.Duration(10) * time.Second)
// 	loc := locations.Locations[0]

// 	expected := fmt.Sprintf("{\"id\":\"112f073f4c\",\"exp\":%v,\"iat\":%v,\"valid\":1,\"loc\":\"DHBW Mosbach\",\""+
// 		"qr\":\"iVBORw0KGgoAAAANSUhEUgAAAQAAAAEAAQMAAABmvDolAAAABlBMVEX///8AAABVwtN+AAABnElEQVR42uyYsZGEMAxFxRBs"+
// 		"SAlbCqXZpbkUStiQwON/879g4e649GatQZk9jwBZ0pdkt91228faBNpqM4ZqE4q191UoYDGzkSdUntOwHleRgBltXCdgqI/XsxxXQYF"+
// 		"mZk+AfogM0BZjVIcEFMJTSTaK+SPsewe8IE0leVQzea+KWOfAzhUjtrBGXQtP3wBDWD+dUB9Y5gz5ofUF2BMZ1aaSIN1Mys08HK8ZAn"+
// 		"A/4DVnHeds9AZ+vmbvgJlOE7KZchN48WpcgwGWXD023WxqErwlCgT4iX4YpZtY3/kcCKBIqj0QgJIGPnc+h30IQEMJ/UBsF5R0bg8iA"+
// 		"CzFbWSNcsWBUthaMIAi481Qc6coqvO5TQoB8IWpObvibD3tOexDAFSchz+3Htq/qeEARrWnMFVWlr9Nap8PvGM3Dz5OsgTNFwNI38Cx"+
// 		"0drNlwMXi7uugX1qtu3XIT9cjNWdA74D0dQMzSa0FhRAVk9bElbT+s5CAsWUvBqr2dPWaICvrzZvlOQ1aogGnDZaOnAQw+8i1jlw222"+
// 		"3/bt9BQAA//+H9RJw44jhWwAAAABJRU5ErkJggg==\"}", exp.Unix(), iat.Unix())

// 	token := newAccessToken(id, iat, time.Duration(10)*time.Second, 1, loc, "localhost", 8081)
// 	actual, err := token.MarshalJSON()
// 	assert.NoError(t, err)
// 	assert.Equal(t, expected, string(actual))
// }

// func TestUnmarshalJSONAccessToken(t *testing.T) {
// 	// Get iat and exp but with precicion of seconds
// 	iat := time.Unix(time.Now().Unix(), 0)
// 	exp := iat.Add(time.Duration(10) * time.Second)
// 	qr, err := qrcode.Encode("localhost:8081?token=aabbccddee", qrcode.Medium, 256)
// 	assert.NoError(t, err)

// 	expected := AccessToken{"aabbccddee", exp, iat, 1, "DHBW Mosbach", qr}

// 	json := fmt.Sprintf(`{
// 		"id": "aabbccddee",
// 		"exp": %v,
// 		"iat": %v,
// 		"valid": 1,
// 		"loc": "DHBW Mosbach",
// 		"qr": "%v"
// 	}`, exp.Unix(), iat.Unix(), base64.StdEncoding.EncodeToString(qr))

// 	var actual AccessToken
// 	err = actual.UnmarshalJSON([]byte(json))

// 	assert.NoError(t, err)
// 	assert.Equal(t, expected, actual)
// }

// func TestGenerateAccessToken(t *testing.T) {
// 	id := "aabbccddee"
// 	iat := time.Now()
// 	loc := locations.Locations[0]

// 	// Setup channels
// 	tokenQueue := make(chan TokenQueueItem, 1)
// 	idGenerator := make(chan string)

// 	expected := generateAccessToken(tokenQueue, idGenerator, id, iat, time.Duration(1)*time.Second, 0, loc, "localhost", 8081)
// 	actual := <-tokenQueue

// 	// Token should be added to queue
// 	assert.Equal(t, Add, actual.action)
// 	assert.Equal(t, expected, actual.token)

// 	actual = <-tokenQueue

// 	// Token should not be refreshed -> Invalidate this token
// 	assert.Equal(t, Invalidate, actual.action)
// 	assert.Equal(t, actual.token, expected)
// }

// func TestGenerateAccessTokenRefresh(t *testing.T) {
// 	id := "aabbccddee"
// 	iat := time.Now()
// 	loc := locations.Locations[0]

// 	// Setup channels
// 	tokenQueue := make(chan TokenQueueItem, 1)
// 	idGenerator := make(chan string, 1)
// 	idGenerator <- "ffgghhiijj"

// 	expected := generateAccessToken(tokenQueue, idGenerator, id, iat, time.Duration(1)*time.Second, 1, loc, "localhost", 8081)
// 	actual := <-tokenQueue

// 	// Token should be added to queue
// 	assert.Equal(t, Add, actual.action)
// 	assert.Equal(t, expected, actual.token)

// 	actual = <-tokenQueue

// 	// New AccessToken for same location should be created
// 	assert.Equal(t, Add, actual.action)
// 	assert.Equal(t, "ffgghhiijj", actual.token.ID)
// 	assert.Equal(t, 1, actual.token.Valid)
// 	assert.Equal(t, loc, actual.token.Location)

// 	actual = <-tokenQueue

// 	// Token should be invalidated
// 	assert.Equal(t, Invalidate, actual.action)
// 	assert.Equal(t, actual.token, expected)

// 	actual = <-tokenQueue

// 	// Token should be refreshed -> Valid decreased by 1
// 	assert.Equal(t, Add, actual.action)
// 	assert.Equal(t, id, actual.token.ID)
// 	assert.Equal(t, 0, actual.token.Valid)
// 	assert.Equal(t, loc, actual.token.Location)
// 	assert.Equal(t, expected.QR, actual.token.QR)
// }

// func TestGetAll(t *testing.T) {
// 	// Prepare validTokens
// 	validTokens := new(ValidTokens)

// 	for i := 0; i < len(tokens); i++ {
// 		validTokens.Store(tokens[i].ID, &tokens[i])
// 	}

// 	// Get all tokens
// 	all := validTokens.GetAll()

// 	found := 0
// 	// Check if tokens are in all.
// 	validTokens.Range(func(key, value interface{}) bool {
// 		val := value.(*AccessToken)
// 		for _, token := range all {
// 			if val.ID == token.ID {
// 				found++

// 				// Tokens must be equal
// 				assert.Equal(t, token, val)
// 				break
// 			}
// 		}

// 		return true
// 	})

// 	// Found must be equal length of tokens.
// 	// All tokens get
// 	assert.Equal(t, len(tokens), found)
// }

// func TestGetAllEmptyMap(t *testing.T) {
// 	validTokens := new(ValidTokens)
// 	expected := make([]*AccessToken, 0)

// 	actual := validTokens.GetAll()
// 	assert.Equal(t, expected, actual)
// }

// func TestGetAcceessTokenForLocation(t *testing.T) {
// 	validTokens := new(ValidTokens)

// 	for i := 0; i < len(tokens); i++ {
// 		validTokens.Store(tokens[i].ID, &tokens[i])
// 	}

// 	expected := &tokens[1]
// 	actual, ok := validTokens.GetAccessTokenForLocation("DHBW Mosbach")

// 	assert.True(t, ok)
// 	assert.Equal(t, expected, actual)
// }

// func TestGetAcceessTokenForLocationNotFound(t *testing.T) {
// 	validTokens := new(ValidTokens)

// 	for i := 0; i < len(tokens); i++ {
// 		validTokens.Store(tokens[i].ID, &tokens[i])
// 	}

// 	actual, ok := validTokens.GetAccessTokenForLocation("Night Club")

// 	assert.False(t, ok)
// 	assert.Nil(t, actual)
// }

// func TestRunValidTokens(t *testing.T) {
// 	validTokens := new(ValidTokens)

// 	tokenQueue := make(chan TokenQueueItem)
// 	log := make(chan TokenQueueItem)

// 	// Run goroutine in background
// 	validTokens.run(tokenQueue, log)

// 	// Add item
// 	addItem := TokenQueueItem{Add, &tokens[0]}
// 	tokenQueue <- addItem
// 	reflectedItem := <-log

// 	assert.Equal(t, addItem, reflectedItem)

// 	value, ok := validTokens.Load(tokens[0].ID)
// 	assert.True(t, ok)

// 	actual, ok := value.(*AccessToken)
// 	assert.True(t, ok)
// 	assert.Equal(t, &tokens[0], actual)

// 	// Remove item
// 	invalidateItem := TokenQueueItem{Invalidate, &tokens[0]}

// 	tokenQueue <- invalidateItem
// 	reflectedItem = <-log

// 	value, ok = validTokens.Load(tokens[0].ID)
// 	assert.False(t, ok)
// 	assert.Nil(t, value)
// }

// func TestGenerateAccessTokens(t *testing.T) {
// 	exp := time.Duration(10) * time.Second
// 	validTokens, log := locations.GenerateAccessTokens(10, exp, "localhost", 8081)

// 	// Wait until all items are stored
// 	for i := 0; i < len(locations.Locations); i++ {
// 		<-log
// 	}

// 	found := 0
// 	for _, loc := range locations.Locations {
// 		validTokens.Range(func(key, value interface{}) bool {
// 			val := value.(*AccessToken)
// 			if val.Location == loc {
// 				found++

// 				// Check if token was created with valid data
// 				assert.Equal(t, 1, val.Valid)
// 				assert.Equal(t, exp, val.Exp.Sub(val.Iat))

// 				qr, err := qrcode.Encode(fmt.Sprintf("localhost:8081?token=%v", val.ID), qrcode.Medium, 256)
// 				assert.NoError(t, err)
// 				assert.Equal(t, qr, val.QR)
// 				return false
// 			}

// 			return true
// 		})
// 	}

// 	// Check if all tokens are found
// 	assert.Equal(t, len(locations.Locations), found)
// }

// func TestContains(t *testing.T) {
// 	actual := locations.Contains("DHBW Mosbach")
// 	assert.True(t, actual)
// }

// func TestContainsNotContains(t *testing.T) {
// 	actual := locations.Contains("Night Club")
// 	assert.False(t, actual)
// }
