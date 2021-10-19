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
