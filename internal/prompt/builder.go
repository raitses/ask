package prompt

import (
	"github.com/raitses/ask/internal/api"
)

// Message represents a simple message structure to avoid import cycles
type Message struct {
	Role    string
	Content string
}

// AnalysisCache represents cached analysis data
type AnalysisCache struct {
	FileTree       string
	ReadmeContent  string
	PrimaryConfigs []string
}

// BuildMessages converts messages to API messages with system prompt
func BuildMessages(directory, osType string, messages []Message, analysis *AnalysisCache, useClaudeCache bool) []api.ChatMessage {
	apiMessages := make([]api.ChatMessage, 0, len(messages)+1)

	// Build system prompt
	systemPrompt := BaseSystemPrompt(osType, directory)

	// Add analysis if available
	if analysis != nil {
		systemPrompt += AnalysisSystemPrompt(
			analysis.FileTree,
			analysis.ReadmeContent,
			analysis.PrimaryConfigs,
		)
	}

	// Add system message with cache control for Claude API
	systemMsg := api.ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	}

	// Mark for caching if using Claude API
	// This caches the entire system prompt + analysis (typically 4,000+ tokens)
	if useClaudeCache {
		systemMsg.CacheControl = &api.CacheControl{Type: "ephemeral"}
	}

	apiMessages = append(apiMessages, systemMsg)

	// Add conversation history (skip old system messages)
	for _, msg := range messages {
		if msg.Role == "system" {
			// Skip old system messages - we built a fresh one
			continue
		}
		apiMessages = append(apiMessages, api.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return apiMessages
}
