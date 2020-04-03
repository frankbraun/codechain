package secpkg

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"io/ioutil"
	mrand "math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/hex"
)

func signHead(head string) (ssot.SignedHead, error) {
	buf, err := hex.Decode(head, 32)
	if err != nil {
		return nil, err
	}
	var hb [32]byte
	copy(hb[:], buf)
	_, sec, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	var sk [64]byte
	copy(sk[:], sec)
	return ssot.SignHeadV2(hb, 2, 0, sk, nil, ssot.MaximumValidity)
}

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

	sh, err := signHead(pkg.Head)
	if err != nil {
		t.Fatalf("signHead() failed: %v", err)
	}
	res, err := newMockResolver()
	if err != nil {
		t.Fatal(err)
	}
	res.Heads["binpkg.secpkg.net"] = sh
	url := "https://frankbraun.org/secpkg/binpkg"
	res.URLs["binpkg.secpkg.net"] = []string{url}
	fn = "7705087e3d673d1089ea77bf567263c51b427371b293553f53ef23e254d1a3e1.tar.gz"
	res.Files[url+"/"+fn] = filepath.Join("testdata", "binpkg", fn)

	// install
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	err = pkg.Install(context.Background(), res, tmpdir)
	if err != nil {
		t.Fatal(err)
	}

	// make sure binpkg is installed and a binary
	bin := filepath.Join(tmpdir, "local", "bin", "binpkg")
	fi, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("binpkg not installed: %v", err)
	}
	if fi.Mode()&0100 != 0100 {
		t.Fatal("binpkg is not an executable")
	}

	// package already installed
	err = pkg.Install(context.Background(), res, tmpdir)
	if err == nil {
		t.Fatal("second install should fail")
	}
}

func TestInstallBinpkgFail(t *testing.T) {
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

	sh, err := signHead("3918a460d2145d1c4e65b7962c880ea3e4af3454b89cac210bc40b6d34d7bb4a")
	if err != nil {
		t.Fatalf("signHead() failed: %v", err)
	}
	res, err := newMockResolver()
	if err != nil {
		t.Fatal(err)
	}
	res.Heads["binpkg.secpkg.net"] = sh
	url := "https://frankbraun.org/secpkg/binpkg"
	res.URLs["binpkg.secpkg.net"] = []string{url}
	fn = "3918a460d2145d1c4e65b7962c880ea3e4af3454b89cac210bc40b6d34d7bb4a.tar.gz"
	res.Files[url+"/"+fn] = filepath.Join("testdata", "binpkg", fn)

	// install
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	err = pkg.Install(context.Background(), res, tmpdir)
	if err != ErrNoValidDNSEntry {
		t.Fatalf("failed with: %v", err)
	}
}

func TestInstallBinpkg2(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "secpkg_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	fn := filepath.Join("testdata", "binpkg", "binpkg2.secpkg")
	pkg, err := Load(fn)
	if err != nil {
		t.Fatal(err)
	}

	sh, err := signHead("3918a460d2145d1c4e65b7962c880ea3e4af3454b89cac210bc40b6d34d7bb4a")
	if err != nil {
		t.Fatalf("signHead() failed: %v", err)
	}
	res, err := newMockResolver()
	if err != nil {
		t.Fatal(err)
	}
	res.Heads["binpkg.secpkg.net"] = sh
	url := "https://frankbraun.org/secpkg/binpkg"
	res.URLs["binpkg.secpkg.net"] = []string{url}
	fn = "3918a460d2145d1c4e65b7962c880ea3e4af3454b89cac210bc40b6d34d7bb4a.tar.gz"
	res.Files[url+"/"+fn] = filepath.Join("testdata", "binpkg", fn)

	sh, err = signHead(pkg.Head)
	if err != nil {
		t.Fatalf("signHead() failed: %v", err)
	}
	res.Heads["binpkg.taz0.org"] = sh
	url0 := "https://taz0.org/secpkg/binpkg"
	url1 := "https://taz1.org/secpkg/binpkg"
	res.URLs["binpkg.taz0.org"] = []string{url0, url1}
	fn0 := "3918a460d2145d1c4e65b7962c880ea3e4af3454b89cac210bc40b6d34d7bb4a.tar.gz"
	fn1 := "7705087e3d673d1089ea77bf567263c51b427371b293553f53ef23e254d1a3e1.tar.gz"
	res.Files[url0+"/"+fn0] = filepath.Join("testdata", "binpkg", fn0)
	res.Files[url1+"/"+fn1] = filepath.Join("testdata", "binpkg", fn1)

	// make pkg.DNSRecords() non-deterministic
	mrand.Seed(0)

	// install
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	err = pkg.Install(context.Background(), res, tmpdir)
	if err != nil {
		t.Fatal(err)
	}

	// make sure binpkg is installed and a binary
	bin := filepath.Join(tmpdir, "local", "bin", "binpkg")
	fi, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("binpkg not installed: %v", err)
	}
	if fi.Mode()&0100 != 0100 {
		t.Fatal("binpkg is not an executable")
	}
}