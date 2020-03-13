package hashchain

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/frankbraun/codechain/util/hex"
)

const helloHashHex = "5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92"

var (
	pubA      [32]byte
	secA      [64]byte
	pubB      [32]byte
	secB      [64]byte
	helloHash [32]byte
)

func init() {
	pub, sec, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	copy(secA[:], sec[:])
	copy(pubA[:], pub[:])
	pub, sec, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	copy(secB[:], sec[:])
	copy(pubB[:], pub[:])
	hash, err := hex.Decode(helloHashHex, 32)
	if err != nil {
		panic(err)
	}
	copy(helloHash[:], hash)
}

func TestStartEmpty(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "hashchain_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	filename := filepath.Join(tmpdir, "hashchain")
	c, l, err := Start(filename, secA, nil)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer c.Close()
	fmt.Println(l)

	_, err = ReadFile(filename)
	if err == nil {
		t.Error("ReadFile() should fail (lockfile)")
	}
	err = c.Close()
	if err != nil {
		t.Fatalf("c.Close() failed: %v", err)
	}
	c2, err := ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile() 2 failed: %v", err)
	}
	defer c2.Close()

	_, ln := c2.LastSignedHead()
	if ln != 0 {
		t.Errorf("wrong signed line number %d != 0", ln)
	}
}

func TestStartSourceSign(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "hashchain_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// start empty chain
	filename := filepath.Join(tmpdir, "hashchain")
	c, l, err := Start(filename, secA, []byte("comment"))
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer c.Close()
	fmt.Println(l)

	// add hello.go
	l, err = c.Source(helloHash, secA, []byte("add hello.go"))
	if err != nil {
		t.Fatalf("c.Source() failed: %v", err)
	}
	fmt.Println(l)

	// sign hello.go (detached)
	detachedSig, err := c.Signature(c.Head(), secA, true)
	if err != nil {
		t.Fatalf("c.Signature() failed: %v", err)
	}
	parts := strings.SplitN(detachedSig, " ", 3)

	// add detached signature
	l, err = c.DetachedSignature(parts[0], parts[1], parts[2])
	if err != nil {
		t.Fatalf("c.DetachedSignature() failed: %v", err)
	}
	fmt.Println(l)
}

func TestStartAddKeySignSigCtlSign(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "hashchain_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// start empty chain
	filename := filepath.Join(tmpdir, "hashchain")
	c, l, err := Start(filename, secA, []byte("this is a comment"))
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer c.Close()
	fmt.Println(l)

	// addkey pubB hello.go
	sig := ed25519.Sign(secB[:], pubB[:])
	var signature [64]byte
	copy(signature[:], sig)
	l, err = c.AddKey(1, pubB, signature, nil)
	if err != nil {
		t.Fatalf("c.AddKey() failed: %v", err)
	}
	fmt.Println(l)
	if c.state.N() != 1 {
		t.Errorf("total weight should be n=1")
	}
	if c.state.HeadN() != 2 {
		t.Errorf("total weight including unconfirmed should be 2")
	}

	// sign other signer
	l, err = c.Signature(c.Head(), secA, false)
	if err != nil {
		t.Fatalf("c.Signature() failed: %v", err)
	}
	fmt.Println(l)
	if c.state.N() != 2 {
		t.Errorf("total weight should be n=2")
	}
	if !c.state.HasSigner(pubB) {
		t.Errorf("pubB should be a signer")
	}

	// sigctl
	_, err = c.SignatureControl(3)
	if err != ErrMLargerThanN {
		t.Errorf("should fail with ErrMLargerThanN")
	}
	l, err = c.SignatureControl(2)
	if err != nil {
		t.Fatalf("c.SignatureControl() failed")
	}
	fmt.Println(l)

	// sign sigctl
	l, err = c.Signature(c.Head(), secB, false)
	if err != nil {
		t.Fatalf("c.Signature() failed: %v", err)
	}
	fmt.Println(l)
}

