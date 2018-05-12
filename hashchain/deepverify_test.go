package hashchain

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/internal/def"
)

func TestDeepVerify(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "hashchain_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	c, err := ReadFile(filepath.Join("..", def.HashchainFile))
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	defer c.Close()

	c.DeepVerify(tmpdir, filepath.Join("..", def.PatchDir), def.ExcludePaths)
	if err != nil {
		t.Fatalf("DeepVerify() failed: %v", err)
	}
}
