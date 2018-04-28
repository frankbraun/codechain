package hashchain

import (
	"crypto/sha256"
	"os"

	"github.com/frankbraun/codechain/util/lockfile"
)

// HashChain of threshold signatures over a chain of code changes.
type HashChain struct {
	lock           lockfile.Lock
	fp             *os.File
	chain          []*link
	m              int               // signature threshold
	n              int               // total weight of signers
	signerWeights  map[string]int    // map signer pubkeys (in base64) to their weights
	signerComments map[string]string // map signer pubkeys (in base64) to their comments
}

// LastEntryHash returns the hash of the last entry.
func (c *HashChain) LastEntryHash() [32]byte {
	return sha256.Sum256([]byte(c.chain[len(c.chain)-1].String()))
}

// EntryHash returns the entry hash for the given treeHash.
func (c *HashChain) EntryHash(treeHash [32]byte) [32]byte {
	var h [32]byte
	// TODO: implement
	return h
}

// Close the underlying file pointer of hash chain and release lock.
func (c *HashChain) Close() error {
	err := c.fp.Close()
	if err != nil {
		c.lock.Release()
		return err
	}
	return c.lock.Release()
}
