package hashchain

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/internal/hex"
	"golang.org/x/crypto/ed25519"
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

	_, err = Read(filename)
	if err == nil {
		t.Error("Read() should fail (lockfile)")
	}
	err = c.Close()
	if err != nil {
		t.Fatalf("c.Close() failed: %v", err)
	}
	c2, err := Read(filename)
	if err != nil {
		t.Fatalf("Read() 2 failed: %v", err)
	}
	defer c2.Close()
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

	// sign hello.go
	l, err = c.Signature(c.LastEntryHash(), secA, false)
	if err != nil {
		t.Fatalf("c.Signature() failed: %v", err)
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
	l, err = c.Signature(c.LastEntryHash(), secA, false)
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
	l, err = c.Signature(c.LastEntryHash(), secB, false)
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
	l, err = c.Signature(c.LastEntryHash(), secA, false)
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
	l, err = c.Signature(c.LastEntryHash(), secB, false)
	if err != nil {
		t.Fatalf("c.Signature() failed: %v", err)
	}
	fmt.Println(l)

	// read
	err = c.Close()
	if err != nil {
		t.Fatalf("c.Close() failed: %v", err)
	}
	c2, err := Read(filename)
	if err != nil {
		t.Fatalf("Read() 2 failed: %v", err)
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
