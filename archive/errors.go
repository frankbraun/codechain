package archive

import (
	"errors"
)

// ErrUnknownFile is returned if an archive contains an unknown file.
var ErrUnknownFile = errors.New("archive: contains unknown file, not a codechain archive?")

// ErrCannotDecrypt is returned if an encrypted archive cannot be decrypted.
var ErrCannotDecrypt = errors.New("archive: cannot decrypt")
