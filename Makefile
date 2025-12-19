.PHONY: build install clean test build-all release-test

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Build the binary
build:
	nix-shell --run "go build -ldflags '$(LDFLAGS)' -o ask ./cmd/ask"

# Install to /usr/local/bin
install: build
	sudo cp ask /usr/local/bin/

# Clean build artifacts
clean:
	rm -f ask ask-* dist/

# Run tests
test:
	nix-shell --run "go test -v ./..."

# Run tests with race detector
test-race:
	nix-shell --run "go test -race -v ./..."

# Build for multiple platforms
build-all:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/ask-darwin-amd64 ./cmd/ask
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/ask-darwin-arm64 ./cmd/ask
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/ask-linux-amd64 ./cmd/ask
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/ask-linux-arm64 ./cmd/ask
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/ask-windows-amd64.exe ./cmd/ask

# Test goreleaser locally
release-test:
	goreleaser release --snapshot --clean --skip=publish

# Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"
