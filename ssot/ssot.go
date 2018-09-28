// Package ssot implements a single source of truth (SSOT)
// with DNS TXT records.
package ssot

// SignedHead is a signed Codechain head ready for publication as a SSOT with
// DNS TXT records.
type SignedHead struct {
	pubkey       [32]byte // Ed25519 public key of SSOT head signer
	pubkeyRotate [32]byte // Ed25519 pubkey to rotate to, all 0 if unused
	validFrom    int64    // this signed head is valid from the given Unix time
	validTo      int64    // this signed head is valid to the given Unix time
	counter      int64    // signature counter
	head         [32]byte // the Codechain head to sign
	signature    [64]byte // signature with pubkey over all previous fields
}

// SignHead signs the given Codechain head.
func SignHead(head [32]byte) (*SignedHead, error) {
	// TODO

	return nil, nil
}
