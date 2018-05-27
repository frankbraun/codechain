// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package blake2b

import (
	"encoding/binary"
)

// magicUnknownOutputLength is a magic value for the output size that indicates
// an unknown number of output bytes.
const magicUnknownOutputLength = (1 << 32) - 1

// maxOutputLength is the absolute maximum number of bytes to produce when the
// number of output bytes is unknown.
const maxOutputLength = (1 << 32) * 64

type xof struct {
	d                digest
	length           uint32
	remaining        uint64
	cfg, root, block [Size]byte
	offset           int
	nodeOffset       uint32
	readMode         bool
}

func (x *xof) Write(p []byte) (n int, err error) {
	if x.readMode {
		panic("blake2b: write to XOF after read")
	}
	return x.d.Write(p)
}

func (x *xof) Reset() {
	x.cfg[0] = byte(Size)
	binary.LittleEndian.PutUint32(x.cfg[4:], uint32(Size)) // leaf length
	binary.LittleEndian.PutUint32(x.cfg[12:], x.length)    // XOF length
	x.cfg[17] = byte(Size)                                 // inner hash size

	x.d.Reset()
	x.d.h[1] ^= uint64(x.length) << 32

	x.remaining = uint64(x.length)
	if x.remaining == magicUnknownOutputLength {
		x.remaining = maxOutputLength
	}
	x.offset, x.nodeOffset = 0, 0
	x.readMode = false
}
