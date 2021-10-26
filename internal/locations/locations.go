package locations

import (
	"encoding/xml"
	"fmt"
	"github.com/dateiexplorer/attendancelist/internal/token"

	//"github.com/dateiexplorer/attendancelist/internal/timeutil"
	//"github.com/dateiexplorer/attendancelist/internal/token"
	"github.com/skip2/go-qrcode"
	"io/ioutil"
	"time"
	//"path"

	//"io/ioutil"
	"os"
)

// Locations are a collection of places where people can access
type Locations struct {
	Locations []Location `xml:"Location"`
}

// A Location represents a place where a Person can be associated with.
type Location string

// LocGenerator returns Locations struct where single Location from the given xml file are stored
//
// Used to Convert Data from XML file to Locations struct
func LocGenerator(path string) (Locations) {

	var locations Locations
	file, err := os.Open(path)
	if err != nil {
		fmt.Errorf("can't opne xml file: %w", err)
		//panic()
	}
	defer file.Close()
	bytes, errIO := ioutil.ReadAll(file)
	if errIO != nil {
		fmt.Errorf("cant parse xml: %w",err)
	}
	xml.Unmarshal(bytes, &locations)
	return locations
}

// ValidTokens is a map with all valid AccessToken
type ValidTokens map[string]*AccessToken

// An AccessToken represents a token for a location
type AccessToken struct {
	id string
	location *Location
	expires time.Time
	valid int
	qr []byte
}

func NewAccessToken(loc Location, newId string, expireTime time.Time) AccessToken{
	newToken := AccessToken{}
	newToken.id = newId
	newToken.location = &loc
	newToken.expires = expireTime
	newToken.valid = 2

	//prüfen, ob size mit string passt bei eingabe
	tmp, err :=  qrcode.Encode("http://localhost:8081?token=" + newToken.id, qrcode.Medium, 256)
	if err != nil {
		fmt.Errorf("QRCode can not be generated: %w", err)
	}
	newToken.qr = tmp

	return newToken
}


// AccessTokenValidator triggers the generation of AccessTokens(GenerateNewAT) for the ValidTokens Map
// and handels the validation of the expiretime
// vl new ids übergeben
func AccessTokenValidator(tokens *ValidTokens, path string, sec time.Duration) {
	locs := LocGenerator(path)
	expireTime := make(chan time.Time, 1000)
	var newIds = token.RandIDGenerator(10, 1000)
	go func() {
		for {
		select {
		case now :=<-time.After(sec*time.Second):
			GenerateNewAT(now, locs, newIds, expireTime, sec, tokens)
		case <-expireTime:
			ValidateAT(tokens)
	}
	}}()
}

// GenerateNewAT generates new AccessToken for the ValidTokens Map given by the Locations
func GenerateNewAT(now time.Time, locations Locations, newIds <-chan string, expireTime chan<- time.Time, sec time.Duration, tokens *ValidTokens) {
	tempExp := now.Add(sec*time.Second)
	expireTime <- tempExp
	for _,loc := range locations.Locations {
		newId := <-newIds
		temp := NewAccessToken(loc, newId, tempExp)
		(*tokens)[newId] = &temp
	}
}

// ValidateAT validates the valid state of each AccessToken in the given ValidTokens Map
func ValidateAT(tokens *ValidTokens) {
	for key, val := range *tokens{
		if val.valid == 2 {
			(*tokens)[key].valid = 1
		} else if val.valid == 1 {
			delete(*tokens, key)
		}}
}
