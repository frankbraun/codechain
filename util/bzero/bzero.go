// Package bzero defines helper functions to zero sensitive memory.
package bzero

// Bytes sets all entries in the given byte slice buffer to zero.
func Bytes(buffer []byte) {
	for i := 0; i < len(buffer); i++ {
		buffer[i] = 0
	}
}
