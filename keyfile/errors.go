package keyfile

import (
	"errors"
)

// ErrDecrypt is returned if the keyfile could not be decrypted (wrong passphrase).
var ErrDecrypt = errors.New("keyfile: cannot decrypt")
