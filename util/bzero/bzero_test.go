package bzero

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestBytes(t *testing.T) {
	zero := make([]byte, 1024)
	buf := make([]byte, 1024)
	// compare new buffer
	if !bytes.Equal(buf, zero) {
		t.Error("buffers differ")
	}
	// fill buffer with random data
	if _, err := rand.Read(buf); err != nil {
		t.Fatal(err)
	}
	// zero
	Bytes(buf)
	// compare reset buffer
	if !bytes.Equal(buf, zero) {
		t.Error("buffers differ")
	}
}
