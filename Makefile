all:
	go install -v github.com/frankbraun/codechain/...

.PHONY: test update-vendor
test:
	go get github.com/frankbraun/gocheck
	gocheck -g -c

update-vendor:
	rm -f Gopkg.lock Gopkg.toml
	rm -rf vendor
	dep init -v
	slimdep -r -v -a github.com/frankbraun/codechain
