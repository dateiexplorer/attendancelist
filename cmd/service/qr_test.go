// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dateiexplorer/attendancelist/internal/timeutil"
	"github.com/dateiexplorer/attendancelist/internal/web"
	"github.com/stretchr/testify/assert"
)

func TestGetTokens(t *testing.T) {
	tokens := []*web.AccessToken{
		{ID: "a", Iat: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 0).Time, Exp: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 10).Time, Valid: 0, Location: "DHBW Mosbach", QR: []byte{}},
		{ID: "b", Iat: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 10).Time, Exp: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 20).Time, Valid: 1, Location: "DHBW Mosbach", QR: []byte{}},
		{ID: "c", Iat: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 0).Time, Exp: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 10).Time, Valid: 0, Location: "Alte Mälzerei", QR: []byte{}},
		{ID: "d", Iat: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 10).Time, Exp: timeutil.NewTimestamp(2021, 11, 30, 0, 0, 20).Time, Valid: 1, Location: "Alte Mälzerei", QR: []byte{}},
	}

	validTokens := new(web.ValidTokens)
	validTokens.Add(tokens)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getToken(w, r, validTokens)
	}))
	defer ts.Close()

	// Get all tokens
	res, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var actual []*web.AccessToken
	err = json.Unmarshal(body, &actual)
	assert.NoError(t, err)

	assert.Equal(t, len(tokens), len(actual))
	for _, a := range actual {
		assert.Contains(t, tokens, a)
	}

	// Get newest token for specific location
	jsonToken, err := json.Marshal(tokens[1])
	assert.NoError(t, err)

	res, err = http.Get(fmt.Sprintf("%v/?location=%v", ts.URL, "DHBW+Mosbach"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err = ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, jsonToken, body)

	// Get error if location doesn't exist
	res, err = http.Get(fmt.Sprintf("%v/?location=%v", ts.URL, "Night+Club"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err = ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, "{\"err\":\"no valid token found for this location\"}", string(body))
}
