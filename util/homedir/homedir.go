// Package homedir implements helper methods to get the home directories of
// various tools.
package homedir

import (
	"os"

	"github.com/frankbraun/codechain/util/home"
	"github.com/frankbraun/codechain/util/log"
)

func Codechain() string {
	if homeDir := os.Getenv("CODECHAINHOMEDIR"); homeDir != "" {
		log.Printf("$CODECHAINHOMEDIR=%s", homeDir)
		return homeDir
	}
	homeDir := home.AppDataDir("codechain", false)
	log.Printf("homeDir: %s", homeDir)
	return homeDir
}

func SSOTPub() string {
	if homeDir := os.Getenv("SSOTPUBHOMEDIR"); homeDir != "" {
		log.Printf("$SSOTPUBHOMEDIR=%s", homeDir)
		return homeDir
	}
	homeDir := home.AppDataDir("ssotpub", false)
	log.Printf("homeDir: %s", homeDir)
	return homeDir
}
