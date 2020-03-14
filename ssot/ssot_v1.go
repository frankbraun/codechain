package ssot

import (
	"crypto/ed25519"
	"encoding/binary"

	"github.com/frankbraun/codechain/util/base64"
	"github.com/frankbraun/codechain/util/hex"
)

// SignedHeadV1 is a signed Codechain head ready for publication as a SSOT with
// DNS TXT records (version 1).
type SignedHeadV1 struct {
	pubKey       [32]byte // Ed25519 public key of SSOT head signer
	pubKeyRotate [32]byte // Ed25519 pubkey to rotate to, all 0 if unused
	validFrom    int64    // this signed head is valid from the given Unix time
	validTo      int64    // this signed head is valid to the given Unix time
	counter      uint64   // signature counter
	head         [32]byte // the Codechain head to sign
	signature    [64]byte // signature with pubkey over all previous fields
}

// marshal signed head without signature.
func (sh *SignedHeadV1) marshal() [120]byte {
	var m [120]byte
	var b [8]byte
	copy(m[:32], sh.pubKey[:])
	copy(m[32:64], sh.pubKeyRotate[:])
	binary.BigEndian.PutUint64(b[:], uint64(sh.validFrom))
	copy(m[64:72], b[:])
	binary.BigEndian.PutUint64(b[:], uint64(sh.validTo))
	copy(m[72:80], b[:])
	binary.BigEndian.PutUint64(b[:], sh.counter)
	copy(m[80:88], b[:])
	copy(m[88:120], sh.head[:])
	return m
}

// Marshal signed head with signature and encode it as base64.
func (sh *SignedHeadV1) Marshal() string {
	var m [184]byte
	b := sh.marshal()
	copy(m[:120], b[:])
	copy(m[120:184], sh.signature[:])
	return base64.Encode(m[:])
}

func unmarshalV1(m [184]byte) (*SignedHeadV1, error) {
	var sh SignedHeadV1
	copy(sh.pubKey[:], m[:32])
	copy(sh.pubKeyRotate[:], m[32:64])
	sh.validFrom = int64(binary.BigEndian.Uint64(m[64:72]))
	sh.validTo = int64(binary.BigEndian.Uint64(m[72:80]))
	sh.counter = binary.BigEndian.Uint64(m[80:88])
	copy(sh.head[:], m[88:120])
	copy(sh.signature[:], m[120:184])
	msg := sh.marshal()
	if !ed25519.Verify(sh.pubKey[:], msg[:], sh.signature[:]) {
		return nil, ErrSignedHeadSignature
	}
	return &sh, nil
}

// Version returns the version of signed head.
func (sh *SignedHeadV1) Version() int {
	return 1
}

// Head returns the signed head.
func (sh *SignedHeadV1) Head() string {
	return hex.Encode(sh.head[:])
}

// PubKey returns the public key in base64 notation.
func (sh *SignedHeadV1) PubKey() string {
	return base64.Encode(sh.pubKey[:])
}

// PubKeyRotate returns the public key rotate in base64 notation.
func (sh *SignedHeadV1) PubKeyRotate() string {
	return base64.Encode(sh.pubKeyRotate[:])
}

// ValidFrom returns the valid from field of signed head.
func (sh *SignedHeadV1) ValidFrom() int64 {
	return sh.validFrom
}

// ValidTo returns the valid to field of signed head.
func (sh *SignedHeadV1) ValidTo() int64 {
	return sh.validTo
}

// Counter returns the counter of signed head.
func (sh *SignedHeadV1) Counter() uint64 {
	return sh.counter
}

// Line always returns 0 (signed head version 1 doesn't contain line numbers,
// but this method is required to satisfy the SignedHead interface).
func (sh *SignedHeadV1) Line() int {
	return 0
}

// Signature returns the base64-encoded signature of the signed head.
func (sh *SignedHeadV1) Signature() string {
	return base64.Encode(sh.signature[:])
}

// HeadBuf returns the signed head.
func (sh *SignedHeadV1) HeadBuf() [32]byte {
	var b [32]byte
	copy(b[:], sh.head[:])
	return b
}
