.PHONY: build build-all clean install test test-e2e build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-linux-arm64 build-windows-amd64

VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o agentviewer .

build-all: build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-linux-arm64 build-windows-amd64

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/agentviewer-darwin-arm64 .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/agentviewer-darwin-amd64 .

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/agentviewer-linux-amd64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/agentviewer-linux-arm64 .

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/agentviewer-windows-amd64.exe .

clean:
	rm -rf dist/ agentviewer

install:
	go install $(LDFLAGS) .

test:
	go test ./...

test-e2e:
	go test -tags=e2e ./...
