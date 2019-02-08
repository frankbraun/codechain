// Package archive implements a simple archive format for `codechain apply -f`.
package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
	"golang.org/x/crypto/nacl/secretbox"
)

var (
	globalHashchainFile = path.Join(def.CodechainDir, "hashchain")
	globalPatchDir      = path.Join(def.CodechainDir, "patches")
)

// Create a new archive for the given hash chain and write it to w.
// patchDir must contain all the necessary patch files.
// The validity of the patch files is not verified!
func Create(w io.Writer, c *hashchain.HashChain, patchDir string) error {
	var buf bytes.Buffer
	zw := gzip.NewWriter(w)
	tw := tar.NewWriter(zw)

	// write hashchain file
	c.Fprint(&buf)
	hdr := &tar.Header{
		Name: globalHashchainFile,
		Mode: 0644,
		Size: int64(buf.Len()),
	}
	log.Printf("archive: write %s", globalHashchainFile)
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(buf.Bytes()); err != nil {
		return err
	}

	// write patch files
	treeHashes := c.TreeHashes()
	for i := 0; i < len(treeHashes)-1; i++ {
		treeHash := treeHashes[i]
		patch, err := ioutil.ReadFile(filepath.Join(patchDir, treeHash))
		if err != nil {
			return err
		}
		patchFile := path.Join(globalPatchDir, treeHash)
		hdr := &tar.Header{
			Name: patchFile,
			Mode: 0644,
			Size: int64(len(patch)),
		}
		log.Printf("archive: write %s", patchFile)
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write(patch); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}
	return zw.Close()
}

// Apply the archive read from r to the given hashchainFile and patchDir.
// If the hashchainFile is already present it must be transformable by
// appending to the hashchain present in r, otherwise an error is returned.
// If head is not nil the hash chain read from r must contain the given head.
func Apply(hashchainFile, patchDir string, r io.Reader, head *[32]byte) error {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	tr := tar.NewReader(zr)

	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break // end of archive
			}
			return err
		}
		log.Printf("archive: read %s", hdr.Name)
		if hdr.Name == globalHashchainFile {
			log.Printf("hashchainFile: %s", hashchainFile)
			exists, err := file.Exists(hashchainFile)
			if err != nil {
				return err
			}
			log.Printf("exists: %s", strconv.FormatBool(exists))
			if exists {
				// try to merge hashchain files
				c, err := hashchain.ReadFile(hashchainFile)
				if err != nil {
					return err
				}
				src, err := hashchain.Read(tr)
				if err != nil {
					c.Close()
					return err
				}
				if head != nil {
					if err := src.CheckHead(*head); err != nil {
						c.Close()
						return err
					}
				}
				err = c.Merge(src)
				if err != nil {
					c.Close()
					return nil
				}
				if err := c.Close(); err != nil {
					return err
				}
			} else {
				if err := os.MkdirAll(filepath.Dir(hashchainFile), 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(patchDir, 0755); err != nil {
					return err
				}
				src, err := hashchain.Read(tr)
				if err != nil {
					return err
				}
				if head != nil {
					if err := src.CheckHead(*head); err != nil {
						return err
					}
				}
				// save new hashchain file
				f, err := os.Create(hashchainFile)
				if err != nil {
					return err
				}
				if err := src.Fprint(f); err != nil {
					f.Close()
					os.Remove(f.Name())
					return err
				}
				if err := f.Close(); err != nil {
					return err
				}
			}
		} else if path.Dir(hdr.Name) == globalPatchDir {
			patchFile := filepath.Join(patchDir, path.Base(hdr.Name))
			exists, err := file.Exists(patchFile)
			if err != nil {
				return err
			}
			if exists {
				// we already have the patch file, skip it
				if _, err := io.Copy(ioutil.Discard, tr); err != nil {
					return err
				}
			} else {
				// save new patch file
				f, err := os.Create(patchFile)
				if err != nil {
					return err
				}
				if _, err := io.Copy(f, tr); err != nil {
					f.Close()
					return err
				}
				if err := f.Close(); err != nil {
					return err
				}
			}
		} else {
			log.Printf("hdr.Name:       %s", hdr.Name)
			log.Printf("globalPatchDir: %s", globalPatchDir)
			return ErrUnknownFile
		}
	}

	return zr.Close()
}

// ApplyFile applies the archive in filename to the given hashchainFile and patchDir.
// If the hashchainFile is already present it must be transformable by
// appending to the hashchain present in r, otherwise an error is returned.
// If head is not nil the hash chain read from filename must contain the given head.
func ApplyFile(hashchainFile, patchDir, filename string, head *[32]byte) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("applying distribution '%s'", filename)
	return Apply(hashchainFile, patchDir, f, head)
}

// ApplyEncryptedFile applies the encrypted archive in filename to the given
// hashchainFile and patchDir. If the hashchainFile is already present it must
// be transformable by appending to the hashchain present in r, otherwise an
// error is returned. If head is not nil the hash chain read from filename
// must contain the given head.
func ApplyEncryptedFile(hashchainFile, patchDir, filename string, head, key *[32]byte) error {
	log.Printf("applying encrypted distribution '%s'", filename)
	enc, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var nonce [24]byte
	copy(nonce[:], enc[:24])
	msg, verify := secretbox.Open(nil, enc[24:], &nonce, key)
	if !verify {
		return ErrCannotDecrypt
	}
	return Apply(hashchainFile, patchDir, bytes.NewBuffer(msg), head)
}
