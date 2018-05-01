// Package state implements the state of a hashchain.
package state

import (
	"errors"

	"github.com/frankbraun/codechain/internal/base64"
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
	signedTreeComments []string          // all signed tree comments
	unconfirmedOPs     []op              // unconfirmed operations
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
		signedTreeComments: []string{""},
		unconfirmedOPs:     []op{nop},
	}
	s.signerWeights[pubKey] = 1 // default weight for first signer
	s.signerComments[pubKey] = comment
	s.signerBarriers[pubKey] = 0
	return s
}

// N returns the total weight of all signers.
func (s *State) N() int {
	return s.n
}

// HeadN returns the total weight of all signers, including unconfirmed
// key additions and removals.
func (s *State) HeadN() int {
	n := s.n
	for i := s.signedLine + 1; i < len(s.unconfirmedOPs); i++ {
		switch op := s.unconfirmedOPs[i].(type) {
		case *addKeyOP:
			n += op.weight
		case *remKeyOP:
			n -= op.weight
		}
	}
	return n
}

// OPs returns the number of operations in the hash chain.
func (s *State) OPs() int {
	return len(s.unconfirmedOPs)
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

// LinkHashes returns the number of link hashes contained in state.
func (s *State) LinkHashes() int {
	return len(s.linkHashes)
}

// HasSigner checks wether the state s contains a valid the signer with
// pubKey.
func (s *State) HasSigner(pubKey [32]byte) bool {
	_, ok := s.signerWeights[base64.Encode(pubKey[:])]
	return ok
}

// LastWeight returns the last weight added for given pubKey (unconfirmed or
// confirmed).
func (s *State) LastWeight(pubKey [32]byte) (int, error) {
	pub := base64.Encode(pubKey[:])
	for i := len(s.unconfirmedOPs) - 1; i > s.signedLine; i-- {
		switch op := s.unconfirmedOPs[i].(type) {
		case *addKeyOP:
			if op.pubKey == pub {
				return op.weight, nil
			}
		case *remKeyOP:
			if op.pubKey == pub {
				return 0, errors.New("state: duplicate remkey")
			}
		}
	}
	w, ok := s.signerWeights[pub]
	if !ok {
		return 0, errors.New("state: unknown remkey")
	}
	return w, nil
}

// AddSourceHash adds treeHash at given linkHash to state.
func (s *State) AddSourceHash(linkHash, treeHash, pubKey [32]byte, comment string) {
	link := hex.Encode(linkHash[:])
	tree := hex.Encode(treeHash[:])
	pub := base64.Encode(pubKey[:])
	s.treeHashes[tree] = link
	op := newSourceOP(tree, pub, comment)
	s.unconfirmedOPs = append(s.unconfirmedOPs, op)
}

// AddSigner adds pubKey with weight to state (unconfirmed).
func (s *State) AddSigner(pubKey [32]byte, weight int, comment string) {
	pub := base64.Encode(pubKey[:])
	op := newAddKeyOP(pub, weight, comment)
	s.unconfirmedOPs = append(s.unconfirmedOPs, op)
}

// RemoveSigner removes pubKey with weight (must equal last addition) from
// state (unconfirmed).
func (s *State) RemoveSigner(pubKey [32]byte, weight int) {
	pub := base64.Encode(pubKey[:])
	op := newRemKeyOP(pub, weight)
	s.unconfirmedOPs = append(s.unconfirmedOPs, op)
}

// SetSignatureControl sets new signature control m (unconfirmed).
func (s *State) SetSignatureControl(m int) {
	op := newSigCtlOp(m)
	s.unconfirmedOPs = append(s.unconfirmedOPs, op)
}

// Sign signs the given linkHash with pubKey.
func (s *State) Sign(linkHash, pubKey [32]byte) error {
	link := hex.Encode(linkHash[:])
	pub := base64.Encode(pubKey[:])
	line, ok := s.linkHashes[link]
	if !ok {
		return errors.New("state: Sign(): unknown linkHash")
	}
	weight, ok := s.signerWeights[pub]
	if !ok {
		return errors.New("state: Sign(): unknown pubKey")
	}
	// sign lines
	for i := s.signedLine + 1; i <= line; i++ {
		s.unconfirmedOPs[i].sign(weight)
	}
	// check if we can commit stuff
	var i int
	for i = s.signedLine + 1; i <= line; i++ {
		if s.unconfirmedOPs[i].signatures() >= s.m {
			switch op := s.unconfirmedOPs[i].(type) {
			case *noOP:
				continue
			case *sourceOP:
				_, ok := s.signerWeights[op.pubKey]
				if !ok {
					return errors.New("state: Sign(): unknown source pubKey")
				}
				s.signedTreeHashes = append(s.signedTreeHashes, op.treeHash)
				s.signedTreeComments = append(s.signedTreeComments, op.comment)
			case *addKeyOP:
				s.n += op.weight
				s.signerWeights[op.pubKey] = op.weight
				s.signerComments[op.pubKey] = op.comment
				s.signerBarriers[op.pubKey] = i
			case *remKeyOP:
				s.n -= op.weight
				delete(s.signerWeights, op.pubKey)
				delete(s.signerComments, op.pubKey)
				delete(s.signerBarriers, op.pubKey)
			case *sigCtlOp:
				s.m = op.m
			default:
				return errors.New("state: Sign(): unknown OP type")
			}
			s.unconfirmedOPs[i] = nop
		} else {
			// nothing to commit
			break
		}
	}
	s.signerBarriers[pub] = i - 1
	s.unconfirmedOPs = append(s.unconfirmedOPs, nop)
	return nil
}
