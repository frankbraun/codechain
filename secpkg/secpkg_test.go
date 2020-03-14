package secpkg

import (
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/util/hex"
)

func TestLoadMarshalNew(t *testing.T) {
	codechainSecPkg := filepath.Join("testdata", "codechain.secpkg")
	pkg, err := Load(codechainSecPkg)
	if err != nil {
		t.Fatalf("Load(%s) failed: %v", codechainSecPkg, err)
	}
	if pkg.Key != "" {
		t.Error("pkg.Key not empty")
	}
	_ = pkg.Marshal()
	h, err := hex.Decode(pkg.Head, 32)
	if err != nil {
		t.Fatalf("hex.Decode(%s, 32) failed: %v", pkg.Head, err)
	}
	var head [32]byte
	copy(head[:], h)
	encPkg, err := New(pkg.Name, pkg.DNS, nil, head, true)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	_, err = hex.Decode(encPkg.Key, 32)
	if err != nil {
		t.Fatalf("hex.Decode(%s, 32) failed: %v", encPkg.Key, err)
	}

	codechain2SecPkg := filepath.Join("testdata", "codechain2.secpkg")
	pkg, err = Load(codechain2SecPkg)
	if err != nil {
		t.Fatalf("Load(%s) failed: %v", codechainSecPkg, err)
	}
	if len(pkg.DNS2) != 2 {
		t.Fatal("two secondary DNS entries expected")
	}
	_, err = New(pkg.Name, pkg.DNS, pkg.DNS2, head, true)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
}
