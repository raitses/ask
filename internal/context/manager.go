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
	// Check if we need emergency pruning BEFORE adding messages
	if err := m.checkEmergencyPrune(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Emergency pruning failed: %v\n", err)
	}

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

	// Check if we're way over limits after adding response
	if err := m.checkEmergencyPrune(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Emergency pruning failed: %v\n", err)
	}

	// Check if normal pruning is needed
	if err := m.checkAndPrune(); err != nil {
		// Log warning but don't fail the query
		fmt.Fprintf(os.Stderr, "Warning: Context pruning failed: %v\n", err)
	}

	// Save context
	if err := m.store.Save(); err != nil {
		return "", fmt.Errorf("failed to save context: %w", err)
	}

	return response, nil
}

// checkEmergencyPrune performs aggressive pruning if we're way over limits
func (m *Manager) checkEmergencyPrune() error {
	tokens := m.store.EstimateTokens()
	messages := len(m.store.Messages)

	// Emergency thresholds (150% of hard limits)
	emergencyTokens := 37500  // 1.5 * 25000
	emergencyMessages := 150  // 1.5 * 100

	if tokens > emergencyTokens || messages > emergencyMessages {
		fmt.Fprintf(os.Stderr, "⚠️  Emergency pruning: context way over limits (%d tokens, %d messages)\n",
			tokens, messages)

		// Check if the problem is the analysis cache
		if m.store.AnalysisCache != nil {
			analysisTokens := m.estimateAnalysisCacheTokens()

			// If analysis cache is > 50% of the tokens, it's the problem
			if analysisTokens > tokens/2 {
				fmt.Fprintf(os.Stderr, "⚠️  Analysis cache is the issue (%d of %d tokens) - clearing it\n",
					analysisTokens, tokens)

				// Clear the analysis cache entirely
				m.store.AnalysisCache = nil
				m.store.LastAnalysisAt = nil

				fmt.Fprintf(os.Stderr, "Analysis cache cleared. Tokens reduced from %d to %d\n",
					tokens, m.store.EstimateTokens())

				// Re-check tokens after clearing analysis
				tokens = m.store.EstimateTokens()
			}
		}

		// If still over limits, prune messages
		if tokens > emergencyTokens || messages > emergencyMessages {
			pruner := NewPruner(m.store, m.client)
			if err := pruner.pruneHard(); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Emergency pruning complete: %d messages remain (%d tokens)\n",
				len(m.store.Messages), m.store.EstimateTokens())
		}
	}

	return nil
}

// estimateAnalysisCacheTokens estimates tokens used by analysis cache
func (m *Manager) estimateAnalysisCacheTokens() int {
	if m.store.AnalysisCache == nil {
		return 0
	}

	tokens := 0
	// File tree tokens
	tokens += int(float64(len(m.store.AnalysisCache.FileTree)) / 3.5)
	// README tokens
	tokens += int(float64(len(m.store.AnalysisCache.ReadmeContent)) / 3.5)
	// Config list overhead
	tokens += len(m.store.AnalysisCache.PrimaryConfigs) * 2

	return tokens
}

// checkAndPrune checks if pruning is needed and performs it
func (m *Manager) checkAndPrune() error {
	pruner := NewPruner(m.store, m.client)

	shouldPrune, reason := pruner.ShouldPrune()
	if !shouldPrune {
		return nil
	}

	fmt.Fprintf(os.Stderr, "Context pruning triggered: %s\n", reason)

	if err := pruner.Prune(); err != nil {
		return fmt.Errorf("pruning failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Context pruned: %d messages remain (%d tokens estimated)\n",
		len(m.store.Messages), m.store.EstimateTokens())

	return nil
}

// Reset clears the conversation context
func (m *Manager) Reset() error {
	m.store.Reset()
	if err := m.store.Save(); err != nil {
		return fmt.Errorf("failed to save reset context: %w", err)
	}
	return nil
}

// Analyze performs directory analysis and caches the results
func (m *Manager) Analyze() error {
	if err := AnalyzeDirectory(m.store); err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if err := m.store.Save(); err != nil {
		return fmt.Errorf("failed to save analysis: %w", err)
	}

	return nil
}

// GetInfo returns information about the current context
func (m *Manager) GetInfo() string {
	info := fmt.Sprintf("Context for %s\n", m.store.Directory)
	info += fmt.Sprintf("Messages: %d\n", m.store.Metadata.TotalMessages)
	info += fmt.Sprintf("Estimated tokens: %d\n", m.store.Metadata.TotalTokensEstimate)
	info += fmt.Sprintf("Prune count: %d\n", m.store.Metadata.PruneCount)

	if m.store.LastAnalysisAt != nil {
		info += fmt.Sprintf("Last analysis: %s\n", m.store.LastAnalysisAt.Format("2006-01-02 15:04:05"))
	}

	info += fmt.Sprintf("Last updated: %s\n", m.store.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Show pruning status
	pruner := NewPruner(m.store, m.client)
	if shouldPrune, reason := pruner.ShouldPrune(); shouldPrune {
		info += fmt.Sprintf("\n⚠️  Pruning will be triggered soon: %s\n", reason)
	}

	return info
}
