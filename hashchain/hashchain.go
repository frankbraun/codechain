package hashchain

import (
	"os"

	"github.com/frankbraun/codechain/hashchain/internal/state"
	"github.com/frankbraun/codechain/util/lockfile"
)

// HashChain of threshold signatures over a chain of code changes.
type HashChain struct {
	lock  lockfile.Lock
	fp    *os.File
	chain []*link
	state *state.State
}

// LastEntryHash returns the hash of the last entry.
func (c *HashChain) LastEntryHash() [32]byte {
	return c.chain[len(c.chain)-1].Hash()
}

// LastTreeHash returns the most current tree hash (can be unsigned).
func (c *HashChain) LastTreeHash() string {
	return c.state.LastTreeHash()
}

// LastSignedTreeHash returns the last signed tree hash.
func (c *HashChain) LastSignedTreeHash() string {
	return c.state.LastSignedTreeHash()
}

// TreeHashes returns a list of all tree hashes in order (starting from
// tree.EmptyHash).
func (c *HashChain) TreeHashes() []string {
	return c.state.TreeHashes()
}

// Signer returns a map containing all active signers for hash chain.
func (c *HashChain) Signer() map[string]bool {
	return c.state.Signer()
}

// EntryHash returns the entry hash for the given treeHash.
func (c *HashChain) EntryHash(treeHash [32]byte) [32]byte {
	var h [32]byte
	// TODO: implement
	return h
}

// Close the underlying file pointer of hash chain and release lock.
func (c *HashChain) Close() error {
	if c.fp == nil {
		return c.lock.Release()
	}
	err := c.fp.Close()
	if err != nil {
		c.lock.Release()
		return err
	}
	c.fp = nil
	return c.lock.Release()
}
