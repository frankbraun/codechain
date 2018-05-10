// Package def defines default values used in Codechain.
package def

// CodechainDir is the default directory used for Codechain related files.
const CodechainDir = ".codechain"

// ExcludePaths is the default list of paths not considered by Codechain.
var ExcludePaths = []string{
	CodechainDir,
	".git",
	".gitignore",
	".travis.yml",
}
