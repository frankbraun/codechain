package hashchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/internal/hex"
	"github.com/frankbraun/codechain/tree"
	"golang.org/x/crypto/ed25519"
)

// hash-of-previous current-time cstart pubkey nonce signature [comment]
func (c *HashChain) verifyChainStartType(i int, fields []string) error {
	// check arguments
	if i != 0 {
		return ErrIllegalCStart
	}
	if len(fields) != 3 && len(fields) != 4 {
		return ErrWrongTypeFields
	}

	// parse type fields
	pub := fields[0]
	pubKey, err := base64.Decode(pub, 32)
	if err != nil {
		return err
	}
	nonce, err := base64.Decode(fields[1], 24)
	if err != nil {
		return err
	}
	sig, err := base64.Decode(fields[2], 64)
	if err != nil {
		return err
	}
	var comment string
	if len(fields) == 4 {
		comment = fields[3]
	}

	// validate fields
	msg := append(pubKey, nonce...)
	msg = append(msg, comment...)
	if !ed25519.Verify(pubKey, msg, sig) {
		return ErrWrongSigCStart
	}

	// update state
	c.signerWeights[pub] = 1 // default weight for first signer
	c.signerComments[pub] = comment
	c.signerComments[pub] = ""
	return nil
}

// hash-of-previous current-time source source-hash pubkey signature [comment]
func (c *HashChain) verifySourceType(i int, fields []string) error {
	// check arguments
	if i == 0 {
		return ErrMustStartWithCStart
	}
	if len(fields) != 3 && len(fields) != 4 {
		return ErrWrongTypeFields
	}

	// parse type fields
	tree := fields[0]
	treeHash, err := hex.Decode(tree, 32)
	if err != nil {
		return err
	}
	pub := fields[1]
	pubKey, err := base64.Decode(pub, 32)
	if err != nil {
		return err
	}
	sig, err := base64.Decode(fields[2], 64)
	if err != nil {
		return err
	}
	var comment string
	if len(fields) == 4 {
		comment = fields[3]
	}

	// validate fields
	msg := append(treeHash, comment...)
	if !ed25519.Verify(pubKey, msg, sig) {
		return ErrWrongSigCStart
	}
	// make sure pubkey it is a valid signer
	if c.signerWeights[pub] <= 0 {
		return fmt.Errorf("hashchain: %s is not a valid signer", pub)
	}

	// update state
	c.treeHashes = append(c.treeHashes, tree)
	return nil
}

// hash-of-previous current-time signtr hash-of-chain-entry pubkey signature
func (c *HashChain) verifySignatureType(i int, fields []string) error {
	// check arguments
	if i == 0 {
		return ErrMustStartWithCStart
	}
	if len(fields) != 3 {
		return ErrWrongTypeFields
	}

	// parse type fields
	// TODO
	return nil
}

// hash-of-previous current-time addkey pubkey-add w pubkey signature [comment]
func (c *HashChain) verifyAddKeyType(i int, fields []string) error {
	// check arguments
	if i == 0 {
		return ErrMustStartWithCStart
	}
	if len(fields) != 3 && len(fields) != 4 {
		return ErrWrongTypeFields
	}

	// parse type fields
	// TODO
	return nil
}

// hash-of-previous current-time remkey pubkey
func (c *HashChain) verifyRemoveKeyType(i int, fields []string) error {
	// check arguments
	if i == 0 {
		return ErrMustStartWithCStart
	}
	if len(fields) != 2 {
		return ErrWrongTypeFields
	}

	// parse type fields
	// TODO
	return nil
}

// hash-of-previous current-time sigctl m
func (c *HashChain) verifySignatureControlType(i int, fields []string) error {
	// check arguments
	if i == 0 {
		return ErrMustStartWithCStart
	}
	if len(fields) != 2 {
		return ErrWrongTypeFields
	}

	// parse type fields
	// TODO
	return nil
}

// verify hash chain.
func (c *HashChain) verify() error {
	// basic check
	if len(c.chain) == 0 {
		return ErrEmpty
	}

	// set start state
	c.m = 1
	c.n = 1
	c.signerWeights = make(map[string]int)
	c.signerComments = make(map[string]string)
	c.entryHashes = make(map[string]int)
	c.treeHashes = []string{tree.EmptyHash}

	// iterate over all links
	prevHash := emptyTree
	var prevDatum int64
	for i, l := range c.chain {
		// store entry hash
		h := l.Hash()
		c.entryHashes[hex.Encode(h[:])] = i

		// make sure we actually have a hash chain
		if !bytes.Equal(prevHash[:], l.previous[:]) {
			return ErrLinkBroken
		}

		// make sure time is ascending
		if l.datum < prevDatum {
			return ErrDescendingTime
		}

		var err error
		switch l.linkType {
		case chainStartType:
			err = c.verifyChainStartType(i, l.typeFields)
		case sourceType:
			err = c.verifySourceType(i, l.typeFields)
		case signatureType:
			err = c.verifySignatureType(i, l.typeFields)
		case addKeyType:
			err = c.verifyAddKeyType(i, l.typeFields)
		case removeKeyType:
			err = c.verifyRemoveKeyType(i, l.typeFields)
		case signatureControlType:
			err = c.verifySignatureControlType(i, l.typeFields)
		default:
			err = ErrUnknownLinkType
		}
		if err != nil {
			return err
		}

		// prepare for next entry
		prevHash = sha256.Sum256([]byte(l.String()))
		prevDatum = l.datum
	}

	// all clear
	return nil
}
