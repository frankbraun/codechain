package hashchain

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"

	"github.com/frankbraun/codechain/hashchain/internal/state"
	"github.com/frankbraun/codechain/hashchain/linktype"
	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/log"
	"golang.org/x/crypto/ed25519"
)

// hash-of-previous current-time cstart pubkey nonce signature [comment]
func (c *HashChain) verifyChainStartType(i int, fields []string) error {
	log.Printf("%d verify cstart", i)
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

	// start state
	c.state = state.New(pub, comment)
	return nil
}

// hash-of-previous current-time source source-hash pubkey signature [comment]
func (c *HashChain) verifySourceType(i int, fields []string) error {
	log.Printf("%d verify source", i)
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
		return ErrWrongSigSource
	}
	// make sure pubkey it is a valid signer
	var p [32]byte
	copy(p[:], pubKey)
	if !c.state.HasSigner(p) {
		return fmt.Errorf("hashchain: not a valid signer: %s", pub)
	}
	// make sure treehash has not been published before
	if err = c.state.NotPublished(tree); err != nil {
		return err
	}

	// update state
	var t [32]byte
	copy(t[:], treeHash)
	c.state.AddSourceHash(c.chain[i].Hash(), t, p, comment)
	return nil
}

// hash-of-previous current-time signtr hash-of-chain-entry pubkey signature
func (c *HashChain) verifySignatureType(i int, fields []string) error {
	log.Printf("%d verify signtr", i)
	// check arguments
	if i == 0 {
		return ErrMustStartWithCStart
	}
	if len(fields) != 3 {
		return ErrWrongTypeFields
	}

	// parse type fields
	link := fields[0]
	linkHash, err := hex.Decode(link, 32)
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

	// validate fields
	if !ed25519.Verify(pubKey, linkHash, sig) {
		return ErrWrongSigSignature
	}
	// make sure link hash does exist
	var l [32]byte
	copy(l[:], linkHash)
	if !c.state.HasLinkHash(l) {
		return fmt.Errorf("hashchain: link hash doesn't exist: %s", link)
	}

	// update state
	var p [32]byte
	copy(p[:], pubKey[:])
	return c.state.Sign(l, p)
}

// hash-of-previous current-time addkey pubkey-add w pubkey signature [comment]
func (c *HashChain) verifyAddKeyType(i int, fields []string) error {
	log.Printf("%d verify addkey", i)
	// check arguments
	if i == 0 {
		return ErrMustStartWithCStart
	}
	if len(fields) != 3 && len(fields) != 4 {
		return ErrWrongTypeFields
	}

	// parse type fields
	w := fields[0]
	weight, err := strconv.Atoi(w)
	if err != nil {
		return fmt.Errorf("hashchain: cannot parse weight: %s", w)
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
	var p [32]byte
	copy(p[:], pubKey)
	if !ed25519.Verify(p[:], append(pubKey, comment...), sig) {
		return ErrWrongSigAddKey
	}
	if err = c.state.NotSigner(p); err != nil {
		return err
	}

	// update state
	c.state.AddSigner(p, weight, comment)

	return nil
}

// hash-of-previous current-time remkey pubkey
func (c *HashChain) verifyRemoveKeyType(i int, fields []string) error {
	log.Printf("%d verify remkey", i)
	// check arguments
	if i == 0 {
		return ErrMustStartWithCStart
	}
	if len(fields) != 1 {
		return ErrWrongTypeFields
	}

	// parse type fields
	pub := fields[0]
	pubKey, err := base64.Decode(pub, 32)
	if err != nil {
		return err
	}

	// update state
	var p [32]byte
	copy(p[:], pubKey)
	return c.state.RemoveSigner(p)
}

// hash-of-previous current-time sigctl m
func (c *HashChain) verifySignatureControlType(i int, fields []string) error {
	log.Printf("%d verify sigctl", i)
	// check arguments
	if i == 0 {
		return ErrMustStartWithCStart
	}
	if len(fields) != 1 {
		return ErrWrongTypeFields
	}

	// parse type fields
	m, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("hashchain: cannot parse m: %d", m)
	}

	// validate fields
	if m <= 0 {
		return ErrSignatureThresholdNonPositive
	}
	if m > c.state.HeadN() {
		return ErrMLargerThanN
	}

	// update state
	c.state.SetSignatureControl(m)

	return nil
}

// verify hash chain.
func (c *HashChain) verify() error {
	// basic check
	if len(c.chain) == 0 {
		return ErrEmpty
	}

	// iterate over all links
	prevHash := emptyTree
	var prevDatum int64
	for i, l := range c.chain {
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
		case linktype.ChainStart:
			err = c.verifyChainStartType(i, l.typeFields)
		case linktype.Source:
			err = c.verifySourceType(i, l.typeFields)
		case linktype.Signature:
			err = c.verifySignatureType(i, l.typeFields)
		case linktype.AddKey:
			err = c.verifyAddKeyType(i, l.typeFields)
		case linktype.RemoveKey:
			err = c.verifyRemoveKeyType(i, l.typeFields)
		case linktype.SignatureControl:
			err = c.verifySignatureControlType(i, l.typeFields)
		default:
			err = ErrUnknownLinkType
		}
		if err != nil {
			return err
		}

		// store link hash and line number
		c.state.AddLinkHash(l.Hash(), i)
		if c.state.LinkHashes() != c.state.OPs() {
			return errors.New("c.state.LinkHashes() != c.state.OPs()") // should never happen
		}

		// prepare for next entry
		prevHash = sha256.Sum256([]byte(l.String()))
		prevDatum = l.datum
	}

	// all clear
	return nil
}
