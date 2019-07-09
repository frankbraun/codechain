package ssot

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/log"
	"golang.org/x/crypto/ed25519"
)

// File defines the default file name for a signed head.
const File = "signed_head"

// MaximumValidity of signed heads.
const MaximumValidity = 30 * 24 * time.Hour // 30d

// MinimumValidity of signed heads.
const MinimumValidity = 1 * time.Hour // 1h

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

// MarshalText marshals signed head as text (for status output).
func (sh *SignedHead) MarshalText() string {
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
	fmt.Fprintf(&b, "SIGNATURE:     %s\n", base64.Encode(sh.signature[:]))
	return b.String()
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

// Load and verify a base64 encoded signed head from filename.
func Load(filename string) (*SignedHead, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	sh, err := Unmarshal(string(b))
	if err != nil {
		return nil, err
	}
	return sh, nil
}

// LookupHead and verify base64 encoded signed head from dns.
func LookupHead(ctx context.Context, dns string) (*SignedHead, error) {
	txts, err := net.DefaultResolver.LookupTXT(ctx, def.CodechainHeadName+dns)
	if err != nil {
		return nil, err
	}
	var sh *SignedHead
	for _, txt := range txts {
		// parse TXT records and look for signed head
		sh, err = Unmarshal(txt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ssot: cannot unmarshal: %s\n", txt)
			sh = nil // reset head (invalid)
			continue // try next TXT record
		}
		log.Printf("ssot: signed head found: %s\n", sh.Head())
		if err := sh.Valid(); err != nil {
			fmt.Printf("ssot: not valid: %v\n", err)
			sh = nil // reset head (invalid)
			continue // try next TXT record
		}
		break // valid TXT record found
	}
	if sh == nil {
		return nil, errors.New("ssot: no valid TXT record for head found")
	}
	return sh, nil
}

// LookupURL looks up URL from dns and returns it.
func LookupURL(ctx context.Context, dns string) (string, error) {
	txts, err := net.DefaultResolver.LookupTXT(ctx, def.CodechainURLName+dns)
	if err != nil {
		return "", err
	}
	var URL string
	for _, txt := range txts {
		// parse TXT records as URL
		if _, err := url.Parse(txt); err != nil {
			fmt.Fprintf(os.Stderr, "cannot parse as URL: %s\n", txt)
			continue
		}
		URL = txt
		fmt.Printf("URL found: %s\n", URL)
		break // valid TXT record found
	}
	if URL == "" {
		return "", errors.New("ssot: no valid TXT record for URL found")
	}
	return URL, nil
}

// Head returns the signed head.
func (sh *SignedHead) Head() string {
	return hex.Encode(sh.head[:])
}

// PubKey returns the public key in base64 notation.
func (sh *SignedHead) PubKey() string {
	return base64.Encode(sh.pubKey[:])
}

// PubKeyRotate returns the public key rotate in base64 notation.
func (sh *SignedHead) PubKeyRotate() string {
	return base64.Encode(sh.pubKeyRotate[:])
}

// Counter returns the counter of signed head.
func (sh *SignedHead) Counter() uint64 {
	return sh.counter
}

// HeadBuf returns the signed head.
func (sh *SignedHead) HeadBuf() [32]byte {
	var b [32]byte
	copy(b[:], sh.head[:])
	return b
}

// TXTPrintHead prints the TXT record to publish the signed head.
func (sh *SignedHead) TXTPrintHead(dns string) {
	fmt.Printf("%s%s.\t\t%d\tIN\tTXT\t\"%s\"\n",
		def.CodechainHeadName, dns, TTL, sh.Marshal())
}

// TXTPrintURL prints the TXT record to publish the url.
func TXTPrintURL(dns, url string) {
	fmt.Printf("%s%s.\t\t%d\tIN\tTXT\t\"%s\"\n",
		def.CodechainURLName, dns, TTL, url)
}
