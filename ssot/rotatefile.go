package ssot

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
)

// RotateFile rotates the pkgDir/signed_head to pkgDir/previous_signed_head and saves
// signed head sh to pkgDir/signed_head.
func (sh *SignedHead) RotateFile(pkgDir string) error {
	prevSignedHeadFile := filepath.Join(pkgDir, "previous_signed_head")
	exists, err := file.Exists(prevSignedHeadFile)
	if err != nil {
		return err
	}
	if exists {
		if err := os.Remove(prevSignedHeadFile); err != nil {
			return err
		}
	}
	signedHeadFile := filepath.Join(pkgDir, "signed_head")
	if err := file.Copy(signedHeadFile, prevSignedHeadFile); err != nil {
		return err
	}
	newSignedHeadFile := filepath.Join(pkgDir, "new_signed_head")
	err = ioutil.WriteFile(newSignedHeadFile, []byte(sh.Marshal()+"\n"), 0644)
	if err != nil {
		return err
	}
	if err := os.Rename(newSignedHeadFile, signedHeadFile); err != nil {
		return err
	}
	log.Printf("ssot: %s: written\n", signedHeadFile)
	return nil
}
