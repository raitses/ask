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

### Prerequisites

- Nix package manager (for building from source)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/raitses/ask.git
cd ask

# Build
make build

# Install system-wide
make install
```

### Download Binary

(Coming soon - releases will provide pre-built binaries)

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
ask how do I run tests in this project?

ask what's the difference between these two functions?

ask how do I implement authentication?
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
ask --analyze what's the architecture of this codebase?
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

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Build for Multiple Platforms

```bash
make build-all
```

## Context Management

The tool automatically manages conversation context to keep token usage reasonable:

- **Soft Limits**: Pruning triggered at 40 messages or 15,000 tokens
- **Hard Limits**: Maximum 100 messages, 25,000 tokens, or 30 days old
- **AI-Driven Pruning**: When soft limits are reached, AI intelligently selects which exchanges to remove
- **Preservation Rules**: Always keeps recent exchanges, code examples, and important context
- **Fallback**: If AI pruning fails, simple FIFO pruning is used

You can check context status with:
```bash
ask --info
```

## Roadmap

- [x] Phase 1: Core MVP (context persistence, basic queries)
- [x] Phase 2: Directory analysis (`--analyze` flag)
- [x] Phase 3: AI-driven context pruning
- [ ] Phase 4: Multi-platform releases and CI/CD

## Contributing

Contributions welcome! Open an issue or pull request.

## License

MIT License - see LICENSE file for details.
