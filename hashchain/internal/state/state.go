// Package state implements the state of a hashchain.
package state

import (
	"github.com/frankbraun/codechain/internal/hex"
	"github.com/frankbraun/codechain/tree"
)

// State hold the state of a hashchain.
type State struct {
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

// New returns a new state for pubKey with optional comment.
func New(pubKey, comment string) *State {
	s := &State{
		m:                  1,
		n:                  1,
		signedLine:         0,
		signerWeights:      make(map[string]int),
		signerComments:     make(map[string]string),
		signerBarriers:     make(map[string]int),
		linkHashes:         make(map[string]int),
		treeHashes:         make(map[string]string),
		signedTreeHashes:   []string{tree.EmptyHash},
		unsignedTreeHashes: []string{},
	}
	s.signerWeights[pubKey] = 1 // default weight for first signer
	s.signerComments[pubKey] = comment
	return s
}

// N returns the total weight of all signers.
func (s *State) N() int {
	return s.n
}

// AddLinkHash adds linkHash with lineNumber to state.
func (s *State) AddLinkHash(linkHash [32]byte, lineNumber int) {
	s.linkHashes[hex.Encode(linkHash[:])] = lineNumber
}

// HasLinkHash checks wether the state s contains the given linkHash.
func (s *State) HasLinkHash(linkHash [32]byte) bool {
	_, ok := s.linkHashes[hex.Encode(linkHash[:])]
	return ok
}

// HasSigner checks wether the state s contains a valid the signer with
// pubKey.
func (s *State) HasSigner(pubKey [32]byte) bool {
	_, ok := s.signerWeights[hex.Encode(pubKey[:])]
	return ok
}

// AddTreeHash adds treeHash at given linkHash to state.
func (s *State) AddTreeHash(linkHash, treeHash [32]byte) {
	link := hex.Encode(linkHash[:])
	tree := hex.Encode(treeHash[:])
	s.treeHashes[tree] = link
	s.unsignedTreeHashes = append(s.unsignedTreeHashes, tree)
}
