// codechain establishes code trust with a hashchain of threshold signatures.
package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/home"
	"github.com/frankbraun/codechain/util/lockfile"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	codechainDir  = ".codechain"
	hashchainFile = "hashchain"
	secretsDir    = "secrets"
	sigctlType    = "sigctl"
	sourceType    = "source"
	signatureType = "signtr"
	addkeyType    = "addkey"
	remkeyType    = "remkey"
)

var (
	excludePaths = []string{
		codechainDir,
		".git",
		".gitignore",
	}
	emptyTree []byte
)

func init() {
	var err error
	emptyTree, err = hex.DecodeString(tree.EmptyHash)
	if err != nil {
		panic(err)
	}
}

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


init [done]
addkey
remkey
sign
verify [done]
checkout
sigctl
pull
commit
signers

TODO:
- sign pubkey and comment and display that

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



	sigctlType    = "sigctl"
	sourceType    = "source"
	signatureType = "signtr"
	addkeyType    = "addkey"
	remkeyType    = "remkey"
*/

type codeChain []link

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
		strings.Join(l.typeFields, ""))
}

func newCodeChain(m int) codeChain {
	var c codeChain
	l := link{
		previous:   emptyTree,
		datum:      time.Now().UTC().Unix(),
		linkType:   sigctlType,
		typeFields: []string{strconv.Itoa(m)},
	}
	c = append(c, l)
	return c
}

func readCodeChain() (codeChain, error) {
	var c codeChain
	f, err := os.Open(filepath.Join(codechainDir, hashchainFile))
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
	return c, nil
}

func (c codeChain) prevHash() []byte {
	h := sha256.Sum256([]byte(c[len(c)-1].String()))
	return h[:]
}

func (c *codeChain) addKey(pubkey, signature, comment string) error {
	key := []string{pubkey, signature}
	if comment != "" {
		key = append(key, " "+comment)
	}
	l := link{
		previous:   c.prevHash(),
		datum:      time.Now().UTC().Unix(),
		linkType:   addkeyType,
		typeFields: key,
	}
	return c.appendLink(l)
}

func (c codeChain) save(w io.Writer) error {
	for _, link := range c {
		if _, err := fmt.Fprintln(w, link.String()); err != nil {
			return err
		}
	}
	return nil
}

func (c *codeChain) appendLink(l link) error {
	hashfile := filepath.Join(codechainDir, hashchainFile)
	lock, err := lockfile.Create(hashfile)
	if err != nil {
		return err
	}
	defer lock.Release()
	*c = append(*c, l)
	f, err := os.OpenFile(hashfile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprintln(f, l.String()); err != nil {
		return err
	}
	return nil
}

func (c codeChain) verify() error {
	// TODO
	return nil
}

func readPassphrase(confirm bool) ([]byte, error) {
	var (
		pass   []byte
		pass2  []byte
		reader *bufio.Reader
		err    error
	)
	isTerminal := terminal.IsTerminal(syscall.Stdin)
	fmt.Printf("passphrase: ")
	if isTerminal {
		pass, err = terminal.ReadPassword(syscall.Stdin)
		fmt.Println("")
	} else {
		reader = bufio.NewReader(os.Stdin)
		pass, err = reader.ReadBytes('\n')
	}
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("unable to read passphrase")
		}
		return nil, err
	}
	if len(pass) == 0 {
		return nil, errors.New("please provide a passphrase")
	}
	pass = bytes.TrimRight(pass, "\n")
	if confirm {
		fmt.Printf("confirm passphrase: ")
		if isTerminal {
			pass2, err = terminal.ReadPassword(syscall.Stdin)
			fmt.Println("")
		} else {
			pass2, err = reader.ReadBytes('\n')
		}
		if err != nil {
			return nil, err
		}
		defer bzero.Bytes(pass2)
		pass2 = bytes.TrimRight(pass2, "\n")
		if !bytes.Equal(pass, pass2) {
			return nil, errors.New("passphrases don't match")
		}
	}
	return pass, nil
}

