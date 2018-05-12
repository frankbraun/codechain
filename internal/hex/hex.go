// Package hex implements hex encoding related utility functions.
package hex

import (
	"encoding/hex"
	"fmt"
)

// Encode returns the hexadecimal encoding of src.
func Encode(src []byte) string {
	return hex.EncodeToString(src)
}

// Decode returns the bytes represented by the hexadecimal string s. Decode
// expects that src contain only lowercase hexadecimal characters and that the
// resulting byte slice has length l.
func Decode(s string, l int) ([]byte, error) {
	for _, c := range []byte(s) {
		if 'A' <= c && c <= 'F' {
			return nil, fmt.Errorf("hex: only lowercase hexadecimal characters are allowed")
		}
	}
	r, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(r) != l {
		return nil, fmt.Errorf("hex: wrong length %d (expecting %d): %s", 2*len(r), 2*l, s)
	}
	return r, nil
}
