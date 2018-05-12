package sync_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/sync"
)

func TestDir(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "sync_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	c, err := hashchain.ReadFile(filepath.Join("..", def.HashchainFile))
	if err != nil {
		t.Fatalf("hashchain.ReadFile() failed: %v", err)
	}
	defer c.Close()

	treeHashes := c.TreeHashes()

	patchDir := filepath.Join("..", def.PatchDir)
	err = sync.Dir(tmpdir, treeHashes[len(treeHashes)-1], patchDir, treeHashes,
		def.ExcludePaths, false)
	if err != nil {
		t.Fatalf("sync.Dir() failed: %v", err)
	}

	err = os.Remove(filepath.Join(tmpdir, "README.md"))
	if err != nil {
		t.Fatalf("os.Remove() failed: %v", err)
	}

	err = sync.Dir(tmpdir, treeHashes[len(treeHashes)-1], patchDir, treeHashes,
		def.ExcludePaths, false)
	if err != sync.ErrCannotRemove {
		t.Fatalf("sync.Dir() should fail with sync.ErrCannotRemove")
	}

	err = sync.Dir(tmpdir, treeHashes[len(treeHashes)-1], patchDir, treeHashes,
		def.ExcludePaths, true)
	if err != nil {
		t.Fatalf("sync.Dir() failed: %v", err)
	}
}
