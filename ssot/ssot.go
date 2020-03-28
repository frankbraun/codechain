package ssot

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/log"
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
type SignedHead interface {
	Version() int
	PubKey() string
	PubKeyRotate() string
	ValidFrom() int64
	ValidTo() int64
	Counter() uint64
	Head() string
	HeadBuf() [32]byte
	Line() int
	Signature() string
	Marshal() string
}

// MarshalText marshals signed head as text (for status output).
func MarshalText(sh SignedHead) string {
	var (
		b       bytes.Buffer
		expired string
	)
	validFrom := time.Unix(sh.ValidFrom(), 0)
	validTo := time.Unix(sh.ValidTo(), 0)
	if err := Valid(sh); err == ErrSignedHeadExpired {
		expired = color.RedString(" EXPIRED!")
	}
	fmt.Fprintf(&b, "PUBKEY:        %s\n", sh.PubKey())
	fmt.Fprintf(&b, "PUBKEY_ROTATE: %s\n", sh.PubKeyRotate())
	fmt.Fprintf(&b, "VALID_FROM:    %s\n", validFrom.Format(time.RFC3339))
	fmt.Fprintf(&b, "VALID_TO:      %s%s\n", validTo.Format(time.RFC3339), expired)
	fmt.Fprintf(&b, "COUNTER:       %d\n", sh.Counter())
	fmt.Fprintf(&b, "HEAD:          %s\n", sh.Head())
	if sh.Line() > 0 { // version 2
		fmt.Fprintf(&b, "LINE:          %d\n", sh.Line())
	}
	fmt.Fprintf(&b, "SIGNATURE:     %s\n", sh.Signature())
	return b.String()
}

// Unmarshal and verify a base64 encoded signed head.
func Unmarshal(signedHead string) (SignedHead, error) {
	b, err := b64.RawURLEncoding.DecodeString(signedHead)
	if err != nil {
		return nil, err
	}
	if len(b) == 184 { // version 1
		// TODO: remove in 2021.
		var m [184]byte
		copy(m[:], b)
		return unmarshalV1(m)
	}
	version := b[0]
	if version == 2 {
		return unmarshalV2(signedHead)
	}
	return nil, fmt.Errorf("ssot: signed head version %d not supported", version)
}

// Load and verify a base64 encoded signed head from filename.
func Load(filename string) (SignedHead, error) {
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
func LookupHead(ctx context.Context, dns string) (SignedHead, error) {
	txts, err := net.DefaultResolver.LookupTXT(ctx, def.CodechainHeadName+dns)
	if err != nil {
		return nil, err
	}
	var sh SignedHead
	for _, txt := range txts {
		// parse TXT records and look for signed head
		sh, err = Unmarshal(txt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ssot: cannot unmarshal: %s\n", txt)
			sh = nil // reset head (invalid)
			continue // try next TXT record
		}
		log.Printf("ssot: signed head found: %s\n", sh.Head())
		if err := Valid(sh); err != nil {
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

// LookupURLs looks up all URLs from dns and returns it.
func LookupURLs(ctx context.Context, dns string) ([]string, error) {
	txts, err := net.DefaultResolver.LookupTXT(ctx, def.CodechainURLName+dns)
	if err != nil {
		return nil, err
	}
	var URLs []string
	for _, txt := range txts {
		// parse TXT records as URL
		if _, err := url.Parse(txt); err != nil {
			fmt.Fprintf(os.Stderr, "cannot parse as URL: %s\n", txt)
			continue
		}
		URLs = append(URLs, txt)
		fmt.Printf("URL found: %s\n", txt)
	}
	if len(URLs) == 0 {
		return nil, errors.New("ssot: no valid TXT record for URL found")
	}
	return URLs, nil
}

// TXTPrintHead prints the TXT record to publish the signed head.
func TXTPrintHead(sh SignedHead, dns string) {
	fmt.Printf("%s%s.\t\t%d\tIN\tTXT\t\"%s\"\n",
		def.CodechainHeadName, dns, TTL, sh.Marshal())
}

// TXTPrintURL prints the TXT record to publish the url.
func TXTPrintURL(dns, url string) {
	fmt.Printf("%s%s.\t\t%d\tIN\tTXT\t\"%s\"\n",
		def.CodechainURLName, dns, TTL, url)
}
