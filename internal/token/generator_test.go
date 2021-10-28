// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
