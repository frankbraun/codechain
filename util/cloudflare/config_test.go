package cloudflare

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "config_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	c := &Config{"foo", "bar"}

	// Write()
	filename := filepath.Join(tmpdir, ConfigFilename)
	if err := c.Write(filename); err != nil {
		t.Fatalf("c.Write() failed: %v", err)
	}

	// ReadConfig()
	rc, err := ReadConfig(filename)
	if err != nil {
		t.Fatalf("ReadConfig(%s) failed: %v", filename, err)
	}

	// compare
	if !reflect.DeepEqual(c, rc) {
		t.Error("Read config doesn't equal written config")
	}
}
