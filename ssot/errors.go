package ssot

import (
	"errors"
)

// ErrPkgNameWhis returned if a signed head signature does not
// verify.
var ErrSignedHeadSignature = errors.New("ssot: signed head signature does not verify")
