package archive

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
	"golang.org/x/crypto/nacl/secretbox"
)

// CreateDist creates a distribution file with filename for hash chain c.
// Filename must not exist.
func CreateDist(c *hashchain.HashChain, filename string) error {
	exists, err := file.Exists(filename)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("distribution file '%s' exists already", filename)
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("creating distribution '%s'", filename)
	return Create(f, c, def.PatchDir)
}

// CreateEncryptedDist creates an encrypted distribution file with filename
// for hash chain c. Filename must not exists.
func CreateEncryptedDist(c *hashchain.HashChain, filename string, key *[32]byte) error {
	exists, err := file.Exists(filename)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("distribution file '%s' exists already", filename)
	}
	var b bytes.Buffer
	if err := Create(&b, c, def.PatchDir); err != nil {
		return err
	}
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}
	enc := secretbox.Seal(nonce[:], b.Bytes(), &nonce, key)
	log.Printf("creating encrypted distribution '%s'", filename)
	return ioutil.WriteFile(filename, enc, 0666)
}
