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
	l, err = c.Signature(c.LastEntryHash(), secA)
	if err != nil {
		t.Fatalf("c.Signature() failed: %v", err)
	}
	fmt.Println(l)
}
