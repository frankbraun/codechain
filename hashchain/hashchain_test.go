package hashchain

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ed25519"
)

func TestStart(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "hashchain_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	_, sec, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey() failed: %v", err)
	}
	var secKey [64]byte
	copy(secKey[:], sec)

	filename := filepath.Join(tmpdir, "hashchain")
	c, entry, err := Start(filename, secKey, nil)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer c.Close()
	fmt.Println(entry)

	filename = filepath.Join(tmpdir, "hashchain2")
	c, entry, err = Start(filename, secKey, []byte("comment"))
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer c.Close()
	fmt.Println(entry)

}
