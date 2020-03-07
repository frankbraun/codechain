#!/bin/sh -e

# these have to be checked manually!
echo "------------------------------------------------------------------------"
echo "CHECK THESE HASHES MANUALLY!"
CCTREEHASH_SHA256="6a911b51cc047eea7d7b10c75cf1f50a98dea2a39e12ef3414d2f00f31a263f9"
CODECHAIN_BOOTSTRAP_TREE_HASH="7552080c78c26b8a2699c0ac4398028b907b673958f8ba26e01da2ac499b1acb"
echo "cctreehash.go $CCTREEHASH_SHA256"
echo "codechain-bootstrap $CODECHAIN_BOOTSTRAP_TREE_HASH"
echo "------------------------------------------------------------------------"

# select sha256 program
if [ -x "$(command -v sha256sum)" ]
then
  SHA256="sha256sum --tag"
else
  SHA256="sha256"
fi

# download cctreehash.go
cd /tmp
rm -f cctreehash.go
curl -O https://frankbraun.org/cctreehash.go

# verify its SHA-256
CCTREEHASH=$($SHA256 cctreehash.go)
if [ "$CCTREEHASH" = "SHA256 (cctreehash.go) = $CCTREEHASH_SHA256" ]
then
  echo "$SHA256 cctreehash.go matches"
else
  echo "$SHA256 cctreehash.go does not match"
  exit 1
fi

# download Codechain bootstrap
rm -rf codechain-bootstrap
curl -O https://frankbraun.org/codechain-bootstrap.tar.gz

# verify its tree hash
tar xzf codechain-bootstrap.tar.gz
cd codechain-bootstrap
CODECHAIN_BOOTSTRAP=$(go run ../cctreehash.go)
if [ "$CODECHAIN_BOOTSTRAP" = "$CODECHAIN_BOOTSTRAP_TREE_HASH" ]
then
  echo "codechain-bootstrap tree hash matches"
else
  echo "codechain-bootstrap tree hash does not match"
  exit 1
fi

# use the bootstrapped Codechain to install the most current Codechain version
go run cmd/secpkg/secpkg.go install .secpkg

# cleanup
cd ..
rm -rf codechain-bootstrap
rm -f cctreehash.go
