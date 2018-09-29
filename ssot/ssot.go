// Package ssot implements a single source of truth (SSOT)
// with DNS TXT records.
package ssot

import (
	"encoding/binary"
	"time"

	"github.com/frankbraun/codechain/internal/base64"
	"golang.org/x/crypto/ed25519"
)

// MaximumValidity of signed heads.
const MaximumValidity = 30 * 24 * 60 * 60 // 30d

// SignedHead is a signed Codechain head ready for publication as a SSOT with
// DNS TXT records.
type SignedHead struct {
	pubKey       [32]byte // Ed25519 public key of SSOT head signer
	pubKeyRotate [32]byte // Ed25519 pubkey to rotate to, all 0 if unused
	validFrom    int64    // this signed head is valid from the given Unix time
	validTo      int64    // this signed head is valid to the given Unix time
	counter      uint64   // signature counter
	head         [32]byte // the Codechain head to sign
	signature    [64]byte // signature with pubkey over all previous fields
}

// marshal signed head without signature.
func (sh *SignedHead) marshal() [120]byte {
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

// SignHead signs the given Codechain head.
func SignHead(head [32]byte, counter uint64, secKey [64]byte) *SignedHead {
	var sh SignedHead
	copy(sh.pubKey[:], secKey[32:])
	// TODO: allow to set pubKeyRotate
	now := time.Now().UTC().Unix()
	// TODO: allow to set validFrom and validTo
	sh.validFrom = now
	sh.validTo = now + MaximumValidity
	sh.counter = counter
	copy(sh.head[:], head[:])
	m := sh.marshal()
	sig := ed25519.Sign(secKey[:], m[:])
	copy(sh.signature[:], sig)
	return &sh
}

// Marshal signed head with signature and encode it as base64.
func (sh *SignedHead) Marshal() string {
	var m [184]byte
	b := sh.marshal()
	copy(m[:120], b[:])
	copy(m[120:184], sh.signature[:])
	return base64.Encode(m[:])
}
