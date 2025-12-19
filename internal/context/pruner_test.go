package context

import (
	"strings"
	"testing"
	"time"
)

func TestPrunerShouldPrune(t *testing.T) {
	tests := []struct {
		name          string
		messageCount  int
		shouldPrune   bool
		reasonContains string
	}{
		{
			name:         "No pruning needed - empty",
			messageCount: 0,
			shouldPrune:  false,
		},
		{
			name:         "No pruning needed - few messages",
			messageCount: 10,
			shouldPrune:  false,
		},
		{
			name:           "Soft limit - messages",
			messageCount:   40,
			shouldPrune:    true,
			reasonContains: "soft limit: messages",
		},
		{
			name:           "Hard limit - messages",
			messageCount:   100,
			shouldPrune:    true,
			reasonContains: "hard limit: messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore("/test/dir")

			// Add messages
			for i := 0; i < tt.messageCount; i++ {
				role := "user"
				if i%2 == 1 {
					role = "assistant"
				}
				store.AddMessage(role, "test message "+string(rune(i)))
			}

			pruner := NewPruner(store, nil)
			shouldPrune, reason := pruner.ShouldPrune()

			if shouldPrune != tt.shouldPrune {
				t.Errorf("ShouldPrune() = %v, want %v", shouldPrune, tt.shouldPrune)
			}

			if tt.shouldPrune && !strings.Contains(reason, tt.reasonContains) {
				t.Errorf("Reason %q should contain %q", reason, tt.reasonContains)
			}
		})
	}
}

func TestPrunerHardPrune(t *testing.T) {
	store := NewStore("/test/dir")

	// Add 50 messages
	for i := 0; i < 50; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		store.AddMessage(role, "Message "+string(rune('A'+i)))
	}

	pruner := NewPruner(store, nil)
	limits := DefaultPruningLimits()

	if err := pruner.pruneHard(); err != nil {
		t.Fatalf("pruneHard() failed: %v", err)
	}

	// Should be pruned to target
	if len(store.Messages) > limits.TargetMessages {
		t.Errorf("After pruning: got %d messages, want <= %d", len(store.Messages), limits.TargetMessages)
	}

	// Should preserve recent messages
	if len(store.Messages) < 4 {
		t.Errorf("Should preserve at least 4 messages, got %d", len(store.Messages))
	}

	// Prune count should be incremented
	if store.Metadata.PruneCount != 1 {
		t.Errorf("PruneCount = %d, want 1", store.Metadata.PruneCount)
	}

	t.Logf("Pruned from 50 to %d messages", len(store.Messages))
}

func TestPrunerPreservation(t *testing.T) {
	store := NewStore("/test/dir")

	// Add various types of messages
	store.AddMessage("user", "Simple question")
	store.AddMessage("assistant", "Simple answer")
	store.AddMessage("user", "Question with code")
	store.AddMessage("assistant", "Here's code:\n```go\nfunc main() {}\n```")
	store.AddMessage("user", "Question about structure")
	store.AddMessage("assistant", "The project architecture includes...")
	store.AddMessage("user", "Recent question 1")
	store.AddMessage("assistant", "Recent answer 1")
	store.AddMessage("user", "Recent question 2")
	store.AddMessage("assistant", "Recent answer 2")

	pruner := NewPruner(store, nil)

	tests := []struct {
		index    int
		preserve bool
		reason   string
	}{
		{0, false, "old simple message"},
		{1, false, "old simple message"},
		{2, false, "old message despite being about code"},
		{3, true, "contains code block"},
		{4, true, "mentions structure (keyword)"},
		{5, true, "mentions architecture (keyword)"},
		{6, true, "recent message (last 4)"},
		{7, true, "recent message (last 4)"},
		{8, true, "recent message (last 4)"},
		{9, true, "recent message (last 4)"},
	}

	for _, tt := range tests {
		if tt.index >= len(store.Messages) {
			continue
		}

		msg := store.Messages[tt.index]
		preserve := pruner.ShouldPreserve(msg, tt.index)

		if preserve != tt.preserve {
			t.Errorf("Index %d (%s): ShouldPreserve() = %v, want %v",
				tt.index, tt.reason, preserve, tt.preserve)
		}
	}
}

