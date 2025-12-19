package context

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/raitses/ask/internal/api"
	"github.com/raitses/ask/internal/config"
	"github.com/raitses/ask/internal/prompt"
)

// Manager handles context operations
type Manager struct {
	store  *Store
	config *config.Config
	client *api.Client
}

// NewManager creates a new context manager for the current directory
func NewManager(cfg *config.Config) (*Manager, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	absPath, err := filepath.Abs(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	store, err := Load(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load context: %w", err)
	}

	client := api.NewClient(cfg)

	return &Manager{
		store:  store,
		config: cfg,
		client: client,
	}, nil
}

// Query sends a query to the LLM with conversation context
func (m *Manager) Query(userQuery string) (string, error) {
	// Add user message to context
	m.store.AddMessage("user", userQuery)

	// Convert store messages to prompt messages
	promptMessages := make([]prompt.Message, len(m.store.Messages))
	for i, msg := range m.store.Messages {
		promptMessages[i] = prompt.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Convert analysis cache if present
	var analysis *prompt.AnalysisCache
	if m.store.AnalysisCache != nil {
		analysis = &prompt.AnalysisCache{
			FileTree:       m.store.AnalysisCache.FileTree,
			ReadmeContent:  m.store.AnalysisCache.ReadmeContent,
			PrimaryConfigs: m.store.AnalysisCache.PrimaryConfigs,
		}
	}

	// Build messages for API
	messages := prompt.BuildMessages(m.store.Directory, m.config.OS, promptMessages, analysis)

	// Get response from API
	response, err := m.client.ChatCompletion(messages)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	// Add assistant response to context
	m.store.AddMessage("assistant", response)

	// Save context
	if err := m.store.Save(); err != nil {
		return "", fmt.Errorf("failed to save context: %w", err)
	}

	return response, nil
}

// Reset clears the conversation context
func (m *Manager) Reset() error {
	m.store.Reset()
	if err := m.store.Save(); err != nil {
		return fmt.Errorf("failed to save reset context: %w", err)
	}
	return nil
}

// GetInfo returns information about the current context
func (m *Manager) GetInfo() string {
	info := fmt.Sprintf("Context for %s\n", m.store.Directory)
	info += fmt.Sprintf("Messages: %d\n", m.store.Metadata.TotalMessages)
	info += fmt.Sprintf("Estimated tokens: %d\n", m.store.Metadata.TotalTokensEstimate)

	if m.store.LastAnalysisAt != nil {
		info += fmt.Sprintf("Last analysis: %s\n", m.store.LastAnalysisAt.Format("2006-01-02 15:04:05"))
	}

	info += fmt.Sprintf("Last updated: %s\n", m.store.UpdatedAt.Format("2006-01-02 15:04:05"))

	return info
}
