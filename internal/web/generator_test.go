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
	"errors"
	"testing"

	"github.com/dateiexplorer/dhbw-attendancelist/internal/journal"
	"github.com/stretchr/testify/assert"
)

type test struct{}

func (t test) MarshalJSON() ([]byte, error) {
	return nil, errors.New("cannot marshal test data")
}

func TestRandIdGeneratorEvenId(t *testing.T) {
	stream := RandIDGenerator(8, 10)
	for i := 0; i < 10; i++ {
		assert.Equal(t, 8, len(<-stream))
	}
}

func TestRandIdGeneratorOddId(t *testing.T) {
	stream := RandIDGenerator(5, 10)
	for i := 0; i < 10; i++ {
		assert.Equal(t, 5, len(<-stream))
	}
}

func TestHash(t *testing.T) {
	person1 := journal.NewPerson("Max", "Mustermann", "Musterstraße", "20", "74821", "Mosbach")
	person2 := journal.NewPerson("Max", "Mustermann", "Musterstraße", "20", "74821", "Mosbach")
	privkey := "privServerSecret"

	hashPerson1, err := Hash(person1, privkey)
	assert.NoError(t, err)

	// To hashes with the same data should be equal
	hashPerson2, err := Hash(person2, privkey)

	assert.NoError(t, err)
	assert.Equal(t, hashPerson1, hashPerson2)
}

func TestHashShouldNotBeEqual(t *testing.T) {
	person := journal.NewPerson("Max", "Mustermann", "Musterstraße", "20", "74821", "Mosbach")
	privkey1 := "privServerSecret1"
	privkey2 := "privServerSecret2"

	hash1, err := Hash(person, privkey1)
	assert.NoError(t, err)

	hash2, err := Hash(person, privkey2)
	assert.NoError(t, err)

	// To hashes with unequal data should be unequal
	assert.NotEqual(t, hash1, hash2)
}

func TestHashError(t *testing.T) {
	test := test{}
	hash, err := Hash(test, "privServerSecret")

	assert.Error(t, err)
	assert.Equal(t, "", hash)
}
