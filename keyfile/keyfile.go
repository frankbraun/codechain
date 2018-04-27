// Package keyfile provides encrypted secret key storage.
package keyfile

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/frankbraun/codechain/util/file"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
)

// Create keyfile (encrypted with passphrase) and store secretKey, signature,
// and optional comment it.
func Create(filename string, passphrase []byte, secretKey, signature [64]byte, comment []byte) error {
	var (
		salt  [32]byte
		nonce [24]byte
		key   [32]byte
	)
	exists, err := file.Exists(filename)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("file '%s' exists already", filename)
	}
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return err
	}
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}
	derivedKey := argon2.IDKey(passphrase, salt[:], 1, 64*1024, 4, 32)
	copy(key[:], derivedKey)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	msg := append(secretKey[:], signature[:]...)
	msg = append(msg, comment...)
	enc := secretbox.Seal(nil, msg, &nonce, &key)
	if _, err := f.Write(salt[:]); err != nil {
		return err
	}
	if _, err := f.Write(nonce[:]); err != nil {
		return err
	}
	if _, err := f.Write(enc); err != nil {
		return err
	}
	return nil
}

// Read keyfile (encrypted with passphrase) and return secretKey, signature,
// and optional comment.
func Read(filename string, passphrase []byte) (*[64]byte, *[64]byte, []byte, error) {
	var (
		salt  [32]byte
		nonce [24]byte
		key   [32]byte
	)
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	defer f.Close()
	if _, err := f.Read(salt[:]); err != nil {
		return nil, nil, nil, err
	}
	if _, err := f.Read(nonce[:]); err != nil {
		return nil, nil, nil, err
	}
	derivedKey := argon2.IDKey(passphrase, salt[:], 1, 64*1024, 4, 32)
	copy(key[:], derivedKey)
	enc, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, nil, nil, err
	}
	msg, verify := secretbox.Open(nil, enc, &nonce, &key)
	if !verify {
		return nil, nil, nil, fmt.Errorf("cannot decrypt '%s'", filename)
	}
	var sec [64]byte
	var sig [64]byte
	copy(sec[:], msg[:64])
	copy(sig[:], msg[64:128])
	return &sec, &sig, msg[128:], nil
}
