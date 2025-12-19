package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/raitses/ask/internal/config"
	"github.com/raitses/ask/pkg/hash"
)

// Message represents a single message in the conversation
type Message struct {
	Role      string    `json:"role"`      // system, user, assistant
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// AnalysisCache holds cached directory analysis results
type AnalysisCache struct {
	FileTree       string   `json:"file_tree"`
	ReadmeContent  string   `json:"readme_content,omitempty"`
	PrimaryConfigs []string `json:"primary_configs"`
}

// Metadata holds statistics about the conversation
type Metadata struct {
	TotalMessages       int `json:"total_messages"`
	TotalTokensEstimate int `json:"total_tokens_estimate"`
	PruneCount          int `json:"prune_count"`
}

// Store represents the persistent conversation context for a directory
type Store struct {
	Version        string         `json:"version"`
	Directory      string         `json:"directory"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	LastAnalysisAt *time.Time     `json:"last_analysis_at,omitempty"`
	AnalysisCache  *AnalysisCache `json:"analysis_cache,omitempty"`
	Messages       []Message      `json:"messages"`
	Metadata       Metadata       `json:"metadata"`
}

// NewStore creates a new context store for the given directory
func NewStore(directory string) *Store {
	now := time.Now()
	return &Store{
		Version:   "1",
		Directory: directory,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []Message{},
		Metadata: Metadata{
			TotalMessages:       0,
			TotalTokensEstimate: 0,
			PruneCount:          0,
		},
	}
}

// Load reads the context store from disk
func Load(directory string) (*Store, error) {
	path := getContextFilePath(directory)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewStore(directory), nil
		}
		return nil, fmt.Errorf("failed to read context file: %w", err)
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse context file: %w", err)
	}

	// Verify directory matches
	if store.Directory != directory {
		return nil, fmt.Errorf("context file directory mismatch: expected %s, got %s", directory, store.Directory)
	}

	return &store, nil
}

// Save writes the context store to disk
func (s *Store) Save() error {
	s.UpdatedAt = time.Now()

	// Ensure context directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	contextDir := filepath.Join(homeDir, config.ContextDir)
	if err := os.MkdirAll(contextDir, 0700); err != nil {
		return fmt.Errorf("failed to create context directory: %w", err)
	}

	path := getContextFilePath(s.Directory)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	return nil
}

// AddMessage adds a new message to the conversation
func (s *Store) AddMessage(role, content string) {
	msg := Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.Messages = append(s.Messages, msg)
	s.Metadata.TotalMessages = len(s.Messages)
	s.Metadata.TotalTokensEstimate = s.EstimateTokens()
}

// EstimateTokens provides a rough estimate of token count (4 chars ≈ 1 token)
func (s *Store) EstimateTokens() int {
	total := 0
	for _, msg := range s.Messages {
		total += len(msg.Content) / 4
	}
	return total
}

// Reset clears all messages and analysis cache
func (s *Store) Reset() {
	s.Messages = []Message{}
	s.AnalysisCache = nil
	s.LastAnalysisAt = nil
	s.Metadata = Metadata{
		TotalMessages:       0,
		TotalTokensEstimate: 0,
		PruneCount:          s.Metadata.PruneCount, // Preserve prune count
	}
}

// getContextFilePath returns the path to the context file for a directory
func getContextFilePath(directory string) string {
	homeDir, _ := os.UserHomeDir()
	dirHash := hash.DirectoryPath(directory)
	return filepath.Join(homeDir, config.ContextDir, dirHash+".json")
}
