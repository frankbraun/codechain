package lockfile

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLock(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "lockfile_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	anchor := filepath.Join(tmpdir, "testfile")
	l1, err := Create(anchor)
	if err != nil {
		t.Fatalf("l1 Create() failed: %v", err)
	}
	if _, err = Create(anchor); err == nil {
		t.Error("second Create() should fail")
	}
	if err := l1.Release(); err != nil {
		t.Fatalf("l1.Release() failed: %v", err)
	}
	l2, err := Create(anchor)
	if err != nil {
		t.Fatalf("l2 Create() failed: %v", err)
	}
	if err := l2.Release(); err != nil {
		t.Fatalf("l2.Release() failed: %v", err)
	}
}
