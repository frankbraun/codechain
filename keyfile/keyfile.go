// Package keyfile provides encrypted secret key storage.
package keyfile

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/frankbraun/codechain/util/file"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
)

// Create keyfile.
func Create(filename string, pass, sec, sig, comment []byte) error {
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
	if l := len(sec); l != 64 {
		return fmt.Errorf("keyfile: bad sec length: " + strconv.Itoa(l))
	}
	if l := len(sig); l != 64 {
		return fmt.Errorf("keyfile: bad sig length: " + strconv.Itoa(l))
	}
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return err
	}
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}
	derivedKey := argon2.IDKey(pass, salt[:], 1, 64*1024, 4, 32)
	copy(key[:], derivedKey)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	msg := append(sec, sig...)
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

// Read keyfile.
func Read(filename string, pass []byte) ([]byte, []byte, []byte, error) {
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
	derivedKey := argon2.IDKey(pass, salt[:], 1, 64*1024, 4, 32)
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
	return sec[:], sig[:], msg[128:], nil
}
