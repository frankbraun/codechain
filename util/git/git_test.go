package git

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
)

const expectedPatch = `diff --git a/treeA/bar/baz.txt b/treeA/bar/baz.txt
new file mode 100644
index 0000000..5716ca5
--- /dev/null
+++ b/treeA/bar/baz.txt
@@ -0,0 +1 @@
+bar
diff --git a/treeA/foo.txt b/treeA/foo.txt
new file mode 100644
index 0000000..257cc56
--- /dev/null
+++ b/treeA/foo.txt
@@ -0,0 +1 @@
+foo
`

func TestDiffApply(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "git_test")
	if err != nil {
		t.Fatalf("TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	// fill treeA with testdata
	treeA := filepath.Join(tmpdir, "treeA")
	if err := file.CopyDir("testdata", treeA); err != nil {
		t.Fatalf("CopyDir() failed: %v", err)
	}
	// make empty treeB
	treeB := filepath.Join(tmpdir, "treeB")
	if err := os.Mkdir(treeB, 0700); err != nil {
		t.Fatalf("os.Mkdir() failed: %v", err)
	}
	// change into tmpdir
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatalf("os.Chdir() should not fail: %v", err)
	}
	// diff trees
	patch, err := Diff("treeB", "treeA")
	if err != nil {
		t.Fatalf("Diff() failed: %v", err)
	}
	if patch != expectedPatch {
		t.Error("patch != expectedPatch")
	}
	// apply patch to treeB
	r := bytes.NewBufferString(patch)
	err = Apply(r, 2, treeB, false)
	if err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}
	// compare tree hashes
	treeHashA, err := tree.Hash(treeA, nil)
	if err != nil {
		t.Fatalf("tree.Hash(treeA) failed: %v", err)
	}
	treeHashB, err := tree.Hash(treeB, nil)
	if err != nil {
		t.Fatalf("tree.Hash(treeB) failed: %v", err)
	}
	if !bytes.Equal(treeHashA[:], treeHashB[:]) {
		t.Error("treeHashA should equal treeHashB")
	}
	// try to apply patch to treeB again (should fail)
	r = bytes.NewBufferString(patch)
	err = Apply(r, 2, treeB, false)
	if err == nil {
		t.Error("Apply() should fail")
	}
	// diff trees again (should be empty)
	patch2, err := Diff("treeA", "treeB")
	if err != nil {
		t.Fatalf("Diff() failed: %v", err)
	}
	if patch2 != "" {
		t.Error("patch2 should be empty")
	}
	// apply patch to treeB in reverse
	r = bytes.NewBufferString(patch)
	err = Apply(r, 2, treeB, true)
	if err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}
	// make sure treeB is empty again
	treeHashB, err = tree.Hash(treeB, nil)
	if err != nil {
		t.Fatalf("tree.Hash(treeB) failed: %v", err)
	}
	if hex.EncodeToString(treeHashB[:]) != tree.EmptyHash {
		t.Errorf("tree.Hash(treeHashB) should return the EmptyHash")
	}
}
