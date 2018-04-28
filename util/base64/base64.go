// Package base64 implements base64 encoding related utility functions.
package base64

import (
	"encoding/base64"
)

// Encode returns the base64 encoding of src (URL encoding without padding).
func Encode(src []byte) string {
	return base64.RawURLEncoding.EncodeToString(src)
}

// Decode returns the bytes represented by the base64 string s
// (assuming that s is URL encoded without padding).
func Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
