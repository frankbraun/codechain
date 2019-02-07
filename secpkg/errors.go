package secpkg

import (
	"errors"
)

// ErrPkgNameWhitespace is returned if a package name contains a white space character.
var ErrPkgNameWhitespace = errors.New("secpkg: package name contains white space character")

// ErrNoKey is returned if a package has no secretbox encryption key.
var ErrNoKey = errors.New("secpkg: package has no secretbox encryption key")
