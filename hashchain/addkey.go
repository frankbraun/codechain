package hashchain

import (
	"fmt"

	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/util/time"
	"golang.org/x/crypto/ed25519"
)

// AddKey adds pubkey with signature and optional comment to hash chain.
func (c *HashChain) AddKey(pubKey [32]byte, signature [64]byte, comment []byte) (string, error) {
	// check arguments
	if !ed25519.Verify(pubKey[:], append(pubKey[:], comment...), signature[:]) {
		return "", fmt.Errorf("signature does not verify")
	}

	// create entry
	typeFields := []string{
		base64.Encode(pubKey[:]),
		base64.Encode(signature[:]),
	}
	if len(comment) > 0 {
		typeFields = append(typeFields, string(comment))
	}
	prev := c.LastEntryHash()
	l := &link{
		previous:   prev[:],
		datum:      time.Now(),
		linkType:   addKeyType,
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
