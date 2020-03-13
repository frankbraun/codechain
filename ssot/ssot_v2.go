package ssot

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/frankbraun/codechain/util/base64"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/hex"
)

// SignedHeadV2 is a signed Codechain head ready for publication as a SSOT with
// DNS TXT records (version 2).
type SignedHeadV2 struct {
	version      uint8    // the version of the signed head
	pubKey       [32]byte // Ed25519 public key of SSOT head signer
	pubKeyRotate [32]byte // Ed25519 pubkey to rotate to, all 0 if unused
	validFrom    int64    // this signed head is valid from the given Unix time
	validTo      int64    // this signed head is valid to the given Unix time
	counter      uint64   // signature counter
	head         [32]byte // the Codechain head to sign
	line         uint32   // the last signed line number
	signature    [64]byte // signature with pubkey over all previous fields
}

// marshal signed head without signature.
func (sh *SignedHeadV2) marshal() [125]byte {
	var m [125]byte
	var b [8]byte
	var l [4]byte
	m[0] = sh.version
	copy(m[1:33], sh.pubKey[:])
	copy(m[33:65], sh.pubKeyRotate[:])
	binary.BigEndian.PutUint64(b[:], uint64(sh.validFrom))
	copy(m[65:73], b[:])
	binary.BigEndian.PutUint64(b[:], uint64(sh.validTo))
	copy(m[73:81], b[:])
	binary.BigEndian.PutUint64(b[:], sh.counter)
	copy(m[81:89], b[:])
	copy(m[89:121], sh.head[:])
	binary.BigEndian.PutUint32(l[:], sh.line)
	copy(m[121:125], l[:])
	return m
}

// Marshal signed head with signature and encode it as base64.
func (sh *SignedHeadV2) Marshal() string {
	var m [189]byte
	b := sh.marshal()
	copy(m[:125], b[:])
	copy(m[125:189], sh.signature[:])
	return base64.Encode(m[:])
}

// MarshalText marshals signed head as text (for status output).
func (sh *SignedHeadV2) MarshalText() string {
	var (
		b       bytes.Buffer
		expired string
	)
	validFrom := time.Unix(sh.validFrom, 0)
	validTo := time.Unix(sh.validTo, 0)
	if err := sh.Valid(); err == ErrSignedHeadExpired {
		expired = color.RedString(" EXPIRED!")
	}
	fmt.Fprintf(&b, "PUBKEY:        %s\n", base64.Encode(sh.pubKey[:]))
	fmt.Fprintf(&b, "PUBKEY_ROTATE: %s\n", base64.Encode(sh.pubKeyRotate[:]))
	fmt.Fprintf(&b, "VALID_FROM:    %s\n", validFrom.Format(time.RFC3339))
	fmt.Fprintf(&b, "VALID_TO:      %s%s\n", validTo.Format(time.RFC3339), expired)
	fmt.Fprintf(&b, "COUNTER:       %d\n", sh.counter)
	fmt.Fprintf(&b, "HEAD:          %s\n", hex.Encode(sh.head[:]))
	fmt.Fprintf(&b, "LINE:          %d\n", sh.line)
	fmt.Fprintf(&b, "SIGNATURE:     %s\n", base64.Encode(sh.signature[:]))
	return b.String()
}

func unmarshalV2(m [189]byte) (*SignedHeadV2, error) {
	var sh SignedHeadV2
	sh.version = m[0]
	copy(sh.pubKey[:], m[1:33])
	copy(sh.pubKeyRotate[:], m[33:65])
	sh.validFrom = int64(binary.BigEndian.Uint64(m[65:73]))
	sh.validTo = int64(binary.BigEndian.Uint64(m[73:81]))
	sh.counter = binary.BigEndian.Uint64(m[81:89])
	copy(sh.head[:], m[88:121])
	sh.line = binary.BigEndian.Uint32(m[121:125])
	copy(sh.signature[:], m[125:189])
	msg := sh.marshal()
	if !ed25519.Verify(sh.pubKey[:], msg[:], sh.signature[:]) {
		return nil, ErrSignedHeadSignature
	}
	return &sh, nil
}

// Head returns the signed head.
func (sh *SignedHeadV2) Head() string {
	return hex.Encode(sh.head[:])
}

// PubKey returns the public key in base64 notation.
func (sh *SignedHeadV2) PubKey() string {
	return base64.Encode(sh.pubKey[:])
}

// PubKeyRotate returns the public key rotate in base64 notation.
func (sh *SignedHeadV2) PubKeyRotate() string {
	return base64.Encode(sh.pubKeyRotate[:])
}

// Counter returns the counter of signed head.
func (sh *SignedHeadV2) Counter() uint64 {
	return sh.counter
}

// HeadBuf returns the signed head.
func (sh *SignedHeadV2) HeadBuf() [32]byte {
	var b [32]byte
	copy(b[:], sh.head[:])
	return b
}

// Line returns the last signed line number of signed head.
func (sh *SignedHeadV2) Line() int {
	return int(sh.line)
}

// TXTPrintHead prints the TXT record to publish the signed head.
func (sh *SignedHeadV2) TXTPrintHead(dns string) {
	fmt.Printf("%s%s.\t\t%d\tIN\tTXT\t\"%s\"\n",
		def.CodechainHeadName, dns, TTL, sh.Marshal())
}
