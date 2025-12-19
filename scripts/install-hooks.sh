#!/bin/bash
# Install git hooks for local development

set -e

HOOKS_DIR=".git/hooks"
SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPTS_DIR")"

echo "Installing git hooks..."

# Install pre-push hook
cat > "$HOOKS_DIR/pre-push" << 'EOF'
#!/bin/bash
# Pre-push hook to run linting and tests before pushing

set -e

echo "Running pre-push checks..."

# Check if we have shell.nix, use nix-shell
if [ -f "shell.nix" ]; then
    echo "→ Running tests with nix-shell..."
    nix-shell --run "go test ./..."

    echo "→ Running golangci-lint with nix-shell..."
    if [ -f "$HOME/go/bin/golangci-lint" ]; then
        nix-shell --run "$HOME/go/bin/golangci-lint run"
    else
        echo "⚠️  golangci-lint not found, skipping lint check"
        echo "   Install with: nix-shell --run 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"
    fi
else
    # Run tests
    echo "→ Running tests..."
    go test ./...

    # Run linter
    echo "→ Running golangci-lint..."
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run
    elif [ -f "$HOME/go/bin/golangci-lint" ]; then
        "$HOME/go/bin/golangci-lint" run
    else
        echo "⚠️  golangci-lint not found, skipping lint check"
        echo "   Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    fi
fi

echo "✅ All pre-push checks passed!"
EOF

chmod +x "$HOOKS_DIR/pre-push"

echo "✅ Git hooks installed successfully!"
echo ""
echo "The following hooks are now active:"
echo "  - pre-push: Runs tests and linter before pushing"
echo ""
echo "To install golangci-lint (required for linting):"
echo "  nix-shell --run 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"
