package command

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/util/file"
)

const (
	testPubkey = "0CLTw2MXB6kybhvlfTg6kkFxLkB-w7DM1SeNMWhUu2g"
	testSig    = "amHTVxtXCgXO0j8s7MujvbGdOej3OsrxkPlsAVDH1cfNyy4dY16u4fyrZK5KTMAsNvuWsQwiKAsr6R7Jo9IUDg"
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
	// codechain start -m 3
	err = Start("start", "-m", "3")
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
