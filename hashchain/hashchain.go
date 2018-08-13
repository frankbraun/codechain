package hashchain

import (
	"bytes"
	"fmt"
	"io"
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

// M returns the signature threshold.
func (c *HashChain) M() int {
	return c.state.M()
}

// N returns the total weight of all signers.
func (c *HashChain) N() int {
	return c.state.N()
}

// Head returns the hash of the last entry.
func (c *HashChain) Head() [32]byte {
	return c.chain[len(c.chain)-1].Hash()
}

// CheckHead checks wether the hash chain contains the given head as entry.
func (c *HashChain) CheckHead(head [32]byte) error {
	for _, l := range c.chain {
		h := l.Hash()
		if bytes.Equal(h[:], head[:]) {
			return nil
		}
	}
	return ErrHeadNotFound
}

// LastTreeHash returns the most current tree hash (can be unsigned).
func (c *HashChain) LastTreeHash() string {
	return c.state.LastTreeHash()
}

// LastSignedTreeHash returns the last signed tree hash and its index.
// The first signed tree hash is tree.EmptyHash with index 0.
func (c *HashChain) LastSignedTreeHash() (string, int) {
	return c.state.LastSignedTreeHash()
}

// TreeHashes returns a list of all tree hashes in order (starting from
// tree.EmptyHash).
func (c *HashChain) TreeHashes() []string {
	return c.state.TreeHashes()
}

// TreeComments returns a list of all tree comments in order (starting from
// tree.EmptyHash).
func (c *HashChain) TreeComments() []string {
	return c.state.TreeComments()
}

// Signer returns a map containing all active signers for hash chain.
func (c *HashChain) Signer() map[string]bool {
	return c.state.Signer()
}

// SignerComment returns the signer comment for given pubKey.
func (c *HashChain) SignerComment(pubKey string) string {
	return c.state.SignerComment(pubKey)
}

// SignerWeight returns the signer weight for given pubKey.
func (c *HashChain) SignerWeight(pubKey string) int {
	return c.state.SignerWeight(pubKey)
}

// SignerInfo returns signer pubKey and comment for patch with given treeHash.
func (c *HashChain) SignerInfo(treeHash string) (string, string) {
	link := c.chain[c.state.SourceLine(treeHash)]
	pubKey := link.typeFields[1]
	return pubKey, c.state.SignerComment(pubKey)
}

// LinkHash returns the link hash corresponding to given treeHash.
func (c *HashChain) LinkHash(treeHash string) [32]byte {
	return c.state.LinkHash(treeHash)
}

// SourceLine returns the line number where the given tree hash was signed.
func (c *HashChain) SourceLine(treeHash string) int {
	return c.state.SourceLine(treeHash)
}

// UnsignedInfo returns a string slice with information about all unsigned
// entries suitable for printing.
// If TreeHash is defined it returns info until that treeHash.
// If omitSource is true source lines are omitted
func (c *HashChain) UnsignedInfo(pubkey, treeHash string, omitSource bool) ([]string, error) {
	return c.state.UnsignedInfo(pubkey, treeHash, omitSource)
}

// SignerBarrier returns the signer barrier for pubKey.
func (c *HashChain) SignerBarrier(pubKey string) int {
	return c.state.SignerBarrier(pubKey)
}

// Print colorized hash chain on stdout.
func (c *HashChain) Print() {
	for _, l := range c.chain {
		fmt.Println(l.StringColor())
	}
}

// Fprint hash chain to w.
func (c *HashChain) Fprint(w io.Writer) error {
	for _, l := range c.chain {
		if _, err := fmt.Fprintln(w, l.String()); err != nil {
			return err
		}
	}
	return nil
}
