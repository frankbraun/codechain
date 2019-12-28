package hashchain

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/file"
)

func TestDeepVerify(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "hashchain_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	hashChainFile := filepath.Join("..", def.HashchainFile)
	hashChainTmp := filepath.Join(tmpdir, "hashchain")
	err = file.Copy(hashChainFile, hashChainTmp)
	if err != nil {
		t.Fatalf("file.Copy() failed: %v", err)
	}

	c, err := ReadFile(hashChainTmp)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	defer c.Close()

	c.DeepVerify(tmpdir, filepath.Join("..", def.PatchDir), def.ExcludePaths)
	if err != nil {
		t.Fatalf("DeepVerify() failed: %v", err)
	}
}
