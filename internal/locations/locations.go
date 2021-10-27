package locations

import (
	"encoding/xml"
	"fmt"

	"io/ioutil"
	"os"
	"time"

	"github.com/dateiexplorer/attendancelist/internal/journal"
	"github.com/dateiexplorer/attendancelist/internal/token"
	"github.com/skip2/go-qrcode"
)

// Locations are a collection of places that people can access
type Locations struct {
	locations []journal.Location `xml:"Location"`
}

// LocGenerator returns Locations struct where single Location from the given XML file are stored
//
// Used to convert data from XML file to Locations struct
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

// ValidTokens is a map with all valid AccessToken
type ValidTokens map[string]*AccessToken

// An AccessToken represents a token for a location
type AccessToken struct {
	id       string
	location *journal.Location
	expires  time.Time
	valid    int
	qr       []byte
}

func NewAccessToken(loc *journal.Location, id string, expireTime time.Time, baseUrl string, port int) AccessToken {
	token := AccessToken{id: id, location: loc, expires: expireTime, valid: 2}

	qr, err := qrcode.Encode(fmt.Sprintf("%v:%v?token=%v", baseUrl, port, token.id), qrcode.Medium, 256)
	if err != nil {
		panic(fmt.Errorf("Cannot create qr code: %w", err))
	}

	token.qr = qr
	return token
}

func (l *Location) NewAccessTokenGenerator(validTokens *ValidTokens, idGenerator <-chan string, expireInterval time.Duration, baseUrl string, port int) {
	go func() {
		token := NewAccessToken(l, <-idGenerator, time.Now().UTC().Add(expireInterval), baseUrl, port)
		for token.valid > 0 {
			select {
			case <-time.After(expireInterval):
				token.valid--
			}
		}

		delete(*validTokens, token.id)
	}()

	// locs := LocGenerator(path)
	// var expireTime chan time.Time

	//1channel expiretime countdown, wie lang bis expired
	//select
	for {
		select {
		case now := <-time.After(time.Duration(sec) * time.Second):
			expireTime <- now.Add(time.Duration(sec) * time.Second)
			for _, loc := range locs.Locations {
				newId := <-token.RandIDGenerator(10, 1000)
				temp := NewAccessToken(loc, newId, <-expireTime)
				tokens[newId] = &temp
			}
		case <-expireTime:
			for key, val := range tokens {
				if val.valid == 2 {
					tokens[key].valid = 1

				} else if val.valid == 1 { // && endzeit am ende des tages in der map?
					delete(tokens, key)
				}
			}
		}
		/*for {
		validationStart = time.Now()
		expireTime = validationStart.Add(time.Second*time.Duration(sec))

		}



		sleepTime = expireTime.Sub(time.Now())
		fmt.Println(len(tokens) + expireTime.Second())
		time.Sleep(sleepTime)*/
		return expireTime
	}
}
