SOURCE_FILES?=$$(go list ./... | grep -v /vendor/ | grep -v /mocks/)
TEST_PATTERN?=.
TEST_OPTIONS?=-race -v

setup:
	go get -u github.com/alecthomas/gometalinter
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/dave/courtney
	gometalinter --install --update

test:
	echo 'mode: atomic' > coverage.txt && go list ./... | grep -v testing.go | xargs -n1 -I{} sh -c 'go test -v -failfast -p 1 -parallel 1 -timeout=600s -covermode=atomic -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.txt' && rm coverage.tmp

cover: test
	go tool cover -html=coverage.txt
# don't open browser...	go tool cover -html=coverage.txt -o coverage.html

fmt:
	goimports -w .

lint:
	golangci-lint run --tests=false --enable-all --disable lll --disable interfacer --disable gochecknoglobals

ci: lint test

BUILD_TAG := $(shell git describe --tags 2>/dev/null)
BUILD_SHA := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u '+%Y/%m/%d:%H:%M:%S')

build:
	GOOS=darwin CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o ".local_dist/sn-speedtest_darwin_amd64"

build-all:
	GOOS=darwin  CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o ".local_dist/sn-speedtest_darwin_amd64" 
	GOOS=linux   CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o ".local_dist/sn-speedtest_linux_amd64" 

build-linux:
	GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags '-s -w -X "main.version=[$(BUILD_TAG)-$(BUILD_SHA)] $(BUILD_DATE) UTC"' -o ".local_dist/sn-speedtest_linux_amd64" cmd/sn-speedtest/main.go

mac-install: build
	install .local_dist/sn-speedtest_darwin_amd64 /usr/local/bin/sn-speedtest

linux-install: build-linux
	sudo install .local_dist/sn-speedtest_linux_amd64 /usr/local/bin/sn-speedtest

critic:
	gocritic check-project .

find-updates:
	go list -u -m -json all | go-mod-outdated -update -direct

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build
