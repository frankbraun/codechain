prefix ?= /usr/local
exec_prefix ?= $(prefix)
bindir ?= $(exec_prefix)/bin

.PHONY: all install uninstall test test-install update-vendor help

all:
	env GO111MODULE=on go build -mod vendor -v . ./cmd/...

install:
	env GO111MODULE=on GOBIN=$(bindir) go install -mod vendor -v . ./cmd/secpkg ./cmd/ssotpub

uninstall:
	rm -f $(bindir)/codechain $(bindir)/secpkg $(bindir)/ssotpub

test:
	# go get github.com/frankbraun/gocheck
	# gocheck -g -c -v
	gocheck -c -v

test-install:
	go get github.com/frankbraun/gocheck
	go get golang.org/x/tools/cmd/goimports
	go get -u golang.org/x/lint/golint

update-vendor:
	rm -rf vendor
	env GO111MODULE=on go get -u
	env GO111MODULE=on go mod tidy -v
	env GO111MODULE=on go mod vendor

help:
	@echo "\
\n\
Codechain\n\
\n\
Commands:\n\
make (all)         - go build files to ./cmd\n\
make install       - go install required vendor library files (secpkg, ssot)\n\
make uninstall     - rm vendor libraries\n\
make test          - run gocheck\n\
make test-install  - install test dependencies (gocheck, golint, goimports)\n\
make update-vendor - update dependencies\n\
"
