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

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
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
func Apply(hashchainFile, patchDir string, r io.Reader) error {
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
			exists, err := file.Exists(hashchainFile)
			if err != nil {
				return err
			}
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
				err = c.Merge(src)
				if err != nil {
					c.Close()
					return nil
				}
				if err := c.Close(); err != nil {
					return err
				}
			} else {
				if err := os.MkdirAll(patchDir, 0755); err != nil {
					return err
				}
				// save new hashchain file
				f, err := os.Create(hashchainFile)
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