func readComment() ([]byte, error) {
	str, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("unable to read line")
		}
		return nil, err
	}
	return bytes.TrimSpace(str), nil
}

func createSecfile(filename string, pass, sec, sig, comment []byte) error {
	var (
		salt  [32]byte
		nonce [24]byte
		key   [32]byte
	)
	exists, err := file.Exists(filename)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("file '%s' exists already", filename)
	}
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return err
	}
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}
	derivedKey := argon2.IDKey(pass, salt[:], 1, 64*1024, 4, 32)
	copy(key[:], derivedKey)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	msg := append(sec, sig...)
	msg = append(msg, comment...)
	enc := secretbox.Seal(nil, msg, &nonce, &key)
	if _, err := f.Write(salt[:]); err != nil {
		return err
	}
	if _, err := f.Write(nonce[:]); err != nil {
		return err
	}
	if _, err := f.Write(enc); err != nil {
		return err
	}
	return nil
}

func readSecfile(filename string, pass []byte) ([]byte, []byte, []byte, error) {
	var (
		salt  [32]byte
		nonce [24]byte
		key   [32]byte
	)
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	defer f.Close()
	if _, err := f.Read(salt[:]); err != nil {
		return nil, nil, nil, err
	}
	if _, err := f.Read(nonce[:]); err != nil {
		return nil, nil, nil, err
	}
	derivedKey := argon2.IDKey(pass, salt[:], 1, 64*1024, 4, 32)
	copy(key[:], derivedKey)
	enc, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, nil, nil, err
	}
	msg, verify := secretbox.Open(nil, enc, &nonce, &key)
	if !verify {
		return nil, nil, nil, fmt.Errorf("cannot decrypt '%s'", filename)
	}
	return msg[:64], msg[64:128], msg[128:], nil
}

func genKey() error {
	var homeDir string
	app := os.Args[1]
	fs := flag.NewFlagSet(os.Args[0]+" "+app, flag.ExitOnError)
	seckey := fs.String("s", "", "Secret key file")
	fs.Parse(os.Args[2:])
	if *seckey == "" {
		homeDir = home.AppDataDir("codechain", false)
		homeDir = filepath.Join(homeDir, secretsDir)
		if err := os.MkdirAll(homeDir, 0700); err != nil {
			return err
		}
	}
	pass, err := readPassphrase(true)
	if err != nil {
		return err
	}
	defer bzero.Bytes(pass)
	fmt.Println("comment (e.g., name; can be empty):")
	comment, err := readComment()
	if err != nil {
		return err
	}
	pub, sec, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	sig := ed25519.Sign(sec, append(pub, comment...))
	pubEnc := base64.URLEncoding.EncodeToString(pub[:])
	if *seckey != "" {
		if err := createSecfile(*seckey, pass, sec, sig, comment); err != nil {
			return err
		}
	} else {
		filename := filepath.Join(homeDir, pubEnc)
		if err := createSecfile(filename, pass, sec, sig, comment); err != nil {
			return err
		}
		fmt.Println("secret key file created:")
		fmt.Println(filename)
	}
	fmt.Println("public key with signature and optional comment")
	fmt.Printf("%s %s", pubEnc,
		base64.URLEncoding.EncodeToString(sig))
	if len(comment) > 0 {
		fmt.Printf(" %s", string(comment))
	}
	fmt.Println("")
	return nil
}

