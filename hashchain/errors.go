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

// ErrIllegalCStart is returned when a cstart entry appears in a different row
// than row 1.
var ErrIllegalCStart = errors.New("hashchain: cstart is only allowed on start")

// ErrMustStartWithCStart is returned when the hash chain doess not start with
// a cstart entry.
var ErrMustStartWithCStart = errors.New("hashchain: must start with cstart")

// ErrWrongTypeFields is returned when a hash chain entry has the wrong number of type fields.
var ErrWrongTypeFields = errors.New("hashchain: entry has wrong number of type fields")

// ErrWrongSigCStart is returned when the signature of a cstart entry doesn't validate.
var ErrWrongSigCStart = errors.New("hashchain: cstart signature doesn't validate")

// ErrWrongSigSource is returned when the signature of a source entry doesn't validate.
var ErrWrongSigSource = errors.New("hashchain: source signature doesn't validate")

// ErrWrongSigAddKey is returned when the signature of an addkey entry doesn't validate.
var ErrWrongSigAddKey = errors.New("hashchain: addkey signature doesn't validate")

// ErrWrongSigSignature is returned when the signature of a signature entry doesn't validate.
var ErrWrongSigSignature = errors.New("hashchain: signature signature doesn't validate")

// ErrCannotMerge is returned if two hash chains cannot be merged.
var ErrCannotMerge = errors.New("hashchain: cannot merge")

// ErrNothingToMerge is returned if there is nothing to merge.
var ErrNothingToMerge = errors.New("hashchain: nothing to merge")

// ErrHeadNotFound is returned if the head could not be found in hash chain.
var ErrHeadNotFound = errors.New("hashchain: head not found")
