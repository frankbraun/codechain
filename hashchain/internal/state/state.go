// Package state implements the state of a hashchain.
package state

import (
	"errors"
	"fmt"

	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/log"
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

// M returns the signature threshold.
func (s *State) M() int {
	return s.m
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

// SourceLine returns the line number where the given tree hash was signed.
func (s *State) SourceLine(treeHash string) int {
	return s.linkHashes[s.treeHashes[treeHash]]
}

// LinkHash returns the link hash corresponding to given treeHash.
func (s *State) LinkHash(treeHash string) [32]byte {
	linkHash, err := hex.Decode(s.treeHashes[treeHash], 32)
	if err != nil {
		panic(err)
	}
	var lh [32]byte
	copy(lh[:], linkHash)
	return lh
}

// HasSigner checks wether the state s contains a valid the signer with
// pubKey.
func (s *State) HasSigner(pubKey [32]byte) bool {
	_, ok := s.signerWeights[base64.Encode(pubKey[:])]
	return ok
}

// NotSigner makes sure the given pubKey is not a signer (unconfirmed or
// confirmed).
func (s *State) NotSigner(pubKey [32]byte) error {
	pub := base64.Encode(pubKey[:])
	for i := len(s.unconfirmedOPs) - 1; i > s.signedLine; i-- {
		switch op := s.unconfirmedOPs[i].(type) {
		case *addKeyOP:
			if op.pubKey == pub {
				return errors.New("state: duplicate addkey (unsigned)")
			}
		case *remKeyOP:
			if op.pubKey == pub {
				return nil
			}
		}
	}
	_, ok := s.signerWeights[pub]
	if ok {
		return errors.New("state: duplicate addkey (signed)")
	}
	return nil
}

// NotPublished makes sure that the given treeHash has not been published
// before (unconfirmed or confirmed).
func (s *State) NotPublished(treeHash string) error {
	if util.ContainsString(s.signedTreeHashes, treeHash) {
		return errors.New("state: duplicate treehash (signed)")
	}
	for i := s.signedLine + 1; i < len(s.unconfirmedOPs); i++ {
		switch op := s.unconfirmedOPs[i].(type) {
		case *sourceOP:
			if op.treeHash == treeHash {
				return errors.New("state: duplicate treehash (unsigned)")
			}
		}
	}
	return nil
}

// LastTreeHash returns the most current tree hash.
func (s *State) LastTreeHash() string {
	for i := len(s.unconfirmedOPs) - 1; i > s.signedLine; i-- {
		op, ok := s.unconfirmedOPs[i].(*sourceOP)
		if ok {
			return op.treeHash
		}
	}
	return s.signedTreeHashes[len(s.signedTreeHashes)-1]
}

// LastSignedTreeHash returns the last signed tree hash.
func (s *State) LastSignedTreeHash() (string, int) {
	idx := len(s.signedTreeHashes) - 1
	return s.signedTreeHashes[idx], idx
}

// TreeHashes returns a list of all tree hashes in order (starting from
// tree.EmptyHash).
func (s *State) TreeHashes() []string {
	treeHashes := append([]string{}, s.signedTreeHashes...)
	for i := s.signedLine + 1; i < len(s.unconfirmedOPs); i++ {
		op, ok := s.unconfirmedOPs[i].(*sourceOP)
		if ok {
			treeHashes = append(treeHashes, op.treeHash)
		}
	}
	return treeHashes
}

// TreeComments returns a list of all tree comments in order (starting from
// tree.EmptyHash).
func (s *State) TreeComments() []string {
	treeComments := append([]string{}, s.signedTreeComments...)
	for i := s.signedLine + 1; i < len(s.unconfirmedOPs); i++ {
		op, ok := s.unconfirmedOPs[i].(*sourceOP)
		if ok {
			treeComments = append(treeComments, op.comment)
		}
	}
	return treeComments
}

// lastWeight returns the last weight added for given pubKey (unconfirmed or
// confirmed).
func (s *State) lastWeight(pubKey [32]byte) (int, error) {
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
		return 0, errors.New("state: unknown pubkey")
	}
	return w, nil
}

// Signer returns a containing all active signers for state.
func (s *State) Signer() map[string]bool {
	signer := make(map[string]bool)
	for s := range s.signerWeights {
		signer[s] = true
	}
	return signer
}

// SignerComment returns the signer comment for given pubKey.
func (s *State) SignerComment(pubKey string) string {
	return s.signerComments[pubKey]
}

// SignerWeight returns the signer weight for given pubKey.
func (s *State) SignerWeight(pubKey string) int {
	return s.signerWeights[pubKey]
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
func (s *State) RemoveSigner(pubKey [32]byte) error {
	pub := base64.Encode(pubKey[:])
	w, err := s.lastWeight(pubKey)
	if err != nil {
		return err
	}
	op := newRemKeyOP(pub, w)

	m := s.m
	n := s.n
	for i := s.signedLine + 1; i < len(s.unconfirmedOPs); i++ {
		switch op := s.unconfirmedOPs[i].(type) {
		case *noOP:
			continue
		case *sourceOP:
			continue
		case *addKeyOP:
			n += op.weight
		case *remKeyOP:
			n -= op.weight
		case *sigCtlOp:
			m = op.m
		default:
			return errors.New("state: RemoveSigner(): unknown OP type")
		}
	}
	if n-w < m {
		return errors.New("remkey would lead to n < m, lower sigctl first")
	}
	s.unconfirmedOPs = append(s.unconfirmedOPs, op)
	return nil

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
	log.Printf("state.Sign(): pubKey=%s", pub)
	log.Printf("state.Sign(): line=%d", line)
	log.Printf("state.Sign(): signerBarrier=%d", s.signerBarriers[pub])
	// sign lines not signed by this signer yet
	for i := s.signerBarriers[pub] + 1; i <= line; i++ {
		s.unconfirmedOPs[i].sign(weight)
	}
	s.signerBarriers[pub] = line
	log.Printf("state.Sign(): signerBarrier=%d", s.signerBarriers[pub])
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
	s.signedLine = i - 1
	s.unconfirmedOPs = append(s.unconfirmedOPs, nop)
	return nil
}

// UnsignedInfo returns a string slice with information about all unsigned
// entries suitable for printing.
// If TreeHash is defined it returns info until that treeHash.
// If omitSource is true source lines are omitted
func (s *State) UnsignedInfo(pubKey, treeHash string, omitSource bool) ([]string, error) {
	var infos []string
	end := len(s.unconfirmedOPs)
	if treeHash != "" {
		end = s.SourceLine(treeHash)
	}
	i := s.signedLine + 1
	if pubKey != "" {
		i = s.signerBarriers[pubKey] + 1
	}
	for ; i < end; i++ {
		switch op := s.unconfirmedOPs[i].(type) {
		case *noOP:
			continue
		case *sourceOP:
			if omitSource {
				continue
			}
			info := fmt.Sprintf("%d source %s %s", op.signatures(), op.treeHash, op.comment)
			infos = append(infos, info)
		case *addKeyOP:
			info := fmt.Sprintf("%d addkey %d %s %s", op.signatures(), op.weight, op.pubKey, op.comment)
			infos = append(infos, info)
		case *remKeyOP:
			info := fmt.Sprintf("%d remkey %d %s %s", op.signatures(), op.weight, op.pubKey,
				s.signerComments[op.pubKey]) // shows only comments from already confirmed signers, but that's fine
			infos = append(infos, info)
		case *sigCtlOp:
			info := fmt.Sprintf("%d sigctl %d", op.signatures(), op.m)
			infos = append(infos, info)
		default:
			return nil, errors.New("state: Sign(): unknown OP type")
		}
	}
	return infos, nil
}

// SignerBarrier returns the signer barrier for pubKey.
func (s *State) SignerBarrier(pubKey string) int {
	return s.signerBarriers[pubKey]
}
