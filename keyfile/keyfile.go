// Package keyfile provides encrypted secret key storage.
package keyfile

import (
	"bytes"
	"crypto/rand"
	b64 "encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/frankbraun/codechain/internal/base64"
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
	enc := secretbox.Seal(append(salt[:], nonce[:]...), msg, &nonce, &key)
	_, err = fmt.Fprintf(f, "%s %s", base64.Encode(secretKey[32:]),
		base64.Encode(signature[:]))
	if err != nil {
		return err
	}
	if comment != nil {
		_, err := fmt.Fprintf(f, " %s", comment)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(f, "\n%s\n", base64.Encode(enc))
	if err != nil {
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
	c, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	lines := bytes.SplitN(c, []byte("\n"), 2)
	line0 := strings.SplitN(string(lines[0]), " ", 3)
	line1 := string(bytes.TrimSpace(lines[1]))
	pub, err := base64.Decode(line0[0], 32)
	if err != nil {
		return nil, nil, nil, err
	}
	sig, err := base64.Decode(line0[1], 64)
	if err != nil {
		return nil, nil, nil, err
	}
	comment := line0[2]
	r, err := b64.RawURLEncoding.DecodeString(line1)
	if err != nil {
		return nil, nil, nil, err
	}
	expected := len(salt) + len(nonce) + 64 + 64 + secretbox.Overhead
	if len(r) < expected {
		return nil, nil, nil,
			fmt.Errorf("base64: wrong length %d (expecting at least %d): %s",
				2*len(r), 2*expected, line1)
	}
	copy(salt[:], r[:32])
	copy(nonce[:], r[32:56])
	enc := r[56:]
	derivedKey := argon2.IDKey(passphrase, salt[:], 1, 64*1024, 4, 32)
	copy(key[:], derivedKey)
	msg, verify := secretbox.Open(nil, enc, &nonce, &key)
	if !verify {
		return nil, nil, nil, ErrDecrypt
	}
	var sec [64]byte
	var decSig [64]byte
	copy(sec[:], msg[:64])
	copy(decSig[:], msg[64:128])
	decComment := msg[128:]
	if !bytes.Equal(sec[32:], pub) {
		return nil, nil, nil, fmt.Errorf("%s: public keys don't match", filename)
	}
	if !bytes.Equal(decSig[:], sig) {
		return nil, nil, nil, fmt.Errorf("%s: signatures don't match", filename)
	}
	if string(decComment) != comment {
		return nil, nil, nil, fmt.Errorf("%s: signatures don't match", filename)
	}
	return &sec, &decSig, decComment, nil
}
