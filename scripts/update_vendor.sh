#!/bin/sh -ex

rm -f Gopkg.lock Gopkg.toml
rm -rf vendor
dep init -v
# go get -u -v github.com/frankbraun/slimdep
slimdep -r -v -a github.com/frankbraun/codechain
