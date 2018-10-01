prefix ?= /usr/local
exec_prefix ?= $(prefix)
bindir ?= $(exec_prefix)/bin

all:
	go build -v .

install:
	env GOBIN=$(bindir) go install -v .
