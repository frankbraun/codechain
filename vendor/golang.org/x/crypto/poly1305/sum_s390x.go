// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build s390x,go1.11,!gccgo,!appengine

package poly1305

// Sum generates an authenticator for m using a one-time key and puts the
// 16-byte result into out. Authenticating two different messages with the same
// key allows an attacker to forge messages at will.
func Sum(out *[16]byte, m []byte, key *[32]byte) {
	if hasVX {
		var mPtr *byte
		if len(m) > 0 {
			mPtr = &m[0]
		}
		if hasVMSL && len(m) > 256 {
			poly1305vmsl(out, mPtr, uint64(len(m)), key)
		} else {
			poly1305vx(out, mPtr, uint64(len(m)), key)
		}
	} else {
		sumGeneric(out, m, key)
	}
}
