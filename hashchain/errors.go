package hashchain

import (
	"errors"
)

// ErrSignatureThresholdNonPositive is returned when the signature threshold is non-positive.
var ErrSignatureThresholdNonPositive = errors.New("hashchain: signature threshold M must be positive")
