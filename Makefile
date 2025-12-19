.PHONY: build install clean test

# Build the binary
build:
	nix-shell --run "go build -o ask ./cmd/ask"

# Install to /usr/local/bin
install: build
	sudo cp ask /usr/local/bin/

# Clean build artifacts
clean:
	rm -f ask

# Run tests
test:
	nix-shell --run "go test -v ./..."

# Build for multiple platforms
build-all:
	nix-shell --run "GOOS=darwin GOARCH=amd64 go build -o ask-darwin-amd64 ./cmd/ask"
	nix-shell --run "GOOS=darwin GOARCH=arm64 go build -o ask-darwin-arm64 ./cmd/ask"
	nix-shell --run "GOOS=linux GOARCH=amd64 go build -o ask-linux-amd64 ./cmd/ask"
	nix-shell --run "GOOS=windows GOARCH=amd64 go build -o ask-windows-amd64.exe ./cmd/ask"
