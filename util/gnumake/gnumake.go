// Package gnumake contains wrappers around some GNU make commands.
package gnumake

import (
	"fmt"
	"os"
	"os/exec"
)

// Call 'make' with prefix=prefix.
func Call(prefix string) error {
	prefixStr := fmt.Sprintf("prefix=%s", prefix)
	cmd := exec.Command("make", prefixStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Install calls 'make install' with prefix=prefix.
func Install(prefix string) error {
	prefixStr := fmt.Sprintf("prefix=%s", prefix)
	cmd := exec.Command("make", prefixStr, "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Uninstall calls 'make uninstall' with prefix=prefix.
func Uninstall(prefix string) error {
	prefixStr := fmt.Sprintf("prefix=%s", prefix)
	cmd := exec.Command("make", prefixStr, "uninstall")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
