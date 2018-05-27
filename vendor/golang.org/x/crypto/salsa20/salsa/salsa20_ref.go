// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64 appengine gccgo

package salsa

// XORKeyStream crypts bytes from in to out using the given key and counters.
// In and out must overlap entirely or not at all. Counter
// contains the raw salsa20 counter bytes (both nonce and block counter).
func XORKeyStream(out, in []byte, counter *[16]byte, key *[32]byte) {
	var block [64]byte
	var counterCopy [16]byte
	copy(counterCopy[:], counter[:])

	for len(in) >= 64 {
		core(&block, &counterCopy, key, &Sigma)
		for i, x := range block {
			out[i] = in[i] ^ x
		}
		u := uint32(1)
		for i := 8; i < 16; i++ {
			u += uint32(counterCopy[i])
			counterCopy[i] = byte(u)
			u >>= 8
		}
		in = in[64:]
		out = out[64:]
	}

	if len(in) > 0 {
		core(&block, &counterCopy, key, &Sigma)
		for i, v := range in {
			out[i] = v ^ block[i]
		}
	}
}
