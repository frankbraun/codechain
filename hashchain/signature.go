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
// If detached it just returns the signature without adding it.
func (c *HashChain) Signature(linkHash [32]byte, secKey [64]byte, detached bool) (string, error) {
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
		hex.Encode(linkHash[:]),
		base64.Encode(pub[:]),
		base64.Encode(sig),
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

	// detached signature?
	if detached {
		return fmt.Sprintf("%s %s %s", typeFields[0], typeFields[1], typeFields[2]), nil
	}

	// save
	if _, err := fmt.Fprintln(c.fp, l.String()); err != nil {
		return "", err
	}
	return l.StringColor(), nil
}

// DetachedSignature adds a detached signature entry for linkHash signed by
// pubKey to the hash chain.
func (c *HashChain) DetachedSignature(linkHash, pubKey, signature string) (string, error) {
	// decode arguments
	lh, err := hex.Decode(linkHash, 32)
	if err != nil {
		return "", err
	}
	pub, err := base64.Decode(pubKey, 32)
	if err != nil {
		return "", err
	}
	sig, err := base64.Decode(signature, 64)
	if err != nil {
		return "", err
	}

	// check arguments
	// make sure link hash does exist
	var h [32]byte
	copy(h[:], lh)
	if !c.state.HasLinkHash(h) {
		return "", fmt.Errorf("hashchain: link hash doesn't exist: %s", linkHash)
	}
	// make sure secKey is a valid signer
	var p [32]byte
	copy(p[:], pub)
	if !c.state.HasSigner(p) {
		return "", fmt.Errorf("hashchain: not a valid signer: %s", pubKey)
	}
	// TODO: similar checks as for Signature() refactor?

	// verify signature
	if !ed25519.Verify(pub, lh, sig) {
		return "", fmt.Errorf("signature does not verify")
	}

	// create entry
	typeFields := []string{linkHash, pubKey, signature}
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
	if _, err := fmt.Fprintln(c.fp, l.String()); err != nil {
		return "", err
	}
	return l.StringColor(), nil
}
