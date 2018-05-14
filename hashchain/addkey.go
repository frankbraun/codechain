package hashchain

import (
	"fmt"
	"strconv"

	"github.com/frankbraun/codechain/hashchain/linktype"
	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/util/time"
	"golang.org/x/crypto/ed25519"
)

// AddKey adds pubkey with signature and optional comment to hash chain.
func (c *HashChain) AddKey(weight int, pubKey [32]byte, signature [64]byte, comment []byte) (string, error) {
	// check arguments
	if !ed25519.Verify(pubKey[:], append(pubKey[:], comment...), signature[:]) {
		return "", fmt.Errorf("signature does not verify")
	}

	// create entry
	typeFields := []string{
		strconv.Itoa(weight),
		base64.Encode(pubKey[:]),
		base64.Encode(signature[:]),
	}
	if len(comment) > 0 {
		typeFields = append(typeFields, string(comment))
	}
	l := &link{
		previous:   c.Head(),
		datum:      time.Now(),
		linkType:   linktype.AddKey,
		typeFields: typeFields,
	}
	c.chain = append(c.chain, l)

	// verify
	if err := c.verify(); err != nil {
		return "", err
	}

	// save
	if _, err := fmt.Fprintln(c.fp, l.String()); err != nil {
		return "", err
	}
	return l.StringColor(), nil
}
