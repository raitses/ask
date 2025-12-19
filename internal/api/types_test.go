package api

import (
	"encoding/json"
	"testing"
)

func TestChatMessageCacheControl(t *testing.T) {
	tests := []struct {
		name     string
		msg      ChatMessage
		wantJSON string
	}{
		{
			name: "message without cache control",
			msg: ChatMessage{
				Role:    "user",
				Content: "Hello",
			},
			wantJSON: `{"role":"user","content":"Hello"}`,
		},
		{
			name: "message with cache control",
			msg: ChatMessage{
				Role:    "system",
				Content: "You are helpful",
				CacheControl: &CacheControl{Type: "ephemeral"},
			},
			wantJSON: `{"role":"system","content":"You are helpful","cache_control":{"type":"ephemeral"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJSON, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("JSON mismatch:\ngot:  %s\nwant: %s", gotJSON, tt.wantJSON)
			}
		})
	}
}
