## Codechain walkthrough

```
# [single tmux window]
# Let's assume you have `go` installed and $GOPATH set
# (otherwise go to https://golang.org/dl/)

# get Codechain repository from GitHub
go get -u -v github.com/frankbraun/codechain

# change to Codechain directory
cd $GOPATH/src/github.com/frankbraun/codechain

# use Codechain to verify Codechain code and switch to latest published version
codechain cleanslate
codechain apply
go install github.com/frankbraun/codechain

# start Codechain walkthrough with example project
cd doc/hellproject
ls
cat README.md
cat hello.go

# Codechain has various commands
codechain -h

# `codechain treehash` computes the hash of a directory tree
codechain treehash

# the hash is computed by hashing a deterministic tree list
codechain treehash -l
codechain treehash -l | sha256sum

# let's generate a key pair
codechain keygen

# let's start using Codechain for our example project
codechain start -s ~/.config/codechain/secrets/...

# this started the hash chain
cat .codechain/hashchain

# also:
codechain status -p

# see current status of project
codechain status

# let's publish our first release
codechain publish

# the first release has been published, but is not signed yet
codechain status

# let's review and sign it
codechain review

# now we have our first signed release
codechain status

# let's bring a second reviewer on board
# [switch tmux window ctrl+b n]
# the reviewer already has Codechain installed
# generate a key
codechain keygen

# [switch tmux window ctrl+b n]
# add second signer
codechain addkey ...

# increase number of necessary signers
codechain sigctl -m 2
codechain status

# sign-off on second signer
codechain review
codechain status

# add UNLICENSE to project
cp ../../UNLICENSE .

# publish new version
codechain publish
codechain status

# sign new release
codechain review
codechain status

# we still need the second signature, create distribution
codechain createdist -f /tmp/dist.tar.gz

# [switch tmux window]
# apply new distribution
codechain apply -f /tmp/dist.tar.gz

# create detached signature and send it to John
codechain review -d

# [switch tmux window]
# add detached signature
codechain review -a ...

# now we a new signed release
codechain status

# let's remove the second reviewer again
codechain sigctl -m 2
codechain remkey ...
codechain review

# of course, Jane has to sign off on it
codechain status
codechain createdist -f /tmp/dist2.tar.gz

# [switch tmux window]
# apply new distribution
codechain apply -f /tmp/dist.tar.gz

# create detached signature and send it to John
codechain review -d

# [switch tmux window]
# add detached signature
codechain review -a ...

# now we are back to one review
codechain status

```
