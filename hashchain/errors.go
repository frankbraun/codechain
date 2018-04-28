package hashchain

import (
	"errors"
)

// ErrSignatureThresholdNonPositive is returned when the signature threshold is non-positive.
var ErrSignatureThresholdNonPositive = errors.New("hashchain: signature threshold m must be positive")

// ErrMLargerThanN is returned when m > n.
var ErrMLargerThanN = errors.New("hashchain: signature threshold m is larger than total weight of signers n")
