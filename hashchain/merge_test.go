package hashchain

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/util/file"
)

var (
	hashChainA    = filepath.Join("testdata", "hashchain_a")
	hashChainB    = filepath.Join("testdata", "hashchain_b")
	hashChainFork = filepath.Join("testdata", "hashchain_fork")
)

func TestMergeSucces(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "hashchain_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	hashChainTmp := filepath.Join(tmpdir, "hashchain")
	err = file.Copy(hashChainA, hashChainTmp)
	if err != nil {
		t.Fatalf("file.Copy() failed: %v", err)
	}

	chainTmp, err := ReadFile(hashChainTmp)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	defer chainTmp.Close()

	chainB, err := ReadFile(hashChainB)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	defer chainB.Close()

	err = chainTmp.Merge(chainB)
	if err != nil {
		t.Fatalf("Merge() failed: %v", err)
	}

	tmpBuf, err := ioutil.ReadFile(hashChainTmp)
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	bBuf, err := ioutil.ReadFile(hashChainB)
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	if !bytes.Equal(tmpBuf, bBuf) {
		t.Error("tmpBuf differs from bBuf")
	}
}

func TestMergeFail(t *testing.T) {
	chainA, err := ReadFile(hashChainA)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	defer chainA.Close()

	chainB, err := ReadFile(hashChainB)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	defer chainB.Close()

	chainFork, err := ReadFile(hashChainFork)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	defer chainFork.Close()

	err = chainB.Merge(chainA)
	if err != ErrNothingToMerge {
		t.Error("Merge() should fail with ErrNothingToMerge")
	}

	err = chainB.Merge(chainFork)
	if err != ErrCannotMerge {
		t.Error("Merge() should fail with ErrCannotMerge")
	}

}
