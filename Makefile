all:
	env GO111MODULE=on go install -mod vendor -v ./...

.PHONY: install test update-vendor
install: all

test:
	go get github.com/frankbraun/gocheck
	gocheck -g -c

update-vendor:
	rm -rf vendor
	env GO111MODULE=on go get -u
	env GO111MODULE=on go mod vendor
