package secpkg

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/hex"
)

func TestInstallBinpkg(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "secpkg_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	fn := filepath.Join("testdata", "binpkg", "binpkg.secpkg")
	pkg, err := Load(fn)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: factor out
	buf, err := hex.Decode(pkg.Head, 32)
	if err != nil {
		t.Fatalf("hex.Decode() failed: %v", err)
	}
	var head [32]byte
	copy(head[:], buf)
	_, sec, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey() failed: %v", err)
	}
	var sk [64]byte
	copy(sk[:], sec)
	sh, err := ssot.SignHeadV2(head, 2, 0, sk, nil, ssot.MaximumValidity)
	if err != nil {
		t.Fatalf("SignHeadV2() failed: %v", err)
	}

	res := newMockResolver()
	res.Heads["binpkg.secpkg.net"] = sh
	url := "https://frankbraun.org/secpkg/binpkg"
	res.URLs["binpkg.secpkg.net"] = []string{url}
	fn = "7705087e3d673d1089ea77bf567263c51b427371b293553f53ef23e254d1a3e1.tar.gz"
	res.Files[url+"/"+fn] = filepath.Join("testdata", "binpkg", fn)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = pkg.Install(context.Background(), res, tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}

	// TODO: test that binary is installed
}