func TestStartAddKeySignRemKeySign(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "hashchain_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// start empty chain
	filename := filepath.Join(tmpdir, "hashchain")
	c, l, err := Start(filename, secA, []byte("this is a comment"))
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer c.Close()
	fmt.Println(l)

	// addkey pubB hello.go
	sig := ed25519.Sign(secB[:], pubB[:])
	var signature [64]byte
	copy(signature[:], sig)
	l, err = c.AddKey(1, pubB, signature, nil)
	if err != nil {
		t.Fatalf("c.AddKey() failed: %v", err)
	}
	fmt.Println(l)
	if c.state.N() != 1 {
		t.Errorf("total weight should be n=1")
	}
	if c.state.HeadN() != 2 {
		t.Errorf("total weight including unconfirmed should be 2")
	}

	// sign other signer
	l, err = c.Signature(c.Head(), secA, false)
	if err != nil {
		t.Fatalf("c.Signature() failed: %v", err)
	}
	fmt.Println(l)
	if c.state.N() != 2 {
		t.Errorf("total weight should be n=2")
	}
	if !c.state.HasSigner(pubB) {
		t.Errorf("pubB should be a signer")
	}

	// remove key
	l, err = c.RemoveKey(pubA)
	if err != nil {
		t.Fatalf("c.RemoveKey() failed: %v", err)
	}
	fmt.Println(l)

	// sign sigctl
	l, err = c.Signature(c.Head(), secB, false)
	if err != nil {
		t.Fatalf("c.Signature() failed: %v", err)
	}
	fmt.Println(l)

	// read
	err = c.Close()
	if err != nil {
		t.Fatalf("c.Close() failed: %v", err)
	}
	c2, err := ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile() 2 failed: %v", err)
	}
	defer c2.Close()
}

func TestStartAddSameKey(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "hashchain_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// start empty chain
	filename := filepath.Join(tmpdir, "hashchain")
	c, l, err := Start(filename, secA, []byte("this is a comment"))
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer c.Close()
	fmt.Println(l)

	// addkey pubA
	sig := ed25519.Sign(secA[:], pubA[:])
	var signature [64]byte
	copy(signature[:], sig)
	_, err = c.AddKey(1, pubA, signature, nil)
	if err == nil {
		t.Fatalf("c.AddKey() should failv")
	}
}

const headStr = "0636b1e6faf8724ce3145b5de15ba4ffffacc6b852e1074d6a68721bfc0a8ecb"

var head [32]byte

func init() {
	h, err := hex.Decode(headStr, 32)
	if err != nil {
		panic(err)
	}
	copy(head[:], h)
}

func TestCheckHead(t *testing.T) {
	hashChainA = filepath.Join("testdata", "hashchain_a")
	hashChainB = filepath.Join("testdata", "hashchain_b")

	c, err := ReadFile(hashChainA)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	c.Close()
	err = c.CheckHead(head)
	if err != ErrHeadNotFound {
		t.Fatal("CheckHead() should fail with ErrHeadNotFound")
	}

	c, err = ReadFile(hashChainB)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	c.Close()
	err = c.CheckHead(head)
	if err != nil {
		t.Fatalf("CheckHead() failed: %v", err)
	}
}

func TestLastSignedHead(t *testing.T) {
	hashChainA = filepath.Join("testdata", "hashchain_a")
	hashChainB = filepath.Join("testdata", "hashchain_b")

	c, err := ReadFile(hashChainA)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	c.Close()
	h1, ln := c.LastSignedHead()
	h2 := c.Head()
	if !bytes.Equal(h1[:], h2[:]) {
		t.Error("wrong head")
	}
	if ln != 2 {
		t.Errorf("wrong signed line number %d != 2", ln)
	}

	c, err = ReadFile(hashChainB)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	c.Close()
	h1, ln = c.LastSignedHead()
	h2 = c.Head()
	if !bytes.Equal(h1[:], h2[:]) {
		t.Error("wrong head")
	}
	if ln != 4 {
		t.Errorf("wrong signed line number %d != 4", ln)
	}

	// add hello.go
	_, err = c.Source(helloHash, secA, []byte("add hello.go"))
	if err != nil {
		t.Fatalf("c.Source() failed: %v", err)
	}
	newHead, newLineNumber := c.LastSignedHead()
	h2 = c.Head()
	if !bytes.Equal(newHead[:], h1[:]) {
		t.Error("head changed")
	}
	if newLineNumber != ln {
		t.Error("line number changed")
	}
	if bytes.Equal(newHead[:], h1[:]) {
		t.Error("heads should differ")
	}
}
