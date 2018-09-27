all:
	env GO111MODULE=on go install -mod vendor -v ./...

.PHONY: test update-vendor
test:
	go get github.com/frankbraun/gocheck
	gocheck -g -c

update-vendor:
	rm -rf vendor
	env GO111MODULE=on go get -u
	env GO111MODULE=on go mod vendor
