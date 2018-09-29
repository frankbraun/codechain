package ssot

import (
	"time"

	"golang.org/x/crypto/ed25519"
)

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
	msg := sh.marshal()
	sig := ed25519.Sign(secKey[:], msg[:])
	copy(sh.signature[:], sig)
	return &sh
}
