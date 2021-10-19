package token

import (
	"crypto/rand"
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
