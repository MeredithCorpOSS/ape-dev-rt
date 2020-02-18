default: test

installdeps:
	go get -u golang.org/x/tools/cmd/stringer
	go get -u github.com/mitchellh/gox
	go get -u github.com/Masterminds/glide

dev: generate
	@RT_DEV=1 sh -c "'$(CURDIR)/scripts/build.sh'"

build: generate
	@RT_DEV= sh -c "'$(CURDIR)/scripts/build.sh'"

release: test generate
	sh -c "'$(CURDIR)/scripts/build.sh'"
	sh -c "'$(CURDIR)/scripts/release.sh'"

generate:
	go generate $$(go list ./... | grep -v /vendor/)

test: generate
	sh -c "'$(CURDIR)/scripts/fmtcheck.sh'"
	go test -cover $$(go list ./... | grep -v /vendor/)
	go vet

.PHONY: default installdeps dev
