package hashchain

import (
	"fmt"

	"github.com/frankbraun/codechain/hashchain/linktype"
	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/util/time"
	"golang.org/x/crypto/ed25519"
)

// Source adds a source entry for treeHash and optional comment signed by
// secKey to the hash chain.
func (c *HashChain) Source(treeHash [32]byte, secKey [64]byte, comment []byte) (string, error) {
	// check arguments
	// TODO: treeHash
	// TODO: secKey

	// create signature
	pub := secKey[32:]
	msg := treeHash[:]
	if len(comment) > 0 {
		msg = append(msg, comment...)
	}
	sig := ed25519.Sign(secKey[:], msg)

	// create entry
	typeFields := []string{
		base64.Encode(treeHash[:]),
		base64.Encode(pub),
		base64.Encode(sig),
	}
	if len(comment) > 0 {
		typeFields = append(typeFields, string(comment))
	}
	l := &link{
		previous:   c.LastEntryHash(),
		datum:      time.Now(),
		linkType:   linktype.Source,
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
