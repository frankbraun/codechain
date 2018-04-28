package hashchain

import (
	"os"

	"github.com/frankbraun/codechain/util/lockfile"
)

// HashChain of threshold signatures over a chain of code changes.
type HashChain struct {
	// chain
	lock  lockfile.Lock
	fp    *os.File
	chain []*link

	// state
	m                  int               // signature threshold
	n                  int               // total weight of signers
	signedLine         int               // line up to and including every entry is signed
	signerWeights      map[string]int    // pubkey (in base64) -> weight
	signerComments     map[string]string // pubkey (in base64) -> comment
	signerBarriers     map[string]int    // pubkey (in base64) -> line number up to he signed
	linkHashes         map[string]int    // link hash -> line number
	treeHashes         map[string]string // tree hash -> link hash
	signedTreeHashes   []string          // all signed tree hashes, starting from empty tree
	unsignedTreeHashes []string          // all unsigned tree hashes
}

// LastEntryHash returns the hash of the last entry.
func (c *HashChain) LastEntryHash() [32]byte {
	return c.chain[len(c.chain)-1].Hash()
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
