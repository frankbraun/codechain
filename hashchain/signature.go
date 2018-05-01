package hashchain

import (
	"fmt"

	"github.com/frankbraun/codechain/hashchain/linktype"
	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/internal/hex"
	"github.com/frankbraun/codechain/util/time"
	"golang.org/x/crypto/ed25519"
)

// Signature adds a signature entry for entryHash signed by secKey to the hash chain.
func (c *HashChain) Signature(linkHash [32]byte, secKey [64]byte) (string, error) {
	// check arguments
	// make sure link hash does exist
	if !c.state.HasLinkHash(linkHash) {
		return "", fmt.Errorf("hashchain: link hash doesn't exist: %s",
			hex.Encode(linkHash[:]))
	}
	// make sure secKey is a valid signer
	var pub [32]byte
	copy(pub[:], secKey[32:])
	if !c.state.HasSigner(pub) {
		return "", fmt.Errorf("hashchain: not a valid signer: %s",
			hex.Encode(pub[:]))
	}
	// TODO: make sure entryHash is a valid position to sign

	// create signature
	sig := ed25519.Sign(secKey[:], linkHash[:])

	// create entry
	typeFields := []string{
		base64.Encode(linkHash[:]),
		base64.Encode(pub[:]),
		base64.Encode(sig),
	}
	l := &link{
		previous:   c.LastEntryHash(),
		datum:      time.Now(),
		linkType:   linktype.Signature,
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

// DetachedSignature adds a detached signature entry for linkHash signed by
// pubKey to the hash chain.
func (c *HashChain) DetachedSignature(linkHash, pubKey [32]byte, signature [64]byte) (string, error) {
	// check arguments
	// make sure link hash does exist
	if !c.state.HasLinkHash(linkHash) {
		return "", fmt.Errorf("hashchain: link hash doesn't exist: %s",
			hex.Encode(linkHash[:]))
	}
	// make sure secKey is a valid signer
	if !c.state.HasSigner(pubKey) {
		return "", fmt.Errorf("hashchain: not a valid signer: %s",
			hex.Encode(pubKey[:]))
	}
	// TODO: similar checks as for Signature() refactor?

	// verify signature
	if !ed25519.Verify(pubKey[:], linkHash[:], signature[:]) {
		return "", fmt.Errorf("signature does not verify")
	}

	// create entry
	typeFields := []string{
		base64.Encode(linkHash[:]),
		base64.Encode(pubKey[:]),
		base64.Encode(signature[:]),
	}
	l := &link{
		previous:   c.LastEntryHash(),
		datum:      time.Now(),
		linkType:   linktype.Signature,
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
