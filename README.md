# ask

A conversational CLI tool powered by AI that maintains context across queries. Get intelligent, project-aware assistance directly from your terminal.

## Features

- 🧠 Context-aware conversations per directory
- 💬 Maintains conversation history with intelligent pruning
- 🚀 Fast, zero-dependency Go binary
- 🔧 Configurable for different LLM providers
- 🌍 Cross-platform (macOS, Linux, Windows)
- 📊 Directory analysis support (file tree, README, config detection)
- ✂️ AI-driven context pruning to manage token usage

## Installation

### Quick Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/raitses/ask/main/install.sh | bash
```

This will automatically:
- Detect your platform (macOS, Linux, Windows)
- Download the latest release
- Install to `/usr/local/bin`

### Download Pre-built Binary

Download the appropriate binary for your platform from the [latest release](https://github.com/raitses/ask/releases/latest):

**macOS:**
```bash
# Apple Silicon (M1/M2/M3)
curl -L -o ask.tar.gz https://github.com/raitses/ask/releases/latest/download/ask_VERSION_darwin_arm64.tar.gz
tar -xzf ask.tar.gz
sudo mv ask /usr/local/bin/

# Intel
curl -L -o ask.tar.gz https://github.com/raitses/ask/releases/latest/download/ask_VERSION_darwin_amd64.tar.gz
tar -xzf ask.tar.gz
sudo mv ask /usr/local/bin/
```

**Linux:**
```bash
# x86_64
curl -L -o ask.tar.gz https://github.com/raitses/ask/releases/latest/download/ask_VERSION_linux_amd64.tar.gz
tar -xzf ask.tar.gz
sudo mv ask /usr/local/bin/

# ARM64
curl -L -o ask.tar.gz https://github.com/raitses/ask/releases/latest/download/ask_VERSION_linux_arm64.tar.gz
tar -xzf ask.tar.gz
sudo mv ask /usr/local/bin/
```

**Windows:**

Download the `.zip` file from the releases page and extract to a directory in your PATH.

### Build from Source

Prerequisites: Go 1.24+ or Nix

```bash
# Clone the repository
git clone https://github.com/raitses/ask.git
cd ask

# Build
make build

# Install system-wide
make install
```

## Configuration

Configure `ask` using either environment variables or a `.env` file.

### Option 1: Using a `.env` file (Recommended)

Create a `.env` file in one of these locations:
- Current directory: `./.env` (checked second)
- Config directory: `~/.config/ask/.env` (global config)

**Example `.env` file:**
```bash
ASK_API_KEY=your-api-key
ASK_MODEL=gpt-4o
ASK_OS=macOS
ASK_API_URL=https://api.openai.com/v1/chat/completions
```

Get an API key from [platform.openai.com/api-keys](https://platform.openai.com/api-keys)

### Option 2: Using environment variables

Add these to your shell profile (`.bashrc`, `.zshrc`, etc.):

```bash
export ASK_API_KEY="your-api-key"
export ASK_MODEL="gpt-4o"
export ASK_OS="macOS"
export ASK_API_URL="https://api.openai.com/v1/chat/completions"
```

**Note:** Environment variables take precedence over `.env` file values.

### Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `ASK_API_KEY` | _(none)_ | API key (required for OpenAI) |
| `ASK_MODEL` | `gpt-4o` | Model to use |
| `ASK_OS` | `macOS` | Operating system context |
| `ASK_API_URL` | `https://api.openai.com/v1/chat/completions` | API endpoint |

## Usage

### Basic Queries

Simply type `ask` followed by your question:

```bash
ask how do I run tests in this project

ask "what's the difference between these two functions"

ask how do I implement authentication
```

**Note:** If your query contains special shell characters (`?`, `!`, `'`, etc.), wrap it in quotes:
```bash
ask "how does this work?"
ask 'what'\''s the best approach?'
```

### Context Management

View context information:
```bash
ask --info
```

Reset conversation for current directory:
```bash
ask --reset
```

### Directory Analysis

Analyze project structure before asking:
```bash
ask --analyze "what's the architecture of this codebase"

# Or without special characters
ask --analyze what is the project structure
```

The analysis includes:
- File tree (respecting .gitignore)
- README content
- Detected configuration files (go.mod, package.json, etc.)
- Results are cached and included in the AI's context

## How It Works

1. **Per-Directory Context**: Each directory gets its own conversation context stored in `~/.config/ask/contexts/`
2. **Stateful Conversations**: Previous questions and answers inform future responses
3. **Smart Prompts**: The AI knows it's in a CLI tool and can suggest using `--analyze` when needed
4. **Automatic Persistence**: All conversations are automatically saved and restored
5. **Intelligent Pruning**: When conversations grow too large, AI-driven pruning automatically removes less relevant exchanges while preserving:
   - Recent messages (last 2 exchanges)
   - Code examples
   - Project analysis results
   - Architecture discussions

## Cost Considerations

Using OpenAI's API has costs:
- **GPT-4o**: ~$0.0004 per query (~2,500 queries per $1)
- **GPT-3.5-turbo**: Much cheaper, ~$0.00005 per query (~20,000 queries per $1)

To use a cheaper model:
```bash
export ASK_MODEL="gpt-3.5-turbo"
```

## Development

### Prerequisites

- Go 1.24+
- (Optional) Nix for development environment

### Building

```bash
make build
```

### Testing

```bash
# Run tests
make test

# Run tests with race detector
make test-race
```

### Build for Multiple Platforms

```bash
# Build for all platforms
make build-all

# Test goreleaser locally
make release-test
```

### Making a Release

1. Tag the commit:
   ```bash
   git tag -a v0.4.0 -m "Release v0.4.0"
   git push origin v0.4.0
   ```

2. GitHub Actions will automatically:
   - Run tests
   - Build binaries for all platforms
   - Create a GitHub release
   - Upload release artifacts

## Context Management

The tool automatically manages conversation context to keep token usage reasonable:

### Pruning Limits
- **Soft Limits**: Pruning triggered at 40 messages or 15,000 tokens
- **Hard Limits**: Maximum 100 messages, 25,000 tokens, or 30 days old
- **Emergency Limits**: Aggressive pruning at 150 messages or 37,500 tokens
- **AI-Driven Pruning**: When soft limits are reached, AI intelligently selects which exchanges to remove
- **Preservation Rules**: Always keeps recent exchanges, code examples, and important context
- **Fallback**: If AI pruning fails, simple FIFO pruning is used

### Content Size Safeguards
To prevent single messages from blowing past context limits:
- **Message Limit**: Individual messages capped at 50,000 chars (~14k tokens)
- **README Limit**: README content limited to 5KB
- **File Tree Limit**: Directory tree limited to 10KB
- **Directory Depth**: Analysis descends maximum 2 levels
- **Auto-truncation**: Oversized content automatically truncated with warnings
- **Analysis Cache Clearing**: If analysis cache exceeds 50% of token limit, it's automatically cleared

### Monitoring
You can check context status with:
```bash
ask --info
```

The tool will warn you if:
- Content is truncated
- Emergency pruning is triggered
- Context is approaching limits

## Roadmap

- [x] Phase 1: Core MVP (context persistence, basic queries)
- [x] Phase 2: Directory analysis (`--analyze` flag)
- [x] Phase 3: AI-driven context pruning
- [x] Phase 4: Multi-platform releases and CI/CD

## Contributing

Contributions welcome! Open an issue or pull request.

## License

MIT License - see LICENSE file for details.
