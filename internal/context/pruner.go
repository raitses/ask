package context

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/raitses/ask/internal/api"
)

// PruningLimits defines the thresholds for context pruning
type PruningLimits struct {
	// Hard limits (automatic pruning)
	MaxMessages      int
	MaxTokens        int
	MaxAgeDays       int

	// Soft limits (trigger AI-driven pruning)
	SoftMaxMessages  int
	SoftMaxTokens    int

	// Target after pruning
	TargetMessages   int
	TargetTokens     int
}

// DefaultPruningLimits returns the default pruning configuration
func DefaultPruningLimits() PruningLimits {
	return PruningLimits{
		MaxMessages:     100,  // 50 exchanges
		MaxTokens:       25000,
		MaxAgeDays:      30,
		SoftMaxMessages: 40,   // 20 exchanges
		SoftMaxTokens:   15000,
		TargetMessages:  24,   // 12 exchanges
		TargetTokens:    10000,
	}
}

// Pruner handles context pruning operations
type Pruner struct {
	store  *Store
	client *api.Client
	limits PruningLimits
}

// NewPruner creates a new context pruner
func NewPruner(store *Store, client *api.Client) *Pruner {
	return &Pruner{
		store:  store,
		client: client,
		limits: DefaultPruningLimits(),
	}
}

// ShouldPrune checks if pruning is needed based on current context
func (p *Pruner) ShouldPrune() (bool, string) {
	// Check hard limits first
	if len(p.store.Messages) >= p.limits.MaxMessages {
		return true, fmt.Sprintf("hard limit: messages (%d >= %d)", len(p.store.Messages), p.limits.MaxMessages)
	}

	tokens := p.store.EstimateTokens()
	if tokens >= p.limits.MaxTokens {
		return true, fmt.Sprintf("hard limit: tokens (%d >= %d)", tokens, p.limits.MaxTokens)
	}

	// Check age of oldest message
	if len(p.store.Messages) > 0 {
		oldest := p.store.Messages[0].Timestamp
		age := time.Since(oldest)
		if age > time.Duration(p.limits.MaxAgeDays)*24*time.Hour {
			return true, fmt.Sprintf("hard limit: age (%.0f days >= %d days)", age.Hours()/24, p.limits.MaxAgeDays)
		}
	}

	// Check soft limits
	if len(p.store.Messages) >= p.limits.SoftMaxMessages {
		return true, fmt.Sprintf("soft limit: messages (%d >= %d)", len(p.store.Messages), p.limits.SoftMaxMessages)
	}

	if tokens >= p.limits.SoftMaxTokens {
		return true, fmt.Sprintf("soft limit: tokens (%d >= %d)", tokens, p.limits.SoftMaxTokens)
	}

	return false, ""
}

// Prune performs context pruning using AI-driven selection when possible
func (p *Pruner) Prune() error {
	shouldPrune, reason := p.ShouldPrune()
	if !shouldPrune {
		return nil // No pruning needed
	}

	// Check if we can use AI-driven pruning
	if p.client != nil && p.canUseAIPruning() {
		if err := p.pruneWithAI(reason); err != nil {
			// Fall back to hard pruning if AI pruning fails
			return p.pruneHard()
		}
		return nil
	}

	// Use hard pruning as fallback
	return p.pruneHard()
}

// canUseAIPruning checks if conditions are met for AI-driven pruning
func (p *Pruner) canUseAIPruning() bool {
	// Need at least 10 messages to make AI pruning worthwhile
	if len(p.store.Messages) < 10 {
		return false
	}

	// Don't use AI if we're way over hard limits (just cut)
	if len(p.store.Messages) >= p.limits.MaxMessages {
		return false
	}

	tokens := p.store.EstimateTokens()
	if tokens >= p.limits.MaxTokens {
		return false
	}

	return true
}

// pruneWithAI uses AI to intelligently select which messages to remove
func (p *Pruner) pruneWithAI(reason string) error {
	// Build pruning request
	prompt := p.buildPruningPrompt(reason)

	messages := []api.ChatMessage{
		{
			Role:    "system",
			Content: prompt,
		},
	}

	// Get AI's pruning suggestions
	response, err := p.client.ChatCompletion(messages)
	if err != nil {
		return fmt.Errorf("AI pruning request failed: %w", err)
	}

	// Parse the response (expecting JSON array of indices)
	indices, err := p.parsePruningResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse pruning response: %w", err)
	}

	// Apply the pruning
	if len(indices) > 0 {
		p.removeMessagesByIndices(indices)
		p.store.Metadata.PruneCount++
		p.store.Metadata.TotalMessages = len(p.store.Messages)
		p.store.Metadata.TotalTokensEstimate = p.store.EstimateTokens()
	}

	return nil
}

