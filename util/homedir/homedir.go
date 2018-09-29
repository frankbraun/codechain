// Package homedir implements helper methods to get the home directories of
// various tools.
package homedir

import (
	"os"
	"strings"

	"github.com/frankbraun/codechain/util/home"
	"github.com/frankbraun/codechain/util/log"
)

func get(app string) string {
	env := strings.ToUpper(app) + "HOMEDIR"
	if homeDir := os.Getenv(env); homeDir != "" {
		log.Printf("$%s=%s", env, homeDir)
		return homeDir
	}
	homeDir := home.AppDataDir(app, false)
	log.Printf("homeDir: %s", homeDir)
	return homeDir
}

// Codechain returns the home directory for 'codechain'.
func Codechain() string {
	return get("codechain")
}

// SecPkg returns the home directory for 'secpkg'.
func SecPkg() string {
	return get("secpkg")
}

// SSOTPub returns the home directory for 'ssotpub'.
func SSOTPub() string {
	return get("ssotpub")
}
