package context

import (
	"strings"
	"testing"

	"github.com/raitses/ask/internal/config"
)

func TestEmergencyPruneWithHugeAnalysisCache(t *testing.T) {
	store := NewStore("/test/dir")

	// Create a huge analysis cache that would cause token overflow
	hugeFileTree := strings.Repeat("src/\n  file.go\n  utils.go\n", 10000) // ~300KB
	hugeReadme := strings.Repeat("This is documentation.\n", 5000)          // ~120KB

	store.AnalysisCache = &AnalysisCache{
		FileTree:       hugeFileTree,
		ReadmeContent:  hugeReadme,
		PrimaryConfigs: []string{"go.mod", "package.json"},
	}

	initialTokens := store.EstimateTokens()
	t.Logf("Initial tokens with huge cache: %d", initialTokens)

	// Should be way over emergency limits
	if initialTokens < 37500 {
		t.Errorf("Expected to be over emergency limits, got %d tokens", initialTokens)
	}

	// Create a manager to test emergency pruning
	cfg := &config.Config{
		Model:  "test",
		OS:     "macOS",
		APIURL: "http://test",
		APIKey: "test",
	}

	manager := &Manager{
		store:  store,
		config: cfg,
		client: nil, // No API client needed for this test
	}

	// Run emergency prune
	if err := manager.checkEmergencyPrune(); err != nil {
		t.Fatalf("checkEmergencyPrune failed: %v", err)
	}

	finalTokens := store.EstimateTokens()
	t.Logf("Final tokens after emergency prune: %d", finalTokens)

	// Analysis cache should be cleared
	if store.AnalysisCache != nil {
		t.Error("Analysis cache should have been cleared")
	}

	// Should be dramatically reduced
	if finalTokens >= initialTokens/2 {
		t.Errorf("Tokens not reduced enough: %d -> %d", initialTokens, finalTokens)
	}

	// Should be under emergency limits
	if finalTokens > 37500 {
		t.Errorf("Still over emergency limits after pruning: %d tokens", finalTokens)
	}
}

func TestAnalysisCacheTokenEstimation(t *testing.T) {
	store := NewStore("/test/dir")
	cfg := &config.Config{Model: "test", OS: "macOS", APIURL: "http://test", APIKey: "test"}
	manager := &Manager{store: store, config: cfg, client: nil}

	// Create analysis cache
	store.AnalysisCache = &AnalysisCache{
		FileTree:       strings.Repeat("file.go\n", 100),   // ~1KB
		ReadmeContent:  strings.Repeat("docs\n", 200),      // ~1KB
		PrimaryConfigs: []string{"go.mod", "Makefile"},
	}

	analysisTokens := manager.estimateAnalysisCacheTokens()
	totalTokens := store.EstimateTokens()

	t.Logf("Analysis tokens: %d, Total tokens: %d", analysisTokens, totalTokens)

	// Analysis should be a significant portion
	if analysisTokens < 100 {
		t.Errorf("Analysis token estimate seems too low: %d", analysisTokens)
	}

	// But not more than total
	if analysisTokens > totalTokens {
		t.Errorf("Analysis tokens (%d) > total tokens (%d)", analysisTokens, totalTokens)
	}
}

func TestEmergencyPruneWithMessagesAndCache(t *testing.T) {
	store := NewStore("/test/dir")

	// Add some messages
	for i := 0; i < 50; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		store.AddMessage(role, strings.Repeat("Test message ", 100))
	}

	// Add moderate analysis cache
	store.AnalysisCache = &AnalysisCache{
		FileTree:       strings.Repeat("src/file.go\n", 1000),
		ReadmeContent:  strings.Repeat("Documentation\n", 500),
		PrimaryConfigs: []string{"go.mod"},
	}

	initialTokens := store.EstimateTokens()
	initialMessages := len(store.Messages)

	t.Logf("Initial: %d messages, %d tokens", initialMessages, initialTokens)

	cfg := &config.Config{Model: "test", OS: "macOS", APIURL: "http://test", APIKey: "test"}
	manager := &Manager{store: store, config: cfg, client: nil}

	// Only test if we're actually over emergency limits
	if initialTokens > 37500 {
		if err := manager.checkEmergencyPrune(); err != nil {
			t.Fatalf("checkEmergencyPrune failed: %v", err)
		}

		finalTokens := store.EstimateTokens()
		finalMessages := len(store.Messages)

		t.Logf("Final: %d messages, %d tokens", finalMessages, finalTokens)

		// Should have reduced tokens
		if finalTokens >= initialTokens {
			t.Errorf("Tokens should have been reduced: %d -> %d", initialTokens, finalTokens)
		}
	}
	if initialTokens <= 37500 {
		t.Logf("Tokens (%d) not high enough to trigger emergency pruning (threshold: 37500)", initialTokens)
	}
}
