package secpkg

import (
	"context"
	"fmt"
	"os"
	"time"
)

// UpToDateIfInstalled ensures that the package with name is up-to-date, if it is
// installed as a secure package. If the package is not installed as a secure
// package a corresponding message is shown on stderr.
func UpToDateIfInstalled(ctx context.Context, name string) error {
	needsUpdate, err := CheckUpdate(ctx, name)
	if err != nil {
		if err == ErrNotInstalled {
			fmt.Fprintf(os.Stderr, "WARNING: package '%s' not installed via `secpkg install`\n", name)
		} else {
			return err
		}
	} else if needsUpdate {
		fmt.Errorf("tool needs update (`secpkg update %s`)", name)
	}
	return nil
}

// UpToDate ensures that the package with name is up-to-date, if it is
// installed as a secure package. If the package is not installed as a secure
// package a corresponding message is shown on stderr.
//
// UpToDate times out after a while if DNS cannot be queried and return nil.
func UpToDate(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := UpToDateIfInstalled(ctx, name); err != nil {
		return err
	}
	return nil
}
