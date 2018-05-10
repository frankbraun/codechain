package patchfile

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/frankbraun/codechain/util/file"
)

func TestNoDifference(t *testing.T) {
	tree := filepath.Join("testdata", "tree")
	err := Diff(ioutil.Discard, tree, tree, nil)
	if err != ErrNoDifference {
		t.Error("Diff() should fail with ErrNoDifference")
	}

}

func TestErrorCases(t *testing.T) {
	testCases := []struct {
		patch     string
		errorCode error
	}{
		{
			"codechain",
			ErrHeaderFieldsNum,
		},
		{
			"codechain patchfile xxx 1",
			ErrHeaderFieldsText,
		},
		{
			"codechain patchfile version x",
			strconv.ErrSyntax,
		},
		{
			"codechain patchfile version 0",
			ErrHeaderVersion,
		},
		{
			`codechain patchfile version 1
treehash`,
			ErrTreeHashFieldsNum,
		},
		{
			`codechain patchfile version 1
treehashx hex`,
			ErrTreeHashFieldsText,
		},
		{
			`codechain patchfile version 1
treehash hex_hash`,
			hex.InvalidByteError('h'),
		},
		{

			`codechain patchfile version 1
treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
`,
			ErrTreeHashStartMismatch,
		},
		{

			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f xxx
`,
			ErrFileFieldsNum,
		},
		{

			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
x f hex_hash /dev/null 
`,
			ErrFileField0,
		},
		{

			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- y hex_hash /dev/null 
`,
			ErrFileField1,
		},
		{

			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
+ f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
`,
			ErrAddTargetFileExists,
		},
		{

			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello2.go
+ f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
`,
			ErrMoveTargetFileExists,
		},
		{

			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ x 15bb620236c7bba4ff1edbda701444c99ea5111e9d0b133329f8199a30fd26ac hello.go
ascii85 x
`,
			ErrDiffLinesParse,
		},
		{

			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ x 15bb620236c7bba4ff1edbda701444c99ea5111e9d0b133329f8199a30fd26ac hello.go
ascii85 0
`,
			ErrDiffLinesNonPositive,
		},
		{

			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ x 15bb620236c7bba4ff1edbda701444c99ea5111e9d0b133329f8199a30fd26ac hello.go
ascii86 1
`,
			ErrDiffModeUnknown,
		},
		{

			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ x 15bb620236c7bba4ff1edbda701444c99ea5111e9d0b133329f8199a30fd26ac hello.go
ascii85 1
`,
			ErrPrematureDiffEnd,
		},
	}

	helloDir := filepath.Join("testdata", "hello")
	for _, testCase := range testCases {
		err := Apply(helloDir, bytes.NewBufferString(testCase.patch), nil)
		switch e := err.(type) {
		case *strconv.NumError:
			if e.Err != testCase.errorCode {
				t.Fatalf("Apply(%s) should have error code: %v (has %v)", testCase.patch,
					testCase.errorCode, e.Err)
			}
		default:
			if err != testCase.errorCode {
				t.Fatalf("Apply(%s) should have error code: %v (has %v)", testCase.patch,
					testCase.errorCode, err)
			}
		}
	}
}

func TestDiffApply(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "patchfile_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	emptyDir := filepath.Join(tmpdir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("os.Mkdir() failed: %v", err)
	}
	helloDir := filepath.Join("testdata", "hello")
	hello2Dir := filepath.Join("testdata", "hello2")
	helloMoveDir := filepath.Join("testdata", "hellomove")
	helloExecDir := filepath.Join(tmpdir, "helloexec")
	if err := file.CopyDir(helloDir, helloExecDir); err != nil {
		t.Fatalf("file.Copy() failed: %v", err)
	}
	err = os.Chmod(filepath.Join(helloExecDir, "hello.go"), 0755)
	if err != nil {
		t.Fatalf("os.Chmod() failed: %v", err)
	}
	binaryDir := filepath.Join("testdata", "binary")
	binary2Dir := filepath.Join("testdata", "binary2")
	binPatch, err := ioutil.ReadFile(filepath.Join("testdata", "binary.patch"))
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	bin2Patch, err := ioutil.ReadFile(filepath.Join("testdata", "binary2.patch"))
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	xyDir := filepath.Join("testdata", "xy")
	yDir := filepath.Join("testdata", "y")

	testCases := []struct {
		a     string
		b     string
		patch string
	}{
		{
			emptyDir,
			helloDir,
			`codechain patchfile version 1
treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
+ f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
dmppatch 2
@@ -0,0 +1,78 @@
+package main%0A%0Aimport (%0A%09%22fmt%22%0A)%0A%0Afunc main() %7B%0A%09fmt.Println(%22hello world!%22)%0A%7D%0A
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
`,
		},
		{
			helloDir,
			emptyDir,
			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
`,
		},
		{
			emptyDir,
			binaryDir,
			string(binPatch),
		},
		{
			binaryDir,
			emptyDir,
			`codechain patchfile version 1
treehash 8e5e10cf5cb59a4d81f8e145adf208775436577eb53916c7ff195c252cab2989
- f 927d2cae58bb53cdd087bb7178afeff9dab8ec1691cbd01aeccae62559da2791 gopher.png
treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
`,
		},
		{
			helloDir,
			helloExecDir,
			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ x ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
treehash 6defacb74e7e7795c822bb947a19cf5e300e54ddbbd3c889af559785ff2b1a6e
`,
		},
		{
			helloDir,
			hello2Dir,
			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ f 1b239e494fa201667627de82f0e4dc27b7b00b6ec06146e4d062730bf3762141 hello.go
dmppatch 5
@@ -44,35 +44,51 @@
 ) %7B%0A
-%09fmt.Println(%22hello world!%22)%0A
+%09fmt.Println(%22hello world, second version!%22)%0A
 %7D%0A
treehash 127909f57efe02bce8ad9a943e24b09d0d6ee4005e4664d53dc867c27398ee6e
`,
		},
		{
			helloDir,
			helloMoveDir,
			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hellomove.go
dmppatch 2
@@ -0,0 +1,78 @@
+package main%0A%0Aimport (%0A%09%22fmt%22%0A)%0A%0Afunc main() %7B%0A%09fmt.Println(%22hello world!%22)%0A%7D%0A
treehash 5d0d150f44985c9500c43785e7a9f3cce1c458053906d4aef709e8dae19247b6
`,
		},
		{
			binaryDir,
			binary2Dir,
			string(bin2Patch),
		},
		{
			xyDir,
			yDir,
			`codechain patchfile version 1
treehash 1951dd180a9b108c190e95afd5d0635dff0a5f9fa875ecb089e5cde7ccd0da93
- f e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 x.txt
treehash 9794e713367060c0d3ce5e6657af8d2ea7a04ea16e3865e7bea7c2aca0025de6
`,
		},
		{
			yDir,
			xyDir,
			`codechain patchfile version 1
treehash 9794e713367060c0d3ce5e6657af8d2ea7a04ea16e3865e7bea7c2aca0025de6
+ f e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 x.txt
dmppatch 0
treehash 1951dd180a9b108c190e95afd5d0635dff0a5f9fa875ecb089e5cde7ccd0da93
`,
		},
	}

	for i, testCase := range testCases {
		var out bytes.Buffer
		err := Diff(&out, testCase.a, testCase.b, nil)
		if err != nil {
			t.Fatalf("Diff() failed: %v", err)
		}
		if out.String() != testCase.patch {
			t.Errorf("patch differs\nactual:\n%sexpected:\n%s", out.String(),
				testCase.patch)
		}
		applyDir := filepath.Join(tmpdir, strconv.Itoa(i))
		if err := file.CopyDir(testCase.a, applyDir); err != nil {
			t.Fatalf("file.Copydir() failed: %v", err)
		}
		err = Apply(applyDir, bytes.NewBufferString(testCase.patch), nil)
		if err != nil {
			t.Fatalf("Apply() failed: %v", err)
		}
	}
}