func pubKey() error {
	app := os.Args[1]
	fs := flag.NewFlagSet(os.Args[0]+" "+app, flag.ExitOnError)
	seckey := fs.String("s", "", "Secret key file")
	fs.Parse(os.Args[2:])
	if *seckey == "" {
		return fmt.Errorf("%s: option -s is mandatory", app)
	}
	exists, err := file.Exists(*seckey)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s: file '%s' does not exist", app, *seckey)
	}
	pass, err := readPassphrase(false)
	if err != nil {
		return err
	}
	defer bzero.Bytes(pass)
	sec, sig, comment, err := readSecfile(*seckey, pass)
	if err != nil {
		return err
	}
	if !ed25519.Verify(sec[32:], append(sec[32:], comment...), sig) {
		return fmt.Errorf("signature does not verify")
	}
	fmt.Println("public key with signature and optional comment")
	fmt.Printf("%s %s",
		base64.URLEncoding.EncodeToString(sec[32:]),
		base64.URLEncoding.EncodeToString(sig))
	if len(comment) > 0 {
		fmt.Printf(" %s", string(comment))
	}
	fmt.Println("")
	return nil
}

func initChain() error {
	app := os.Args[1]
	fs := flag.NewFlagSet(os.Args[0]+" "+app, flag.ExitOnError)
	m := fs.Int("m", 1, "Signature threshold M")
	fs.Parse(os.Args[2:])
	if *m < 1 {
		return fmt.Errorf("%s: option -m must be >= 1", app)
	}
	if err := os.MkdirAll(codechainDir, 0700); err != nil {
		return err
	}
	hashchain := filepath.Join(codechainDir, hashchainFile)
	exists, err := file.Exists(hashchain)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s: file '%s' exists already", app, hashchain)
	}
	chain := newCodeChain(*m)
	if err := chain.save(os.Stdout); err != nil {
		return err
	}
	f, err := os.Create(hashchain)
	if err != nil {
		return err
	}
	defer f.Close()
	return chain.save(f)
}

func addKey() error {
	app := os.Args[1]
	fs := flag.NewFlagSet(os.Args[0]+" "+app, flag.ExitOnError)
	w := fs.Int("w", 1, "Signature weight W")
	fs.Parse(os.Args[2:])
	if *w < 1 {
		return fmt.Errorf("%s: option -w must be >= 1", app)
	}
	nArg := fs.NArg()
	if nArg != 2 && nArg != 3 {
		return fmt.Errorf("%s: expecting args: pubkey signature [comment]", app)
	}
	pubkey := fs.Arg(0)
	pub, err := base64.URLEncoding.DecodeString(pubkey)
	if err != nil {
		return fmt.Errorf("cannot decode pubkey: %s", err)
	}
	signature := fs.Arg(1)
	sig, err := base64.URLEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("cannot decode signature: %s", err)
	}
	var comment string
	if nArg == 3 {
		comment = fs.Arg(2)
	}
	if !ed25519.Verify(pub, append(pub, []byte(comment)...), sig) {
		return fmt.Errorf("signature does not verify")
	}
	c, err := readCodeChain()
	if err != nil {
		return err
	}
	if err := c.verify(); err != nil {
		return err
	}
	return c.addKey(pubkey, signature, comment)
}

func verifyChain() error {
	c, err := readCodeChain()
	if err != nil {
		return err
	}
	return c.verify()
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
	os.Exit(1)
}

func usage() {
	cmd := os.Args[0]
	fmt.Fprintf(os.Stderr, "Usage: %s treehash\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s treelist\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s genkey [-s seckey.bin]\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s pubkey -s seckey.bin\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s init [-m]\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s addkey [-w] pubkey signature [comment]\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s verify\n", cmd)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "treehash":
		hash, err := tree.Hash(".", excludePaths)
		if err != nil {
			fatal(err)
		}
		fmt.Printf("%x\n", hash[:])
	case "treelist":
		list, err := tree.List(".", excludePaths)
		if err != nil {
			fatal(err)
		}
		os.Stdout.Write(list)
	case "genkey":
		if err := genKey(); err != nil {
			fatal(err)
		}
	case "pubkey":
		if err := pubKey(); err != nil {
			fatal(err)
		}
	case "init":
		if err := initChain(); err != nil {
			fatal(err)
		}
	case "addkey":
		if err := addKey(); err != nil {
			fatal(err)
		}
	case "verify":
		if err := verifyChain(); err != nil {
			fatal(err)
		}
	default:
		usage()
	}
}
