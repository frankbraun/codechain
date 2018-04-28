// Package base64 implements base64 encoding related utility functions.
package base64

import (
	"encoding/base64"
	"fmt"
)

// Encode returns the base64 encoding of src (URL encoding without padding).
func Encode(src []byte) string {
	return base64.RawURLEncoding.EncodeToString(src)
}

// Decode returns the bytes represented by the base64 string s
// (assuming that s is URL encoded without padding).
// Decode expects that the resulting byte slice has length l.
func Decode(s string, l int) ([]byte, error) {
	r, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(r) != l {
		return nil, fmt.Errorf("base64: wrong length %d (expecting %d): %s", 2*len(r), 2*l, s)
	}
	return r, nil
}
