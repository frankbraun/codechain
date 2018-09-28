package sync

import (
	"errors"
)

// ErrCannotRemove is returned if Dir could not find a valid start to apply,
// but cannot remove the directory.
var ErrCannotRemove = errors.New("sync: could not find a valid start to apply, try with empty dir")
