package hashchain

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/lockfile"
	"github.com/frankbraun/codechain/util/time"
	"golang.org/x/crypto/ed25519"
)

const (
	chainStartType       = "cstart"
	sourceType           = "source"
	signatureType        = "signtr"
	addKeyType           = "addkey"
	removeKeyType        = "remkey"
	signatureControlType = "sigctl"
)

var emptyTree []byte

func init() {
	var err error
	emptyTree, err = hex.DecodeString(tree.EmptyHash)
	if err != nil {
		panic(err)
	}
}

type link struct {
	previous   []byte   // hash-of-previous
	datum      int64    // current-time
	linkType   string   // type
	typeFields []string // type-fields ...
}

func (l *link) String() string {
	return fmt.Sprintf("%x %s %s %s",
		l.previous,
		time.Format(l.datum),
		l.linkType,
		strings.Join(l.typeFields, " "))
}

// HashChain of threshold signatures over a chain of code changes.
type HashChain struct {
	lock  lockfile.Lock
	fp    *os.File
	chain []*link
	m     int // signature threshold
}

// verify hash chain.
func (c HashChain) verify() error {
	// TODO: make sure m is set correctly!
	// TODO: make sure the link types are all valid
	// TODO: check for empty hash chain
	return nil
}

// Start returns a new hash chain with signature control list m.
func Start(filename string, secKey [64]byte, comment []byte) (*HashChain, string, error) {
	var c HashChain
	exists, err := file.Exists(filename)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", fmt.Errorf("hashchain: file '%s' exists already", filename)
	}
	c.lock, err = lockfile.Create(filename)
	if err != nil {
		return nil, "", err
	}
	c.fp, err = os.Create(filename)
	if err != nil {
		return nil, "", err
	}

	// hash-of-previous current-time cstart pubkey nonce signature [comment]
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, "", err
	}
	pub := secKey[32:]
	msg := append(pub, nonce[:]...)
	if len(comment) > 0 {
		msg = append(msg, comment...)
	}
	sig := ed25519.Sign(secKey[:], msg)
	typeFields := []string{
		base64.Encode(pub),
		base64.Encode(nonce[:]),
		base64.Encode(sig[:]),
	}
	if len(comment) > 0 {
		typeFields = append(typeFields, string(comment))
	}
	l := &link{
		previous:   emptyTree,
		datum:      time.Now(),
		linkType:   chainStartType,
		typeFields: typeFields,
	}
	c.chain = append(c.chain, l)
	c.m = 1
	entry := l.String()
	if _, err := fmt.Fprintln(c.fp, entry); err != nil {
		return nil, "", err
	}
	return &c, entry, nil
}

// Read hash chain from filename.
func Read(filename string) (*HashChain, error) {
	var c HashChain
	exists, err := file.Exists(filename)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("hashchain: file '%s' doesn't exist", filename)
	}
	c.lock, err = lockfile.Create(filename)
	if err != nil {
		return nil, err
	}
	c.fp, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(c.fp)
	for s.Scan() {
		line := strings.SplitN(s.Text(), " ", 4)
		previous, err := hex.DecodeString(line[0])
		if err != nil {
			return nil, fmt.Errorf("hashchain: cannot decode hash '%s': %s", line[0], err)
		}
		t, err := time.Parse(line[1])
		if err != nil {
			return nil, fmt.Errorf("hashchain: cannot parse time '%s': %s", line[1], err)
		}
		l := &link{
			previous:   previous,
			datum:      t,
			linkType:   line[2],
			typeFields: strings.SplitN(line[3], " ", -1),
		}
		c.chain = append(c.chain, l)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if err := c.verify(); err != nil {
		return nil, err
	}
	c.m = 1
	return &c, nil
}

// Close the underlying file pointer of hash chain and release lock.
func (c *HashChain) Close() error {
	err := c.fp.Close()
	if err != nil {
		c.lock.Release()
		return err
	}
	return c.lock.Release()
}

