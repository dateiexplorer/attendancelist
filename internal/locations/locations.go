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
	errUN := xml.Unmarshal(bytes, &locations)
	if errUN != nil {
		fmt.Errorf("cant unmarshal xml: %w", errUN)
	}
	return locations
}

// ValidTokens is a map with all valid AccessToken
type ValidTokens map[string]*AccessToken

// An AccessToken represents a token for a location
type AccessToken struct {
	Id       string
	Location *Location
	Expires  time.Time
	Valid int
	Qr    []byte
}

func NewAccessToken(loc Location, newId string, expireTime time.Time) AccessToken{
	newToken := AccessToken{}
	newToken.Id = newId
	newToken.Location = &loc
	newToken.Expires = expireTime
	newToken.Valid = 2

	//prüfen, ob size mit string passt bei eingabe
	tmp, err :=  qrcode.Encode("http://localhost:8081?token=" + newToken.Id, qrcode.Medium, 256)
	if err != nil {
		fmt.Errorf("QRCode can not be generated: %w", err)
	}
	newToken.Qr = tmp

	return newToken
}


// AccessTokenValidator triggers the generation of AccessTokens(GenerateNewAT) for the ValidTokens Map
// and handels the validation of the expiretime
// vl new ids übergeben
func AccessTokenValidator(tokens *ValidTokens, path string, sec time.Duration) Locations{
	locs := LocGenerator(path)
	loc ,_ :=time.LoadLocation("Europe/Berlin")
	expireTime := make(chan time.Time, 1000)
	var newIds = token.RandIDGenerator(10, 1000)
	go func() {
		for {
		select {
		case <-time.After(sec):
			now := time.Now().In(loc)
			go ValidateAT(tokens)
			go GenerateNewAT(now, locs, newIds, expireTime, sec, tokens)
	}
	}}()
	return locs
}

// GenerateNewAT generates new AccessToken for the ValidTokens Map given by the Locations
func GenerateNewAT(now time.Time, locations Locations, newIds <-chan string, expireTime chan<- time.Time, sec time.Duration, tokens *ValidTokens) {
	tempExp := now.Add(sec)
	expireTime <- tempExp
	for _,loc := range locations.Locations {
		newId := <-newIds
		temp := NewAccessToken(loc, newId, tempExp)

		(*tokens)[newId] = &temp
	}
	//fmt.Printf("len: %v\n" , len(*tokens))
}

// ValidateAT validates the valid state of each AccessToken in the given ValidTokens Map
func ValidateAT(tokens *ValidTokens) {
	for key, val := range *tokens{
		if val.Valid == 2 {
			(*tokens)[key].Valid = 1
		} else if val.Valid == 1 {
			delete(*tokens, key)
		}}
}
