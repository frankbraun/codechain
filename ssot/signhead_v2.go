package ssot

import (
	"crypto/ed25519"
	"time"
)

// SignHeadV2 signs the given Codechain head.
func SignHeadV2(
	head [32]byte,
	line int,
	counter uint64,
	secKey [64]byte,
	pubKeyRotate *[32]byte,
	validity time.Duration,
) (SignedHead, error) {
	var sh SignedHeadV2
	sh.version = 2
	copy(sh.pubKey[:], secKey[32:])
	if pubKeyRotate != nil {
		copy(sh.pubKeyRotate[:], pubKeyRotate[:])
	}
	now := time.Now().UTC().Unix()
	sh.validFrom = now
	if validity > MaximumValidity {
		return nil, ErrValidityTooLong
	}
	if validity < MinimumValidity {
		return nil, ErrValidityTooShort
	}
	sh.validTo = now + int64(validity/time.Second)
	sh.counter = counter
	copy(sh.head[:], head[:])
	sh.line = uint32(line)
	msg := sh.marshal()
	sig := ed25519.Sign(secKey[:], msg[:])
	copy(sh.signature[:], sig)
	return &sh, nil
}
