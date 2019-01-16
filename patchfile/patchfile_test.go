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
	err := Diff(1, ioutil.Discard, tree, tree, nil)
	if err != ErrNoDifference {
		t.Error("Diff() should fail with ErrNoDifference")
	}

}

/*
func TestDiffNotClean(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "patchfile_test")
	if err != nil {
		t.Fatalf("ioutil.TempDir() failed: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	emptyDir := filepath.Join(tmpdir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("os.Mkdir() failed: %v", err)
	}
	startDir := filepath.Join(tmpdir, "start")
	if err := os.Mkdir(startDir, 0755); err != nil {
		t.Fatalf("os.Mkdir() failed: %v", err)
	}
	dmpDir := filepath.Join("testdata", "dmp")

	// version 1 should fail
	err = Diff(1, ioutil.Discard, emptyDir, dmpDir, nil)
	if err != ErrDiffNotClean {
		t.Fatal("Diff() should fail with ErrDiffNotClean")
	}
	// version 2 should succeed
	err = Diff(2, ioutil.Discard, emptyDir, dmpDir, nil)
	if err != nil {
		t.Fatalf("Diff() failed: %v", err)
	}

	err = ioutil.WriteFile(filepath.Join(startDir, "tables.go"), nil, 0644)
	if err != nil {
		t.Fatalf("ioutil.WriteFile() failed: %v", err)
	}

	// version 1 should fail
	err = Diff(1, ioutil.Discard, startDir, dmpDir, nil)
	if err != ErrDiffNotClean {
		t.Fatal("Diff() should fail with ErrDiffNotClean")
	}
	// version 2 should succeed
	err = Diff(2, ioutil.Discard, startDir, dmpDir, nil)
	if err != nil {
		t.Fatalf("Diff() failed: %v", err)
	}
}
*/

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
	helloMove2Dir := filepath.Join("testdata", "hellomove2")
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
	binPatchV2, err := ioutil.ReadFile(filepath.Join("testdata", "binary.patch.v2"))
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	bin2Patch, err := ioutil.ReadFile(filepath.Join("testdata", "binary2.patch"))
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	bin2PatchV2, err := ioutil.ReadFile(filepath.Join("testdata", "binary2.patch.v2"))
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	scriptfileDir := filepath.Join("testdata", "scriptfile")
	scriptDir := filepath.Join("testdata", "script")
	script2Dir := filepath.Join("testdata", "script2")
	windowsDir := filepath.Join("testdata", "windows")
	xyDir := filepath.Join("testdata", "xy")
	yDir := filepath.Join("testdata", "y")

	testCases := []struct {
		version int
		a       string
		b       string
		patch   string
	}{
		{
			1,
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
			2,
			emptyDir,
			helloDir,
			`codechain patchfile version 2
treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
+ f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
utf8file 10
package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello world!")
}

treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
`,
		},
		{
			1,
			helloDir,
			emptyDir,
			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
`,
		},
		{
			2,
			helloDir,
			emptyDir,
			`codechain patchfile version 2
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
`,
		},
		{
			1,
			emptyDir,
			binaryDir,
			string(binPatch),
		},
		{
			2,
			emptyDir,
			binaryDir,
			string(binPatchV2),
		},
		{
			1,
			binaryDir,
			emptyDir,
			`codechain patchfile version 1
treehash 8e5e10cf5cb59a4d81f8e145adf208775436577eb53916c7ff195c252cab2989
- f 927d2cae58bb53cdd087bb7178afeff9dab8ec1691cbd01aeccae62559da2791 gopher.png
treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
`,
		},
		{
			2,
			binaryDir,
			emptyDir,
			`codechain patchfile version 2
treehash 8e5e10cf5cb59a4d81f8e145adf208775436577eb53916c7ff195c252cab2989
- f 927d2cae58bb53cdd087bb7178afeff9dab8ec1691cbd01aeccae62559da2791 gopher.png
treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
`,
		},
		{
			1,
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
			2,
			helloDir,
			helloExecDir,
			`codechain patchfile version 2
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ x ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
treehash 6defacb74e7e7795c822bb947a19cf5e300e54ddbbd3c889af559785ff2b1a6e
`,
		},
		{
			1,
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
			2,
			helloDir,
			hello2Dir,
			`codechain patchfile version 2
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
			1,
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
			2,
			helloDir,
			helloMoveDir,
			`codechain patchfile version 2
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hellomove.go
utf8file 10
package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello world!")
}

treehash 5d0d150f44985c9500c43785e7a9f3cce1c458053906d4aef709e8dae19247b6
`,
		},
		{
			1,
			helloDir,
			helloMove2Dir,
			`codechain patchfile version 1
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ f 1b239e494fa201667627de82f0e4dc27b7b00b6ec06146e4d062730bf3762141 hellomove.go
dmppatch 2
@@ -0,0 +1,94 @@
+package main%0A%0Aimport (%0A%09%22fmt%22%0A)%0A%0Afunc main() %7B%0A%09fmt.Println(%22hello world, second version!%22)%0A%7D%0A
treehash a0b37dfb7b79f877a922aa4aecc4d9d2a91c4db0f6e337caddbee6bf89f5f0fd
`,
		},
		{
			2,
			helloDir,
			helloMove2Dir,
			`codechain patchfile version 2
treehash 5998c63aca42e471297c0fa353538a93d4d4cfafe9a672df6989e694188b4a92
- f ad125cc5c1fb680be130908a0838ca2235db04285bcdd29e8e25087927e7dd0d hello.go
+ f 1b239e494fa201667627de82f0e4dc27b7b00b6ec06146e4d062730bf3762141 hellomove.go
utf8file 10
package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello world, second version!")
}

treehash a0b37dfb7b79f877a922aa4aecc4d9d2a91c4db0f6e337caddbee6bf89f5f0fd
`,
		},
		{
			1,
			binaryDir,
			binary2Dir,
			string(bin2Patch),
		},
		{
			2,
			binaryDir,
			binary2Dir,
			string(bin2PatchV2),
		},
		{
			1,
			scriptfileDir,
			scriptDir,
			`codechain patchfile version 1
treehash 2ce831fd2aa55ec8295fc16e6e08f4acfac4cc459295695a05005ccf293ab773
- f d02dfe1902fd0fa2d8a65aa3ec659b37b9b05fe5b34634795661370d18dcadc1 script.sh
+ x d02dfe1902fd0fa2d8a65aa3ec659b37b9b05fe5b34634795661370d18dcadc1 script.sh
treehash bb9bf35e526963439e9530539442982acd28cf272250436808d99023cb79a7ff
`,
		},
		{
			2,
			scriptfileDir,
			scriptDir,
			`codechain patchfile version 2
treehash 2ce831fd2aa55ec8295fc16e6e08f4acfac4cc459295695a05005ccf293ab773
- f d02dfe1902fd0fa2d8a65aa3ec659b37b9b05fe5b34634795661370d18dcadc1 script.sh
+ x d02dfe1902fd0fa2d8a65aa3ec659b37b9b05fe5b34634795661370d18dcadc1 script.sh
treehash bb9bf35e526963439e9530539442982acd28cf272250436808d99023cb79a7ff
`,
		},
		{
			1,
			scriptfileDir,
			script2Dir,
			`codechain patchfile version 1
treehash 2ce831fd2aa55ec8295fc16e6e08f4acfac4cc459295695a05005ccf293ab773
- f d02dfe1902fd0fa2d8a65aa3ec659b37b9b05fe5b34634795661370d18dcadc1 script.sh
+ x 63ca0be196825cd98dd10f6e3f8e62729af9fb44acc18f4d495818268bdf4077 script.sh
dmppatch 4
@@ -8,24 +8,40 @@
 sh%0A%0A
-echo %22hello world!%22%0A
+echo %22hello world, second version!%22%0A
treehash 38b133487dbffcc5eb7a46a405495e107444e55664042f08b787764100008acf
`,
		},
		{
			2,
			scriptfileDir,
			script2Dir,
			`codechain patchfile version 2
treehash 2ce831fd2aa55ec8295fc16e6e08f4acfac4cc459295695a05005ccf293ab773
- f d02dfe1902fd0fa2d8a65aa3ec659b37b9b05fe5b34634795661370d18dcadc1 script.sh
+ x 63ca0be196825cd98dd10f6e3f8e62729af9fb44acc18f4d495818268bdf4077 script.sh
dmppatch 4
@@ -8,24 +8,40 @@
 sh%0A%0A
-echo %22hello world!%22%0A
+echo %22hello world, second version!%22%0A
treehash 38b133487dbffcc5eb7a46a405495e107444e55664042f08b787764100008acf
`,
		},
		{
			1,
			scriptDir,
			script2Dir,
			`codechain patchfile version 1
treehash bb9bf35e526963439e9530539442982acd28cf272250436808d99023cb79a7ff
- x d02dfe1902fd0fa2d8a65aa3ec659b37b9b05fe5b34634795661370d18dcadc1 script.sh
+ x 63ca0be196825cd98dd10f6e3f8e62729af9fb44acc18f4d495818268bdf4077 script.sh
dmppatch 4
@@ -8,24 +8,40 @@
 sh%0A%0A
-echo %22hello world!%22%0A
+echo %22hello world, second version!%22%0A
treehash 38b133487dbffcc5eb7a46a405495e107444e55664042f08b787764100008acf
`,
		},
		{
			2,
			scriptDir,
			script2Dir,
			`codechain patchfile version 2
treehash bb9bf35e526963439e9530539442982acd28cf272250436808d99023cb79a7ff
- x d02dfe1902fd0fa2d8a65aa3ec659b37b9b05fe5b34634795661370d18dcadc1 script.sh
+ x 63ca0be196825cd98dd10f6e3f8e62729af9fb44acc18f4d495818268bdf4077 script.sh
dmppatch 4
@@ -8,24 +8,40 @@
 sh%0A%0A
-echo %22hello world!%22%0A
+echo %22hello world, second version!%22%0A
treehash 38b133487dbffcc5eb7a46a405495e107444e55664042f08b787764100008acf
`,
		},
		{
			1,
			xyDir,
			yDir,
			`codechain patchfile version 1
treehash 1951dd180a9b108c190e95afd5d0635dff0a5f9fa875ecb089e5cde7ccd0da93
- f e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 x.txt
treehash 9794e713367060c0d3ce5e6657af8d2ea7a04ea16e3865e7bea7c2aca0025de6
`,
		},
		{
			2,
			xyDir,
			yDir,
			`codechain patchfile version 2
treehash 1951dd180a9b108c190e95afd5d0635dff0a5f9fa875ecb089e5cde7ccd0da93
- f e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 x.txt
treehash 9794e713367060c0d3ce5e6657af8d2ea7a04ea16e3865e7bea7c2aca0025de6
`,
		},
		{
			1,
			yDir,
			xyDir,
			`codechain patchfile version 1
treehash 9794e713367060c0d3ce5e6657af8d2ea7a04ea16e3865e7bea7c2aca0025de6
+ f e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 x.txt
dmppatch 0
treehash 1951dd180a9b108c190e95afd5d0635dff0a5f9fa875ecb089e5cde7ccd0da93
`,
		},
		{
			2,
			yDir,
			xyDir,
			`codechain patchfile version 2
treehash 9794e713367060c0d3ce5e6657af8d2ea7a04ea16e3865e7bea7c2aca0025de6
+ f e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 x.txt
utf8file 1

treehash 1951dd180a9b108c190e95afd5d0635dff0a5f9fa875ecb089e5cde7ccd0da93
`,
		},
		{
			1,
			emptyDir,
			windowsDir,
			`codechain patchfile version 1
treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
+ f c74a0d7ee56e85b9235425adb26c44d56fba04fa8131a7996fe5c540dfdc03e2 test.txt
dmppatch 2
@@ -0,0 +1,15 @@
+foo%0D%0Abar%0D%0Abaz%0D%0A
treehash 51400bd59c9ba8e05474c39e2c7cda018b524a99910521e358d791817a51af55
`,
		},
		{
			2,
			emptyDir,
			windowsDir,
			"codechain patchfile version 2\n" +
				"treehash e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855\n" +
				"+ f c74a0d7ee56e85b9235425adb26c44d56fba04fa8131a7996fe5c540dfdc03e2 test.txt\n" +
				"utf8file 4\n" +
				"foo\r\n" +
				"bar\r\n" +
				"baz\r\n" +
				"\n" +
				"treehash 51400bd59c9ba8e05474c39e2c7cda018b524a99910521e358d791817a51af55\n",
		},
	}

	for i, testCase := range testCases {
		t.Logf("test case %d\n", i+1)
		var out bytes.Buffer
		err := Diff(testCase.version, &out, testCase.a, testCase.b, nil)
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
