package config

const (
	// DefaultModel is the default LLM model to use
	DefaultModel = "gpt-4o"

	// DefaultOS is the default operating system context
	DefaultOS = "macOS"

	// DefaultAPIURL is the default OpenAI API endpoint
	DefaultAPIURL = "https://api.openai.com/v1/chat/completions"

	// ContextDir is the directory where context files are stored
	ContextDir = ".config/ask/contexts"

	// GlobalConfigDir is the directory for global configuration
	GlobalConfigDir = ".config/ask"

	// GlobalEnvFile is the filename for global environment config
	GlobalEnvFile = ".env"

	// LocalEnvFile is the filename for local environment config
	LocalEnvFile = ".env"
)
