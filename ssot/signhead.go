package ssot

import (
	"time"

	"golang.org/x/crypto/ed25519"
)

// SignHead signs the given Codechain head.
func SignHead(
	head [32]byte,
	counter uint64,
	secKey [64]byte,
	validity time.Duration,
) (*SignedHead, error) {
	var sh SignedHead
	copy(sh.pubKey[:], secKey[32:])
	// TODO: allow to set pubKeyRotate
	now := time.Now().UTC().Unix()
	sh.validFrom = now
	if validity > MaximumValidity {
		return nil, ErrValidityTooLong
	}
	if validity < MinimumValidity {
		return nil, ErrValidityTooShort
	}
	sh.validTo = now + int64(validity)
	sh.counter = counter
	copy(sh.head[:], head[:])
	msg := sh.marshal()
	sig := ed25519.Sign(secKey[:], msg[:])
	copy(sh.signature[:], sig)
	return &sh, nil
}
