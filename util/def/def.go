// Package def defines default values used in Codechain.
package def

import (
	"os"
	"path/filepath"
)

// DefaultCodechainDir is the default directory used for Codechain related files.
// Can be overwritten with the environment variable CODECHAIN_DIR.
const DefaultCodechainDir = ".codechain"

// CodechainDir is the directory used for Codechain releated files. If not set
// with the environment variable CODECHAIN_DIR, DefaultCodechainDir is used.
// If CODECHAIN_DIR is used, the environment variable CODECHAIN_EXCLUDE can be
// used to exclude a second Codechain directory from all Codechain commands.
var CodechainDir = DefaultCodechainDir

func init() {
	dir := os.Getenv("CODECHAIN_DIR")
	if dir != "" {
		CodechainDir = dir
		ExcludePaths = append(ExcludePaths, dir)
	}
	exclude := os.Getenv("CODECHAIN_EXCLUDE")
	if exclude != "" {
		ExcludePaths = append(ExcludePaths, exclude)
	}
	HashchainFile = filepath.Join(CodechainDir, "hashchain")
	PatchDir = filepath.Join(CodechainDir, "patches")
}

// SecretsSubDir is the default subdirectory of a tool's home directory used
// to store secret key files
const SecretsSubDir = "secrets"

// CodechainHeadName is the TXT entry used for Codechain's secpkg heads.
const CodechainHeadName = "_codechain-head."

// CodechainURLName is the TXT entry used for Codechain's secpkg URLs.
const CodechainURLName = "_codechain-url."

// CodechainTestName is the TXT entry used to test Dyn credentials.
const CodechainTestName = "_codechain-test."

// ExcludePaths is the default list of paths not considered by Codechain.
// Do not ever change this list! It will break existing Codechains.
var ExcludePaths = []string{
	DefaultCodechainDir,
	".git",
	".gitignore",
	".travis.yml",
}

// HashchainFile is the default name of the hashchain file.
var HashchainFile string

// PatchDir is the default name of the patch file directory.
var PatchDir string
