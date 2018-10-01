package ssot

import (
	"encoding/binary"
	"fmt"

	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/internal/hex"
	"golang.org/x/crypto/ed25519"
)

// File defines the default file name for a signed head.
const File = "signed_head"

// MaximumValidity of signed heads.
const MaximumValidity = 30 * 24 * 60 * 60 // 30d

// TTL of signed head TXT records
const TTL = 600 // 10m

// SignedHead is a signed Codechain head ready for publication as a SSOT with
// DNS TXT records.
type SignedHead struct {
	pubKey       [32]byte // Ed25519 public key of SSOT head signer
	pubKeyRotate [32]byte // Ed25519 pubkey to rotate to, all 0 if unused
	validFrom    int64    // this signed head is valid from the given Unix time
	validTo      int64    // this signed head is valid to the given Unix time
	counter      uint64   // signature counter
	head         [32]byte // the Codechain head to sign
	signature    [64]byte // signature with pubkey over all previous fields
}

// marshal signed head without signature.
func (sh *SignedHead) marshal() [120]byte {
	var m [120]byte
	var b [8]byte
	copy(m[:32], sh.pubKey[:])
	copy(m[32:64], sh.pubKeyRotate[:])
	binary.BigEndian.PutUint64(b[:], uint64(sh.validFrom))
	copy(m[64:72], b[:])
	binary.BigEndian.PutUint64(b[:], uint64(sh.validTo))
	copy(m[72:80], b[:])
	binary.BigEndian.PutUint64(b[:], sh.counter)
	copy(m[80:88], b[:])
	copy(m[88:120], sh.head[:])
	return m
}

// Marshal signed head with signature and encode it as base64.
func (sh *SignedHead) Marshal() string {
	var m [184]byte
	b := sh.marshal()
	copy(m[:120], b[:])
	copy(m[120:184], sh.signature[:])
	return base64.Encode(m[:])
}

func unmarshal(m [184]byte) (*SignedHead, error) {
	var sh SignedHead
	copy(sh.pubKey[:], m[:32])
	copy(sh.pubKeyRotate[:], m[32:64])
	sh.validFrom = int64(binary.BigEndian.Uint64(m[64:72]))
	sh.validTo = int64(binary.BigEndian.Uint64(m[72:80]))
	sh.counter = binary.BigEndian.Uint64(m[80:88])
	copy(sh.head[:], m[88:120])
	copy(sh.signature[:], m[120:184])
	msg := sh.marshal()
	if !ed25519.Verify(sh.pubKey[:], msg[:], sh.signature[:]) {
		return nil, ErrSignedHeadSignature
	}
	return &sh, nil
}

// Unmarshal and verify a base64 encoded signed head.
func Unmarshal(signedHead string) (*SignedHead, error) {
	b, err := base64.Decode(signedHead, 184)
	if err != nil {
		return nil, err
	}
	var m [184]byte
	copy(m[:], b)
	return unmarshal(m)
}

// Head returns the signed head.
func (sh *SignedHead) Head() string {
	return hex.Encode(sh.head[:])
}

// PubKey returns the public key in base64 notation.
func (sh *SignedHead) PubKey() string {
	return base64.Encode(sh.pubKey[:])
}

// HeadBuf returns the signed head.
func (sh *SignedHead) HeadBuf() [32]byte {
	var b [32]byte
	copy(b[:], sh.head[:])
	return b
}

// PrintTXT prints the TXT record to publish the signed head.
func (sh *SignedHead) PrintTXT(dns string) {
	fmt.Printf("%s%s.\t\t%d\tIN\tTXT\t\"%s\"\n",
		def.CodechainTXTName, dns, TTL, sh.Marshal())
}
