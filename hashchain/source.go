package hashchain

import (
	"crypto/ed25519"
	"fmt"

	"github.com/frankbraun/codechain/hashchain/linktype"
	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/time"
)

// Source adds a source entry for treeHash and optional comment signed by
// secKey to the hash chain.
func (c *HashChain) Source(treeHash [32]byte, secKey [64]byte, comment []byte) (string, error) {
	// check arguments
	hash := hex.Encode(treeHash[:])
	if util.ContainsString(c.TreeHashes(), hash) {
		return "", fmt.Errorf("hashchain: treehash %s already published", hash)
	}
	pub := secKey[32:]
	pubKey := base64.Encode(pub)
	signer := c.Signer()
	if !signer[pubKey] {
		return "", fmt.Errorf("hashchain: pubkey %s is not an active signer", pubKey)
	}

	// create signature
	msg := treeHash[:]
	if len(comment) > 0 {
		msg = append(msg, comment...)
	}
	sig := ed25519.Sign(secKey[:], msg)

	// create entry
	typeFields := []string{
		hash,
		pubKey,
		base64.Encode(sig),
	}
	if len(comment) > 0 {
		typeFields = append(typeFields, string(comment))
	}
	l := &link{
		previous:   c.Head(),
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
	if _, err := fmt.Fprintln(c.fp, l.String()); err != nil {
		return "", err
	}
	return l.StringColor(), nil
}
