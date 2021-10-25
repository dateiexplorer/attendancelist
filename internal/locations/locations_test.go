package locations

import (
	"fmt"
	"github.com/dateiexplorer/attendancelist/internal/token"
	"github.com/skip2/go-qrcode"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
	"time"
)

//Data

var locations = Locations{[]Location{"DHBW MOSBACH", "Alte Mältzerei"}}

func TestLocGenerator(t *testing.T) {
	expected := locations
	actual := LocGenerator(path.Join("testdata","locations.xml"))

	assert.EqualValues(t,expected, actual,"wrong locations")
}

func TestLocGeneratorNoFile(t *testing.T) {
	expected := Locations{}
	actual := LocGenerator(path.Join("falsePath","locations.xml"))

	assert.EqualValues(t,expected, actual,"wrong locations")
}

func TestNewAccessToken(t *testing.T) {
	tempId := <-token.RandIDGenerator(10, 10)
	tmp, err := qrcode.Encode("http://localhost:8081?token=" + tempId, qrcode.Medium, 256)
	loc := locations.Locations[0]
	if err != nil {
		fmt.Println(err)
	}
	exp := time.Now()
	expected := AccessToken{
		id:       tempId,
		location: &loc,
		expires:  exp,
		valid:    2,
		qr: tmp,
	}
	actual := NewAccessToken(loc, tempId, exp)
	assert.EqualValues(t, expected, actual, "wrong AccessToken")
}

func TestGenerateNewAT(t *testing.T) {
	locs := LocGenerator(path.Join("testdata","locations.xml"))
	var newIds = token.RandIDGenerator(10, 1000)
	newId1 := <-newIds
	newId2 := <-newIds
	now := time.Now()
	exp := now.Add(100*time.Second)
	tmpIDs:= make(chan string, 2)
	tmpExp:= make(chan time.Time, 2)
	tmpIDs<-newId1
	tmpIDs<-newId2
	expAT1 := NewAccessToken(locs.Locations[0], newId1, exp)
	expAT2 := NewAccessToken(locs.Locations[1], newId2, exp)
	expected := ValidTokens{newId1: &expAT1, newId2: &expAT2}

	actual := ValidTokens{}
	GenerateNewAT(now, locs, tmpIDs, tmpExp, time.Duration(100), &actual)

	assert.EqualValues(t, expected, actual, "no right generation")
}

func TestGenerateNewATNegative(t *testing.T) {
	locs := LocGenerator(path.Join("testdata","locations.xml"))
	var newIds = token.RandIDGenerator(10, 1000)
	newId1 := <-newIds
	newId2 := <-newIds
	now := time.Now()
	exp := now.Add(50*time.Second)
	tmpIDs:= make(chan string, 2)
	tmpExp:= make(chan time.Time, 2)
	tmpIDs<-newId1
	tmpIDs<-newId2
	expAT1 := NewAccessToken(locs.Locations[0], newId1, exp)
	expAT2 := NewAccessToken(locs.Locations[1], newId2, exp)
	expected := ValidTokens{newId1: &expAT1, newId2: &expAT2}

	actual := ValidTokens{}
	GenerateNewAT(now, locs, tmpIDs, tmpExp, time.Duration(100), &actual)

	assert.NotEqualValues(t, expected, actual, "no right generation")
}

/*func TestValidateAT(t *testing.T) {
	locs := LocGenerator(path.Join("testdata","locations.xml"))
	var newIds = token.RandIDGenerator(10, 1000)
	now := time.Now()
	tmpExp:= make(chan time.Time, 2)
	actual := ValidTokens{}
	ValidateAT()
}*/


