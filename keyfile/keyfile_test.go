package keyfile

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ed25519"
)

func TestCreateRead(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "keyfile_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	filename := filepath.Join(tmpdir, "keyfile.bin")
	pass := []byte("passphrase")
	comment := []byte("comment")
	pub, sec, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey() failed: %v", err)
	}
	sig := ed25519.Sign(sec, append(pub, comment...))
	msg := append(sec[32:], comment...)
	if !ed25519.Verify(ed25519.PublicKey(sec[32:]), msg, sig) {
		t.Error("signature does not verify")
	}
	err = Create(filename, pass, sec, sig, comment)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	readSec, readSig, readComment, err := Read(filename, pass)
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}
	if !bytes.Equal(readSec, sec) {
		t.Error("readSec != sec")
	}
	if !bytes.Equal(readSig, sig) {
		t.Error("readSig != sig")
	}
	if !bytes.Equal(readComment, comment) {
		t.Error("readComment != comment")
	}
	readMsg := make([]byte, len(readSec[32:])+len(readComment))
	n := copy(readMsg, readSec[32:])
	copy(readMsg[n:], readComment)
	if !ed25519.Verify(readSec[32:], readMsg, readSig) {
		t.Error("read signature does not verify")
	}
}
