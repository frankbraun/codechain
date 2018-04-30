package command

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/util/file"
)

const (
	testPubkey = "zx4xyVAbEfEdGeP1-yv-Jkv4BI0yoA1ySrAiVrSatb0"
	testSig    = "H8TsdqsqPV7ogkjqkfQq_m7sn2Xb8LzyWCOT0ZURKN4uGDlk_cmktt5bxzfIbJ-PTFj_q1kA1erTdKnZy0i_Aw"
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
	// codechain treehash -l
	err = TreeHash("treelist", "-l")
	if err != nil {
		t.Errorf("TreeList() -l failed: %v", err)
	}
	testPass = "passphrase"
	testComment = "John Doe"
	// codechain keyge -s seckey.bin
	err = KeyGen("keygen", "-s", "seckey.bin")
	if err != nil {
		t.Fatalf("KeyGen() failed: %v ", err)
	}
	// codechain keyfile -s seckey.bin
	err = KeyFile("pubkey", "-s", "testkey.bin")
	if err != nil {
		t.Errorf("KeyFile() failed: %v ", err)
	}
	// codechain start -m 3
	err = Start("start", "-s", "seckey.bin")
	if err != nil {
		t.Errorf("Start() failed: %v ", err)
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
	// codechain status
	err = Status("status")
	if err != nil {
		t.Errorf("Status() failed: %v ", err)
	}
}
