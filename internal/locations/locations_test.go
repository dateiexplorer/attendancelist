package locations

import (
	"fmt"
	"github.com/dateiexplorer/attendancelist/internal/token"
	"github.com/skip2/go-qrcode"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

//Data

var locations = Locations{[]Location{"DHBW MOSBACH", "Alte Mältzerei"}}

func TestLocGenerator(t *testing.T) {
	expected := locations
	actual := LocGenerator("testdata\\locations.xml")

	assert.EqualValues(t,expected, actual,"wrong locations")
}

func TestLocGeneratorNoFile(t *testing.T) {
	expected := Locations{}
	actual := LocGenerator("falsePath\\locations.xml")

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


