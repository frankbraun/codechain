// Package hashchain implements a hash chain of signatures over a chain of code changes.
package hashchain

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/lockfile"
)

/*
hash-of-previous current-time type type-fields
type signature-control-list m
type source-hash hash-root-of-source sig-of-hash-root-of-source-by-pubkey [comment]
type signature pubkey1 sig1 pubkey2 sig2 ...
type signature hash-of-chain-entry pubkey2 sig2 ...
type pubkey-add w pubkey sig-of-pubkey-and-comment-with-pubkey [comment]
type pubkey-remove pubkey

type signature-control-list m
type pubkey-add w pubkey comment sig-of-pubkey-and-comment-with-pubkey
type pubkey-add w pubkey comment sig-of-pubkey-and-comment-with-pubkey
type pubkey-add w pubkey comment sig-of-pubkey-and-comment-with-pubkey
type signature pubkey1 sig1 pubkey2 sig2 ...


init (sigctl)  \
addkey         |
addkey         |-> init phase
addkey         |
init (commit)  /
sign           \
sign           |
sign           |-> setup phase, reach threshold
sign           |
sign           /
normal operation:

source
source
signtr
signtr

state:
- last accepted commit
- last accepted signature control threshold
- last accepted signer list
*/

const (
	sigctlType    = "sigctl"
	sourceType    = "source"
	signatureType = "signtr"
	addkeyType    = "addkey"
	remkeyType    = "remkey"
)

var emptyTree []byte

func init() {
	var err error
	emptyTree, err = hex.DecodeString(tree.EmptyHash)
	if err != nil {
		panic(err)
	}
}

// HashChain of threshold signatures over a chain of code changes.
type HashChain []link

type link struct {
	previous   []byte
	datum      int64
	linkType   string
	typeFields []string
}

func (l *link) String() string {
	return fmt.Sprintf("%x %s %s %s",
		l.previous,
		time.Unix(l.datum, 0).UTC().Format(time.RFC3339),
		l.linkType,
		strings.Join(l.typeFields, " "))
}

// New returns a new hash chain with signature control list m.
func New(m int) (HashChain, error) {
	if m <= 0 {
		return nil, ErrSignatureThresholdNonPositive
	}
	var c HashChain
	l := link{
		previous:   emptyTree,
		datum:      time.Now().UTC().Unix(),
		linkType:   sigctlType,
		typeFields: []string{strconv.Itoa(m)},
	}
	c = append(c, l)
	return c, nil
}

// SigCtl adds a signature control entry to the hash chain.
func (c *HashChain) SigCtl(filename string, m int) (string, error) {
	// TODO: check that we have enough keys to reach m.
	if m <= 0 {
		return "", ErrSignatureThresholdNonPositive
	}
	l := link{
		previous:   c.prevHash(),
		datum:      time.Now().UTC().Unix(),
		linkType:   sigctlType,
		typeFields: []string{strconv.Itoa(m)},
	}
	err := c.appendLink(filename, l)
	if err != nil {
		return "", err
	}
	return l.String(), nil
}

// Read hash chain from filename.
func Read(filename string) (HashChain, error) {
	var c HashChain
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.SplitN(s.Text(), " ", 4)
		t, err := time.Parse(time.RFC3339, line[1])
		if err != nil {
			return nil, err
		}
		l := link{
			previous:   []byte(line[0]),
			datum:      t.UTC().Unix(),
			linkType:   line[2],
			typeFields: strings.SplitN(line[3], " ", -1),
		}
		c = append(c, l)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if err := c.verify(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c HashChain) prevHash() []byte {
	h := sha256.Sum256([]byte(c[len(c)-1].String()))
	return h[:]
}

// AddKey adds pubkey with signature and optional comment to hash chain.
func (c *HashChain) AddKey(filename, pubkey, signature, comment string) (string, error) {
	key := []string{pubkey, signature}
	if comment != "" {
		key = append(key, comment)
	}
	l := link{
		previous:   c.prevHash(),
		datum:      time.Now().UTC().Unix(),
		linkType:   addkeyType,
		typeFields: key,
	}
	err := c.appendLink(filename, l)
	if err != nil {
		return "", err
	}
	return l.String(), nil
}

// RemKey adds pubkey remove entry to hash chain.
func (c *HashChain) RemKey(filename, pubkey string) (string, error) {
	// TODO: check that pubkey is actually active in chain
	// TODO: check that still enough public keys remain to reach M
	l := link{
		previous:   c.prevHash(),
		datum:      time.Now().UTC().Unix(),
		linkType:   remkeyType,
		typeFields: []string{pubkey},
	}
	err := c.appendLink(filename, l)
	if err != nil {
		return "", err
	}
	return l.String(), nil
}

// Save hash chain.
func (c HashChain) Save(w io.Writer) error {
	for _, link := range c {
		if _, err := fmt.Fprintln(w, link.String()); err != nil {
			return err
		}
	}
	return nil
}

func (c *HashChain) appendLink(filename string, l link) error {
	lock, err := lockfile.Create(filename)
	if err != nil {
		return err
	}
	defer lock.Release()
	*c = append(*c, l)
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprintln(f, l.String()); err != nil {
		return err
	}
	return nil
}

// verify hash chain.
func (c HashChain) verify() error {
	// TODO: implement and merge into Read
	return nil
}
