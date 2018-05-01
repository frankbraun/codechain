package file

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/tree"
)

func TestCopyDir(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "file_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	dst := filepath.Join(tmpdir, "dst")
	if err := CopyDir("testdata", dst); err != nil {
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
}
