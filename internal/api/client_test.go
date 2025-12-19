package api

import (
	"testing"

	"github.com/raitses/ask/internal/config"
)

func TestIsClaudeAPI(t *testing.T) {
	tests := []struct {
		name   string
		apiURL string
		want   bool
	}{
		{"OpenAI", "https://api.openai.com/v1/chat/completions", false},
		{"Claude", "https://api.anthropic.com/v1/messages", true},
		{"Claude mixed case", "https://API.ANTHROPIC.COM/v1/messages", true},
		{"Local with claude in name", "http://localhost:8080/claude", true},
		{"Generic local", "http://localhost:8080/v1/chat", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&config.Config{
				APIURL: tt.apiURL,
			})

			if got := client.IsClaudeAPI(); got != tt.want {
				t.Errorf("IsClaudeAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}
