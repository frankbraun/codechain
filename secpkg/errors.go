package secpkg

import (
	"errors"
)

// ErrPkgNameWhitespace is returned if a package name contains a white space character.
var ErrPkgNameWhitespace = errors.New("secpkg: package name contains white space character")
