package command

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/seckey"
)

const (
	testPubkey = "zx4xyVAbEfEdGeP1-yv-Jkv4BI0yoA1ySrAiVrSatb0"
	testSig    = "H8TsdqsqPV7ogkjqkfQq_m7sn2Xb8LzyWCOT0ZURKN4uGDlk_cmktt5bxzfIbJ-PTFj_q1kA1erTdKnZy0i_Aw"
)

func TestKey(t *testing.T) {
	tmpdirDist, err := ioutil.TempDir("", "command_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdirDist)
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

	if err := os.Chdir(".."); err != nil {
		t.Fatalf("os.Chdir() failed: %v", err)
	}
	err = os.MkdirAll(filepath.Join(tmpdirDist, def.CodechainDir), 0755)
	if err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	distFile := filepath.Join(tmpdirDist, def.CodechainDir, "dist.tar.gz")
	// codechain createdist -f
	err = CreateDist("createdist", "-f", distFile)
	if err != nil {
		t.Fatalf("DistFile() failed: %v ", err)
	}

	if err := os.Chdir(tmpdirDist); err != nil {
		t.Fatalf("os.Chdir() failed: %v", err)
	}

	// codechain apply -f -head wrong
	err = Apply("apply", "-f", distFile, "-head",
		"0000000000000000000000000000000000000000000000000000000000000000")
	if err != hashchain.ErrHeadNotFound {
		t.Fatal("Apply() should fail with hashchain.ErrHeadNotFound")
	}

	// codechain apply -f -head right
	err = Apply("apply", "-f", distFile, "-head",
		"734d22f9408b36141f5fe898db45d7095be539210f13f562905cc05baef5fd24")
	if err != nil {
		t.Fatalf("Apply() failed: %v ", err)
	}

	// codechain apply -f
	err = Apply("apply", "-f", distFile)
	if err != nil {
		t.Fatalf("Apply() failed: %v ", err)
	}

	if err := os.Chdir(tmpdir); err != nil {
		t.Fatalf("os.Chdir() failed: %v", err)
	}

	if err = os.RemoveAll(filepath.Join(tmpdir, def.CodechainDir)); err != nil {
		t.Fatalf("os.RemoveAll() failed: %v", err)
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
	seckey.TestPass = "passphrase"
	testComment = "John Doe"
	// codechain keygen -s seckey.bin
	err = KeyGen(homedir.Codechain(), "keygen", "-s", "seckey.bin")
	if err != nil {
		t.Fatalf("KeyGen() failed: %v ", err)
	}
	// codechain keyfile -s seckey.bin
	err = KeyFile(homedir.Codechain(), "pubkey", "-s", "testkey.bin")
	if err != nil {
		t.Errorf("KeyFile() failed: %v ", err)
	}
	// codechain start -m 3
	err = Start("start", "-s", "seckey.bin")
	if err != nil {
		t.Errorf("Start() failed: %v ", err)
	}
	exists, err := file.Exists(def.HashchainFile)
	if err != nil {
		t.Fatalf("file.Exists() failed: %v", err)
	}
	if !exists {
		t.Errorf("file '%s' doesn't exist", def.HashchainFile)
	}
	// codechain addkey -w 2 pubkey signature comment
	err = AddKey("addkey", "-w", "2", testPubkey, testSig, testComment)
	if err != nil {
		t.Errorf("AddKey() failed: %v ", err)
	}
	// codechain sigctl -m 3
	err = SigCtl("sigctl", "-m", "3")
	if err != nil {
		t.Errorf("AddKey() failed: %v ", err)
	}
	// codechain sigctl -m 1
	err = SigCtl("sigctl", "-m", "1")
	if err != nil {
		t.Errorf("AddKey() failed: %v ", err)
	}
	// codechain remkey pubkey
	err = RemKey("remkey", testPubkey)
	if err != nil {
		t.Errorf("RemKey() failed: %v ", err)
	}
	// codechain status
	err = Status("status")
	if err != nil {
		t.Errorf("Status() failed: %v ", err)
	}
	// codechain status -deep-verify
	err = Status("status", "-deep-verify")
	if err != nil {
		t.Errorf("Status() failed: %v ", err)
	}
}

func TestHelp(t *testing.T) {
	// codechain treehash -h
	err := TreeHash("codechain treehash", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain treehash -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain keygen -h
	err = KeyGen(homedir.Codechain(), "codechain keygen", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain keygen -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain keyfile -h
	err = KeyFile(homedir.Codechain(), "codechain keyfile", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain keyfile -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain start -h
	err = Start("codechain start", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain start -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain publish -h
	err = Publish("codechain publish", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain publish -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain review -h
	err = Review("codechain review", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain review -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain addkey -h
	err = AddKey("codechain addkey", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain addkey -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain remkey -h
	err = RemKey("codechain remkey", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain remkey -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain sigctl -h
	err = SigCtl("codechain sigctl", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain sigctl -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain createdist -h
	err = CreateDist("codechain createdist", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain createdist -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain apply -h
	err = Apply("codechain apply", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain apply -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain status -h
	err = Status("codechain status", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain status -h should fail with flag.ErrHelp: %v", err)
	}
	// codechain cleanslate -h
	err = CleanSlate("codechain cleanslate", "-h")
	if err != flag.ErrHelp {
		t.Errorf("codechain cleanslate -h should fail with flag.ErrHelp: %v", err)
	}
}
