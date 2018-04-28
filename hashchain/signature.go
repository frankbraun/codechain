package hashchain

import (
	"fmt"

	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/util/time"
	"golang.org/x/crypto/ed25519"
)

// Signature adds a signature entry for entryHash signed by secKey to the hash chain.
func (c *HashChain) Signature(entryHash [32]byte, secKey [64]byte) (string, error) {
	// check arguments
	// TODO: check that entryHash is valid
	// TODO: make sure secKey is a valid signer
	// TODO: make sure entryHash is a valid position to sign

	// create signature
	pub := secKey[32:]
	sig := ed25519.Sign(secKey[:], entryHash[:])

	// create entry
	typeFields := []string{
		base64.Encode(entryHash[:]),
		base64.Encode(pub),
		base64.Encode(sig),
	}
	prev := c.LastEntryHash()
	l := &link{
		previous:   prev[:],
		datum:      time.Now(),
		linkType:   signatureType,
		typeFields: typeFields,
	}

	// verify
	if err := c.verify(); err != nil {
		return "", err
	}

	// save
	c.chain = append(c.chain, l)
	entry := l.String()
	if _, err := fmt.Fprintln(c.fp, entry); err != nil {
		return "", err
	}
	return entry, nil
}

// DetachedSignature adds a detached signature entry for entryHash signed by
// pubKey to the hash chain.
func (c *HashChain) DetachedSignature(entryHash, pubKey [32]byte, signature [64]byte) (string, error) {
	// check arguments
	// TODO: check that entryHash is valid
	// TODO: make sure secKey is a valid signer
	// TODO: make sure entryHash is a valid position to sign
	// Same checks as for Signature()

	// verify signature
	if !ed25519.Verify(pubKey[:], entryHash[:], signature[:]) {
		return "", fmt.Errorf("signature does not verify")
	}

	// create entry
	typeFields := []string{
		base64.Encode(entryHash[:]),
		base64.Encode(pubKey[:]),
		base64.Encode(signature[:]),
	}
	prev := c.LastEntryHash()
	l := &link{
		previous:   prev[:],
		datum:      time.Now(),
		linkType:   signatureType,
		typeFields: typeFields,
	}
	c.chain = append(c.chain, l)

	// verify
	if err := c.verify(); err != nil {
		return "", err
	}

	// save
	entry := l.String()
	if _, err := fmt.Fprintln(c.fp, entry); err != nil {
		return "", err
	}
	return entry, nil
}
