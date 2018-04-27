package command

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestKey(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "command_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	seckey := filepath.Join(tmpdir, "seckey.bin")
	testPass = "passphrase"
	testComment = "John Doe"
	err = GenKey("genkey", "-s", seckey)
	if err != nil {
		t.Fatalf("GenKey() failed: %v ", err)
	}
	err = PubKey("genkey", "-s", seckey)
	if err != nil {
		t.Errorf("PubKey() failed: %v ", err)
	}
}
