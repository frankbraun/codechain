package archive

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
)

func TestCreateApply(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "archive_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	archiveA, err := ioutil.TempFile("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(archiveA.Name())

	archiveB, err := ioutil.TempFile("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(archiveB.Name())

	srcPatchDir := filepath.Join("..", ".codechain", "patches")
	hashchainFileA := filepath.Join("testdata", "hashchain_a")
	chainA, err := hashchain.ReadFile(hashchainFileA)
	if err != nil {
		t.Fatalf("hashchain.ReadFile() failed: %v", err)
	}
	defer chainA.Close()

	err = Create(archiveA, chainA, srcPatchDir)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	err = chainA.Close()
	if err != nil {
		t.Fatalf("chainA.Close() failed: %v", err)
	}
	err = archiveA.Close()
	if err != nil {
		t.Fatalf("archiveA.Close() failed: %v", err)
	}

	fpA, err := os.Open(archiveA.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer fpA.Close()

	hashchainFile := filepath.Join(tmpdir, def.HashchainFile)
	patchDir := filepath.Join(tmpdir, def.PatchDir)
	err = Apply(hashchainFile, patchDir, fpA)
	if err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	chainA, err = hashchain.ReadFile(hashchainFileA)
	if err != nil {
		t.Fatalf("hashchain.ReadFile() failed: %v", err)
	}
	defer chainA.Close()
	chainA.DeepVerify(tmpdir, patchDir, def.ExcludePaths)

	hashchainFileB := filepath.Join("testdata", "hashchain_b")
	chainB, err := hashchain.ReadFile(hashchainFileB)
	if err != nil {
		t.Fatalf("hashchain.ReadFile() failed: %v", err)
	}
	defer chainB.Close()

	err = Create(archiveB, chainB, srcPatchDir)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	err = chainB.Close()
	if err != nil {
		t.Fatalf("chainB.Close() failed: %v", err)
	}
	err = archiveB.Close()
	if err != nil {
		t.Fatalf("archiveB.Close() failed: %v", err)
	}

	fpB, err := os.Open(archiveB.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer fpB.Close()

	err = Apply(hashchainFile, patchDir, fpB)
	if err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	f, err := os.Open(filepath.Join("testdata", "empty.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	err = Apply(hashchainFile, patchDir, f)
	if err != ErrUnknownFile {
		t.Error("Apply() should fail with ErrUnknownFile")
	}

	chainB, err = hashchain.ReadFile(hashchainFileB)
	if err != nil {
		t.Fatalf("hashchain.ReadFile() failed: %v", err)
	}
	defer chainB.Close()
	chainB.DeepVerify(tmpdir, patchDir, def.ExcludePaths)
}
