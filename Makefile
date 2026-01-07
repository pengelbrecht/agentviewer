.PHONY: build build-all clean install test test-e2e build-darwin-arm64 build-darwin-amd64 build-linux-amd64 build-linux-arm64 build-windows-amd64 package-deb-amd64 package-deb-arm64 package-rpm-amd64 package-rpm-arm64 package-all

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

# Linux package targets (requires nfpm: go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest)
# Uses envsubst to expand ${GOARCH} in nfpm.yaml contents field
package-deb-amd64: build-linux-amd64
	VERSION=$(VERSION) GOARCH=amd64 envsubst < nfpm.yaml | nfpm package --config /dev/stdin --packager deb --target dist/

package-deb-arm64: build-linux-arm64
	VERSION=$(VERSION) GOARCH=arm64 envsubst < nfpm.yaml | nfpm package --config /dev/stdin --packager deb --target dist/

package-rpm-amd64: build-linux-amd64
	VERSION=$(VERSION) GOARCH=amd64 envsubst < nfpm.yaml | nfpm package --config /dev/stdin --packager rpm --target dist/

package-rpm-arm64: build-linux-arm64
	VERSION=$(VERSION) GOARCH=arm64 envsubst < nfpm.yaml | nfpm package --config /dev/stdin --packager rpm --target dist/

package-all: package-deb-amd64 package-deb-arm64 package-rpm-amd64 package-rpm-arm64
