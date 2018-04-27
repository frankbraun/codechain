package command

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/util/file"
)

const (
	testPubkey = "xI-WeBW7mU6Y_l_KSmbJA953t9VAK_pMhdKvMv9SCKE="
	testSig    = "Ea73Y3AXt1myOtnjplhV8Gu9pvQ9a74ZnFBawIBn8cDgp70nfwL0tt9VBkPblt526QpIxdpuqOU9DVamEiHyCg=="
)

func TestKey(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "command_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	err = file.Copy(filepath.Join("testdata", "testkey.bin"),
		filepath.Join(tmpdir, "testkey.bin"))
	if err != nil {
		t.Fatalf("file.Copy() failed: %v", err)
	}
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatalf("os.Chdir() failed: %v", err)
	}
	// codechain treehash
	err = TreeHash("treehash")
	if err != nil {
		t.Errorf("TreeHash() failed: %v", err)
	}
	// codechain treelist
	err = TreeList("treelist")
	if err != nil {
		t.Errorf("TreeList() failed: %v", err)
	}
	testPass = "passphrase"
	testComment = "John Doe"
	// codechain genkey -s seckey.bin
	err = GenKey("genkey", "-s", "seckey.bin")
	if err != nil {
		t.Fatalf("GenKey() failed: %v ", err)
	}
	// codechain pubkey -s seckey.bin
	err = PubKey("pubkey", "-s", "seckey.bin")
	if err != nil {
		t.Errorf("PubKey() failed: %v ", err)
	}
	// codechain init -m 3
	err = InitChain("init", "-m", "3")
	if err != nil {
		t.Errorf("Init() failed: %v ", err)
	}
	exists, err := file.Exists(hashchainFile)
	if err != nil {
		t.Fatalf("file.Exists() failed: %v", err)
	}
	if !exists {
		t.Errorf("file '%s' doesn't exist", hashchainFile)
	}
	// codechain addkey -w 2 pubkey signature comment
	err = AddKey("addkey", "-w", "2", testPubkey, testSig, testComment)
	if err != nil {
		t.Errorf("AddKey() failed: %v ", err)
	}
	// codechain verify
	err = Verify("verify")
	if err != nil {
		t.Errorf("Verify() failed: %v ", err)
	}
}
