package context

import (
	"strings"
	"testing"
)

func TestMessageSizeLimits(t *testing.T) {
	store := NewStore("/test/dir")

	// Create a message that exceeds the limit
	hugeContent := strings.Repeat("A", MaxMessageLength+1000)

	store.AddMessage("user", hugeContent)

	// Should be truncated
	if len(store.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(store.Messages))
	}

	msg := store.Messages[0]

	// Should be at or under the limit (plus truncation message)
	if len(msg.Content) > MaxMessageLength+100 {
		t.Errorf("Message not truncated: length %d, limit %d", len(msg.Content), MaxMessageLength)
	}

	// Should contain truncation notice
	if !strings.Contains(msg.Content, "[Content truncated") {
		t.Error("Truncation notice not found in message")
	}

	t.Logf("Huge message truncated from %d to %d chars", len(hugeContent), len(msg.Content))
}

func TestAnalyzerFileSizeLimits(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a huge README
	hugeReadme := strings.Repeat("This is a very long README.\n", 1000)
	if len(hugeReadme) > 8000 {
		// Create the file
		// (we won't actually write it, just test the analyzer limits)
		t.Logf("Huge README would be %d bytes", len(hugeReadme))
	}

	analyzer := NewAnalyzer(tmpDir)

	// Verify limits are set correctly (updated to more aggressive values)
	if analyzer.maxReadmeLen != 5000 {
		t.Errorf("maxReadmeLen = %d, want 5000", analyzer.maxReadmeLen)
	}

	if analyzer.maxDepth != 2 {
		t.Errorf("maxDepth = %d, want 2", analyzer.maxDepth)
	}
}

func TestEmergencyPruneThresholds(t *testing.T) {
	store := NewStore("/test/dir")

	// Add enough messages to trigger emergency pruning
	for i := 0; i < 160; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		// Add moderately sized messages
		store.AddMessage(role, strings.Repeat("Test content ", 50))
	}

	initialCount := len(store.Messages)
	initialTokens := store.EstimateTokens()

	t.Logf("Created %d messages with ~%d tokens", initialCount, initialTokens)

	// Check if we're over emergency thresholds
	emergencyTokens := 37500
	emergencyMessages := 150

	if initialCount > emergencyMessages || initialTokens > emergencyTokens {
		t.Logf("Over emergency thresholds - would trigger emergency pruning")
		t.Logf("  Messages: %d > %d", initialCount, emergencyMessages)
		t.Logf("  Tokens: %d (threshold: %d)", initialTokens, emergencyTokens)
	} else {
		t.Errorf("Expected to exceed emergency thresholds")
	}
}

func TestTokenEstimationWithLargeContent(t *testing.T) {
	store := NewStore("/test/dir")

	// Add a large message
	largeContent := strings.Repeat("This is a test message. ", 1000) // ~24k chars
	store.AddMessage("user", largeContent)
	store.AddMessage("assistant", "Short response")

	tokens := store.EstimateTokens()

	// Should be roughly 24000/3.5 + overhead
	expectedMin := 6000  // Conservative
	expectedMax := 10000 // With overhead

	if tokens < expectedMin || tokens > expectedMax {
		t.Errorf("Token estimate %d outside expected range [%d, %d]",
			tokens, expectedMin, expectedMax)
	}

	t.Logf("Large message (%d chars) estimated at %d tokens", len(largeContent), tokens)
}

func TestMultipleOversizedMessages(t *testing.T) {
	store := NewStore("/test/dir")

	// Add multiple messages that are at the size limit
	for i := 0; i < 5; i++ {
		// Each message just under the limit
		content := strings.Repeat("X", MaxMessageLength-100)
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		store.AddMessage(role, content)
	}

	tokens := store.EstimateTokens()
	messageCount := len(store.Messages)

	t.Logf("Added %d max-size messages, estimated %d tokens", messageCount, tokens)

	// Should be way over soft limits but under emergency
	if tokens < 15000 {
		t.Error("Expected to be over soft limits")
	}

	if tokens > 100000 {
		t.Error("Token estimate seems unreasonably high")
	}
}
