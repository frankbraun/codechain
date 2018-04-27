// codechain establishes code trust with a hashchain of threshold signatures.
package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/home"
	"github.com/frankbraun/codechain/util/terminal"
	"golang.org/x/crypto/ed25519"
)

const (
	codechainDir = ".codechain"
	secretsDir   = "secrets"
)

var excludePaths = []string{
	codechainDir,
	".git",
	".gitignore",
}

var hashchainFile = filepath.Join(codechainDir, "hashchain")

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
	pass, err := terminal.ReadPassphrase(syscall.Stdin, true)
	if err != nil {
		return err
	}
	defer bzero.Bytes(pass)
	fmt.Println("comment (e.g., name; can be empty):")
	comment, err := terminal.ReadLine(os.Stdin)
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
		if err := keyfile.Create(*seckey, pass, sec, sig, comment); err != nil {
			return err
		}
	} else {
		filename := filepath.Join(homeDir, pubEnc)
		if err := keyfile.Create(filename, pass, sec, sig, comment); err != nil {
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
	pass, err := terminal.ReadPassphrase(syscall.Stdin, false)
	if err != nil {
		return err
	}
	defer bzero.Bytes(pass)
	sec, sig, comment, err := keyfile.Read(*seckey, pass)
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
	exists, err := file.Exists(hashchainFile)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s: file '%s' exists already", app, hashchainFile)
	}
	chain := hashchain.New(*m)
	if err := chain.Save(os.Stdout); err != nil {
		return err
	}
	f, err := os.Create(hashchainFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return chain.Save(f)
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
	c, err := hashchain.Read(hashchainFile)
	if err != nil {
		return err
	}
	if err := c.Verify(); err != nil {
		return err
	}
	return c.AddKey(hashchainFile, pubkey, signature, comment)
}

func verifyChain() error {
	c, err := hashchain.Read(hashchainFile)
	if err != nil {
		return err
	}
	return c.Verify()
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
