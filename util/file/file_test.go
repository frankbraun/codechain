package file_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
)

func TestCopyDir(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "file_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	dst := filepath.Join(tmpdir, "dst")
	err = file.CopyDir("testdata", dst)
	if err != nil {
		t.Fatalf("CopyDir() failed: %v", err)
	}
	srcHash, err := tree.Hash("testdata", nil)
	if err != nil {
		t.Fatalf("tree.Hash(\"testdata\") failed: %v", err)
	}
	dstHash, err := tree.Hash(dst, nil)
	if err != nil {
		t.Fatalf("tree.Hash(dst) failed: %v", err)
	}
	if !bytes.Equal(srcHash[:], dstHash[:]) {
		t.Error("srcHash and dstHash differ")
	}

	// make dir which should be excluded
	err = os.Mkdir(filepath.Join(dst, def.CodechainDir), 0755)
	if err != nil {
		t.Fatalf("os.Mkdir() failed: %v", err)
	}
	dst2 := filepath.Join(tmpdir, "dst2")
	err = file.CopyDirExclude(dst, dst2, def.ExcludePaths)
	if err != nil {
		t.Fatalf("CopyDirExclude() failed: %v", err)
	}
	dst2Hash, err := tree.Hash(dst2, nil)
	if err != nil {
		t.Fatalf("tree.Hash(dst2) failed: %v", err)
	}
	if !bytes.Equal(srcHash[:], dst2Hash[:]) {
		t.Error("srcHash and dst2Hash differ")
	}

	// make sure dir wasn't copied
	exists, err := file.Exists(filepath.Join(dst2, def.CodechainDir))
	if err != nil {
		t.Fatalf("Exists() failed: %v", err)
	}
	if exists {
		t.Error("exists should be false")
	}
}

func TestIsBinaryTrue(t *testing.T) {
	isBinary, err := file.IsBinary(filepath.Join("testdata", "gopher.png"))
	if err != nil {
		t.Fatalf("IsBinary() failed: %v", err)
	}
	if !isBinary {
		t.Error("isBinary should be true")
	}
}

func TestIsBinaryFalse(t *testing.T) {
	isBinary, err := file.IsBinary(filepath.Join("testdata", "foo.txt"))
	if err != nil {
		t.Fatalf("IsBinary() failed: %v", err)
	}
	if isBinary {
		t.Error("isBinary should be false")
	}
}

func TestExistsTrue(t *testing.T) {
	exists, err := file.Exists(filepath.Join("testdata", "foo.txt"))
	if err != nil {
		t.Fatalf("Exists() failed: %v", err)
	}
	if !exists {
		t.Error("exists should be true")
	}
}
