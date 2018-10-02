package ssot

import (
	"time"
)

// Valid checks if the signed head sh is currently valid
// (as defined by validFrom and validTo).
// It returns nil, if the signed check is valid and an error otherwise.
func (sh *SignedHead) Valid() error {
	now := time.Now().UTC().Unix()
	if now < sh.validFrom {
		return ErrSignedHeadFuture
	}
	if now > sh.validTo {
		return ErrSignedHeadExpired
	}
	return nil
}
