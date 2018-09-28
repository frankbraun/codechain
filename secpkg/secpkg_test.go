package secpkg

import (
	"path/filepath"
	"testing"
)

func TestLoadMarshal(t *testing.T) {
	codechainSecPkg := filepath.Join("testdata", "codechain.secpkg")
	p, err := Load(codechainSecPkg)
	if err != nil {
		t.Fatalf("Load(%s) failed: %v", codechainSecPkg, err)
	}
	_ = p.Marshal()
}
