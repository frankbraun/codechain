// Copyright (c) 2013-2014 The btcsuite developers
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// Source: https://github.com/btcsuite/btcutil/blob/master/appdata_test.go

package home

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"
	"unicode"
)

// TestAppDataDir tests the API for AppDataDir to ensure it gives expected
// results for various operating systems.
func TestAppDataDir(t *testing.T) {
	// App name plus upper and lowercase variants.
	appName := "myapp"
	appNameUpper := string(unicode.ToUpper(rune(appName[0]))) + appName[1:]
	appNameLower := string(unicode.ToLower(rune(appName[0]))) + appName[1:]

	// When we're on Windows, set the expected local and roaming directories
	// per the environment vars.  When we aren't on Windows, the function
	// should return the current directory when forced to provide the
	// Windows path since the environment variables won't exist.
	winLocal := "."
	winRoaming := "."
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		roamingAppData := os.Getenv("APPDATA")
		if localAppData == "" {
			localAppData = roamingAppData
		}
		winLocal = filepath.Join(localAppData, appNameUpper)
		winRoaming = filepath.Join(roamingAppData, appNameUpper)
	}

	// Get the home directory to use for testing expected results.
	var homeDir string
	usr, err := user.Current()
	if err != nil {
		t.Errorf("user.Current: %v", err)
		return
	}
	homeDir = usr.HomeDir

	// Mac app data directory.
	macAppData := filepath.Join(homeDir, "Library", "Application Support")

	tests := []struct {
		goos    string
		appName string
		roaming bool
		want    string
	}{
		// Various combinations of application name casing, leading
		// period, operating system, and roaming flags.
		{"windows", appNameLower, false, winLocal},
		{"windows", appNameUpper, false, winLocal},
		{"windows", "." + appNameLower, false, winLocal},
		{"windows", "." + appNameUpper, false, winLocal},
		{"windows", appNameLower, true, winRoaming},
		{"windows", appNameUpper, true, winRoaming},
		{"windows", "." + appNameLower, true, winRoaming},
		{"windows", "." + appNameUpper, true, winRoaming},
		{"linux", appNameLower, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"linux", appNameUpper, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"linux", "." + appNameLower, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"linux", "." + appNameUpper, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"darwin", appNameLower, false, filepath.Join(macAppData, appNameUpper)},
		{"darwin", appNameUpper, false, filepath.Join(macAppData, appNameUpper)},
		{"darwin", "." + appNameLower, false, filepath.Join(macAppData, appNameUpper)},
		{"darwin", "." + appNameUpper, false, filepath.Join(macAppData, appNameUpper)},
		{"openbsd", appNameLower, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"openbsd", appNameUpper, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"openbsd", "." + appNameLower, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"openbsd", "." + appNameUpper, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"freebsd", appNameLower, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"freebsd", appNameUpper, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"freebsd", "." + appNameLower, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"freebsd", "." + appNameUpper, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"netbsd", appNameLower, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"netbsd", appNameUpper, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"netbsd", "." + appNameLower, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"netbsd", "." + appNameUpper, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"plan9", appNameLower, false, filepath.Join(homeDir, appNameLower)},
		{"plan9", appNameUpper, false, filepath.Join(homeDir, appNameLower)},
		{"plan9", "." + appNameLower, false, filepath.Join(homeDir, appNameLower)},
		{"plan9", "." + appNameUpper, false, filepath.Join(homeDir, appNameLower)},
		{"unrecognized", appNameLower, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"unrecognized", appNameUpper, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"unrecognized", "." + appNameLower, false, filepath.Join(homeDir, ".config", appNameLower)},
		{"unrecognized", "." + appNameUpper, false, filepath.Join(homeDir, ".config", appNameLower)},

		// No application name provided, so expect current directory.
		{"windows", "", false, "."},
		{"windows", "", true, "."},
		{"linux", "", false, "."},
		{"darwin", "", false, "."},
		{"openbsd", "", false, "."},
		{"freebsd", "", false, "."},
		{"netbsd", "", false, "."},
		{"plan9", "", false, "."},
		{"unrecognized", "", false, "."},

		// Single dot provided for application name, so expect current
		// directory.
		{"windows", ".", false, "."},
		{"windows", ".", true, "."},
		{"linux", ".", false, "."},
		{"darwin", ".", false, "."},
		{"openbsd", ".", false, "."},
		{"freebsd", ".", false, "."},
		{"netbsd", ".", false, "."},
		{"plan9", ".", false, "."},
		{"unrecognized", ".", false, "."},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		ret := appDataDir(test.goos, test.appName, test.roaming)
		if ret != test.want {
			t.Errorf("appDataDir #%d (%s) does not match - "+
				"expected got %s, want %s", i, test.goos, ret,
				test.want)
			continue
		}
	}
}
