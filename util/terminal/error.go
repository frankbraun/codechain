package terminal

import (
	"errors"
)

// ErrAbort is returned if a user answers 'n' to Confirm.
var ErrAbort = errors.New("aborted")
