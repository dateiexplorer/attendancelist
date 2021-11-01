// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

package secure

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// RandIdGenerator returns a read-only channel which returns strings that can be
// used as hexadecimal IDs.
//
// The idLength determines the length of the produced IDs, the buffer determines
// the internal buffer size of the channel which will be useful to hold up some
// already generated strings which makes the generator ready for concurrent
// channel reads.
//
// Notice that the generator does not guarantee that the generated IDs are unique
// but a larger idLength (>= 8) causes IDs that are effectivly unique.
//
// Credits to this post https://gist.github.com/arxdsilva/8caeca47b126a290c4562a25464895e8
// by arxdsilva which is the inspiration source for this generator.
func RandIDGenerator(idLength, buffer int) <-chan string {
	ids := make(chan string, buffer)
	go func() {
		t := make([]byte, idLength)
		for {
			rand.Read(t)
			// Make sure id length matches idLength cause "%x" produces only strings
			// with even length
			ids <- string([]byte(fmt.Sprintf("%x", t))[:idLength])
		}
	}()

	return ids
}

// Hash hashes the values of a type v with a private key privkey with the sha256
// algorithm and returns the hash in a hexadecimal representation as a string of
// 64 chars.
//
// This method uses the json.Marshal function to representate any data as a byte
// slice, which is necessary for the alogithm.
// Same input produces same hashes.
//
// If some values of a type can't be marshaled by the json.Marshal function this
// functions returns an error and an empty string.
func Hash(v interface{}, privkey string) (string, error) {
	hashData := struct {
		V       interface{} `json:"value"`
		Privkey string      `json:"key"`
	}{
		V:       v,
		Privkey: privkey,
	}

	bytes, err := json.Marshal(hashData)
	if err != nil {
		return "", fmt.Errorf("cannot hash type %T: %w", v, err)
	}

	hash := sha256.Sum256(bytes)
	return fmt.Sprintf("%x", hash), nil
}
