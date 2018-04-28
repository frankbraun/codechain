package hashchain

import (
	"errors"
)

// ErrSignatureThresholdNonPositive is returned when the signature threshold is non-positive.
var ErrSignatureThresholdNonPositive = errors.New("hashchain: signature threshold m must be positive")

// ErrMLargerThanN is returned when m > n.
var ErrMLargerThanN = errors.New("hashchain: signature threshold m is larger than total weight of signers n")

// ErrEmpty is returned when the hash chain is empty.
var ErrEmpty = errors.New("hashchain: is empty")

// ErrLinkBroken is returned when a link in the hash chain is broken.
var ErrLinkBroken = errors.New("hashchain: link broken")

// ErrDescendingTime is returned when the time in the hash chain is not ascending.
var ErrDescendingTime = errors.New("hashchain: time is going backwards")

// ErrUnknownLinkType is returned when the link type is unknown.
var ErrUnknownLinkType = errors.New("hashchain: unknown link type")
