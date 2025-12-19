package prompt

import (
	"strings"
	"testing"
)

func TestBuildMessagesWithoutCache(t *testing.T) {
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	apiMessages := BuildMessages("/test/dir", "macOS", messages, nil, false)

	// Should have system + 2 messages
	if len(apiMessages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(apiMessages))
	}

	// System message should NOT have cache control
	if apiMessages[0].CacheControl != nil {
		t.Error("System message should not have cache control when disabled")
	}
}

func TestBuildMessagesWithCache(t *testing.T) {
	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	apiMessages := BuildMessages("/test/dir", "macOS", messages, nil, true)

	// Should have system + 1 message
	if len(apiMessages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(apiMessages))
	}

	// System message SHOULD have cache control
	if apiMessages[0].CacheControl == nil {
		t.Error("System message should have cache control when enabled")
	}

	if apiMessages[0].CacheControl.Type != "ephemeral" {
		t.Errorf("Cache control type = %s, want ephemeral", apiMessages[0].CacheControl.Type)
	}
}

func TestBuildMessagesWithAnalysisAndCache(t *testing.T) {
	analysis := &AnalysisCache{
		FileTree:       "test tree",
		ReadmeContent:  "test readme",
		PrimaryConfigs: []string{"go.mod"},
	}

	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	apiMessages := BuildMessages("/test/dir", "macOS", messages, analysis, true)

	// System message should contain analysis AND have cache control
	systemMsg := apiMessages[0]

	if systemMsg.CacheControl == nil {
		t.Error("System message with analysis should have cache control")
	}

	// Verify analysis is included in content
	if !strings.Contains(systemMsg.Content, "PROJECT ANALYSIS") {
		t.Error("System message should include analysis content")
	}
}

func TestCompressedSystemPrompt(t *testing.T) {
	prompt := BaseSystemPrompt("macOS", "/test/dir")

	// Should be shorter than original (~680+ chars before compression)
	// Compressed version is ~630 chars, significant reduction
	if len(prompt) > 650 {
		t.Errorf("Compressed prompt too long: %d chars (expected <650)", len(prompt))
	}

	// Should still contain key instructions
	requiredPhrases := []string{
		"AI assistant",
		"ask --analyze",
		"directory:",
		"No markdown",
		"Concise",
		"context window",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(prompt, phrase) {
			t.Errorf("Compressed prompt missing required phrase: %s", phrase)
		}
	}
}