func TestPrunerRemoveByIndices(t *testing.T) {
	store := NewStore("/test/dir")

	// Add 10 messages
	for i := 0; i < 10; i++ {
		store.AddMessage("user", string(rune('A'+i)))
	}

	pruner := NewPruner(store, nil)

	// Remove indices 0, 2, 4, 6, 8 (every other message)
	pruner.removeMessagesByIndices([]int{0, 2, 4, 6, 8})

	// Should have 5 messages remaining
	if len(store.Messages) != 5 {
		t.Errorf("After removal: got %d messages, want 5", len(store.Messages))
	}

	// Check remaining messages are correct
	expected := []string{"B", "D", "F", "H", "J"}
	for i, msg := range store.Messages {
		if msg.Content != expected[i] {
			t.Errorf("Message %d: got %q, want %q", i, msg.Content, expected[i])
		}
	}
}

func TestPrunerAgeLimit(t *testing.T) {
	store := NewStore("/test/dir")

	// Add an old message
	oldMsg := Message{
		Role:      "user",
		Content:   "Old message",
		Timestamp: time.Now().Add(-35 * 24 * time.Hour), // 35 days ago
	}
	store.Messages = append(store.Messages, oldMsg)

	pruner := NewPruner(store, nil)
	shouldPrune, reason := pruner.ShouldPrune()

	if !shouldPrune {
		t.Error("Should prune due to age limit")
	}

	if !strings.Contains(reason, "age") {
		t.Errorf("Reason should mention age, got: %s", reason)
	}
}

func TestPrunerParsePruningResponse(t *testing.T) {
	store := NewStore("/test/dir")
	pruner := NewPruner(store, nil)

	tests := []struct {
		name     string
		response string
		want     []int
		wantErr  bool
	}{
		{
			name:     "Simple array",
			response: "[0, 1, 4, 5]",
			want:     []int{0, 1, 4, 5},
			wantErr:  false,
		},
		{
			name:     "With markdown",
			response: "```json\n[2, 3, 6]\n```",
			want:     []int{2, 3, 6},
			wantErr:  false,
		},
		{
			name:     "With whitespace",
			response: "  [1, 2, 3]  ",
			want:     []int{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "Empty array",
			response: "[]",
			want:     []int{},
			wantErr:  false,
		},
		{
			name:     "Invalid JSON",
			response: "not json",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pruner.parsePruningResponse(tt.response)

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePruningResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("parsePruningResponse() length = %d, want %d", len(got), len(tt.want))
					return
				}

				for i, v := range got {
					if v != tt.want[i] {
						t.Errorf("parsePruningResponse()[%d] = %d, want %d", i, v, tt.want[i])
					}
				}
			}
		})
	}
}

func TestTokenEstimation(t *testing.T) {
	store := NewStore("/test/dir")

	// Add some messages
	store.AddMessage("user", "What is this project?")
	store.AddMessage("assistant", "This is a Go-based CLI tool for conversational AI assistance.")

	tokens := store.EstimateTokens()

	// Should have tokens for:
	// - 2 messages with content
	// - Message overhead
	// - Base system prompt
	if tokens < 50 {
		t.Errorf("Token estimate seems too low: %d", tokens)
	}

	if tokens > 200 {
		t.Errorf("Token estimate seems too high: %d", tokens)
	}

	t.Logf("Estimated tokens for 2 messages: %d", tokens)

	// Add analysis cache
	store.AnalysisCache = &AnalysisCache{
		FileTree:       strings.Repeat("test/\n  file.go\n", 20),
		ReadmeContent:  strings.Repeat("Test content\n", 50),
		PrimaryConfigs: []string{"go.mod", "Makefile"},
	}

	tokensWithAnalysis := store.EstimateTokens()

	// Should be significantly more with analysis
	if tokensWithAnalysis <= tokens {
		t.Errorf("Tokens with analysis (%d) should be > tokens without (%d)",
			tokensWithAnalysis, tokens)
	}

	t.Logf("Estimated tokens with analysis: %d", tokensWithAnalysis)
}