// Source adds a source entry for treeHash and optional comment signed by
// secKey to the hash chain.
func (c *HashChain) Source(treeHash [32]byte, secKey [64]byte, comment []byte) (string, error) {
	pub := secKey[32:]
	msg := treeHash[:]
	if len(comment) > 0 {
		msg = append(msg, comment...)
	}
	sig := ed25519.Sign(secKey[:], msg)
	typeFields := []string{
		base64.Encode(treeHash[:]),
		base64.Encode(pub),
		base64.Encode(sig),
	}
	if len(comment) > 0 {
		typeFields = append(typeFields, string(comment))
	}
	prev := c.LastEntryHash()
	l := &link{
		previous:   prev[:],
		datum:      time.Now(),
		linkType:   sourceType,
		typeFields: typeFields,
	}
	c.chain = append(c.chain, l)
	entry := l.String()
	if _, err := fmt.Fprintln(c.fp, entry); err != nil {
		return "", err
	}
	return entry, nil
}

// Signature adds a signature entry for entryHash signed by secKey to the hash chain.
func (c *HashChain) Signature(entryHash [32]byte, secKey [64]byte) (string, error) {
	// TODO: check that entryHash is valid
	// TODO: make sure secKey is a valid signer
	// TODO: make sure entryHash is a valid position to sign
	pub := secKey[32:]
	sig := ed25519.Sign(secKey[:], entryHash[:])
	typeFields := []string{
		base64.Encode(entryHash[:]),
		base64.Encode(pub),
		base64.Encode(sig),
	}
	prev := c.LastEntryHash()
	l := &link{
		previous:   prev[:],
		datum:      time.Now(),
		linkType:   signatureType,
		typeFields: typeFields,
	}
	c.chain = append(c.chain, l)
	entry := l.String()
	if _, err := fmt.Fprintln(c.fp, entry); err != nil {
		return "", err
	}
	return entry, nil
}

// LastEntryHash returns the hash of the last entry.
func (c *HashChain) LastEntryHash() [32]byte {
	return sha256.Sum256([]byte(c.chain[len(c.chain)-1].String()))
}

// EntryHash returns the entry hash for the given treeHash.
func (c *HashChain) EntryHash(treeHash [32]byte) [32]byte {
	var h [32]byte
	// TODO: implement
	return h
}

// AddKey adds pubkey with signature and optional comment to hash chain.
func (c *HashChain) AddKey(pubKey [32]byte, signature [64]byte, comment []byte) (string, error) {
	if !ed25519.Verify(pubKey[:], append(pubKey[:], comment...), signature[:]) {
		return "", fmt.Errorf("signature does not verify")
	}
	typeFields := []string{
		base64.Encode(pubKey[:]),
		base64.Encode(signature[:]),
	}
	if len(comment) > 0 {
		typeFields = append(typeFields, string(comment))
	}
	prev := c.LastEntryHash()
	l := &link{
		previous:   prev[:],
		datum:      time.Now(),
		linkType:   addKeyType,
		typeFields: typeFields,
	}
	c.chain = append(c.chain, l)
	entry := l.String()
	if _, err := fmt.Fprintln(c.fp, entry); err != nil {
		return "", err
	}
	return entry, nil
}

// RemoveKey adds a pubkey remove entry to hash chain.
func (c *HashChain) RemoveKey(pubKey [32]byte) (string, error) {
	// TODO: check that pubkey is actually active in chain
	// TODO: check that still enough public keys remain to reach m
	prev := c.LastEntryHash()
	l := &link{
		previous:   prev[:],
		datum:      time.Now(),
		linkType:   removeKeyType,
		typeFields: []string{base64.Encode(pubKey[:])},
	}
	c.chain = append(c.chain, l)
	entry := l.String()
	if _, err := fmt.Fprintln(c.fp, entry); err != nil {
		return "", err
	}
	return entry, nil
}

// SignatureControl adds a signature control entry to the hash chain.
func (c *HashChain) SignatureControl(m int) (string, error) {
	// TODO: check that we have enough keys to reach m.
	if m <= 0 {
		return "", ErrSignatureThresholdNonPositive
	}
	prev := c.LastEntryHash()
	l := &link{
		previous:   prev[:],
		datum:      time.Now(),
		linkType:   signatureControlType,
		typeFields: []string{strconv.Itoa(m)},
	}
	c.chain = append(c.chain, l)
	entry := l.String()
	if _, err := fmt.Fprintln(c.fp, entry); err != nil {
		return "", err
	}
	return entry, nil
}