// buildPruningPrompt creates the prompt for AI-driven pruning
func (p *Pruner) buildPruningPrompt(reason string) string {
	tokens := p.store.EstimateTokens()

	// Build conversation summary
	summary := strings.Builder{}
	summary.WriteString("CONVERSATION MESSAGES:\n\n")

	for i, msg := range p.store.Messages {
		// Skip system messages in the list
		if msg.Role == "system" {
			continue
		}

		// Truncate long messages for the summary
		content := msg.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}

		summary.WriteString(fmt.Sprintf("[%d] %s: %s\n\n", i, msg.Role, content))
	}

	return fmt.Sprintf(`You are helping manage a conversation context that has grown too large.

CONTEXT PRUNING REQUIRED:
Reason: %s

Current state:
- Total messages: %d
- Estimated tokens: %d
- Target: Reduce to ~%d tokens (%d messages)

%s

Your task: Analyze the conversation and identify exchanges (user question + assistant response pairs) that are:
1. Least relevant to ongoing work
2. One-off questions that were fully resolved
3. Outdated information that's been superseded
4. Redundant or repetitive

IMPORTANT RULES:
- Always preserve the last 4 messages (most recent 2 exchanges)
- Preserve messages containing code examples (with triple backticks)
- Preserve messages that reference project structure or analysis results
- Return ONLY a JSON array of message indices to remove

Example response format:
[0, 1, 4, 5, 8, 9]

Respond with ONLY the JSON array, no other text.`,
		reason,
		len(p.store.Messages),
		tokens,
		p.limits.TargetTokens,
		p.limits.TargetMessages,
		summary.String())
}

// parsePruningResponse extracts message indices from AI response
func (p *Pruner) parsePruningResponse(response string) ([]int, error) {
	// Clean up response (remove markdown code blocks if present)
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var indices []int
	if err := json.Unmarshal([]byte(response), &indices); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array: %w", err)
	}

	return indices, nil
}

// removeMessagesByIndices removes messages at the specified indices
func (p *Pruner) removeMessagesByIndices(indices []int) {
	// Create a set of indices to remove for O(1) lookup
	toRemove := make(map[int]bool)
	for _, idx := range indices {
		toRemove[idx] = true
	}

	// Build new message list excluding removed indices
	newMessages := make([]Message, 0, len(p.store.Messages)-len(indices))
	for i, msg := range p.store.Messages {
		if !toRemove[i] {
			newMessages = append(newMessages, msg)
		}
	}

	p.store.Messages = newMessages
}

// pruneHard performs simple hard pruning by removing oldest messages
func (p *Pruner) pruneHard() error {
	if len(p.store.Messages) <= p.limits.TargetMessages {
		return nil // Already below target
	}

	// Calculate how many to remove
	toRemove := len(p.store.Messages) - p.limits.TargetMessages

	// Apply preservation rules: keep last 4 messages minimum
	if toRemove >= len(p.store.Messages)-4 {
		toRemove = len(p.store.Messages) - 4
	}

	if toRemove <= 0 {
		return nil
	}

	// Remove oldest messages while preserving system messages and recent exchanges
	preserved := make([]Message, 0, p.limits.TargetMessages)

	// Skip old system messages
	startIdx := 0
	for startIdx < len(p.store.Messages) && p.store.Messages[startIdx].Role == "system" {
		startIdx++
	}

	// Keep messages after removing 'toRemove' count
	preserved = append(preserved, p.store.Messages[startIdx+toRemove:]...)

	p.store.Messages = preserved
	p.store.Metadata.PruneCount++
	p.store.Metadata.TotalMessages = len(p.store.Messages)
	p.store.Metadata.TotalTokensEstimate = p.store.EstimateTokens()

	return nil
}

// ShouldPreserve checks if a message should be preserved during pruning
func (p *Pruner) ShouldPreserve(msg Message, index int) bool {
	// Preserve recent messages (last 4)
	if index >= len(p.store.Messages)-4 {
		return true
	}

	// Preserve messages with code blocks
	if strings.Contains(msg.Content, "```") {
		return true
	}

	// Preserve messages that mention analysis or project structure
	keywords := []string{"analysis", "file tree", "README", "structure", "architecture"}
	content := strings.ToLower(msg.Content)
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}

	return false
}
