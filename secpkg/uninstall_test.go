package secpkg

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestUninstallBinpkgFail(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "secpkg_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// install
	err = Uninstall(tmpdir, "binpkg")
	if err == nil {
		t.Fatal("Uninstall() should fail")
	}
}
