package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzerFileTree(t *testing.T) {
	// Create a temporary test directory
	tmpDir := t.TempDir()
	
	// Create some test files
	_ = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "src/main.go"), []byte("package main"), 0644)
	
	analyzer := NewAnalyzer(tmpDir)
	cache, err := analyzer.Analyze()
	
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if cache.FileTree == "" {
		t.Error("FileTree should not be empty")
	}
	
	if cache.ReadmeContent == "" {
		t.Error("README should have been found")
	}
	
	if len(cache.PrimaryConfigs) == 0 {
		t.Error("go.mod should have been detected")
	}
	
	// Verify go.mod was found
	found := false
	for _, cfg := range cache.PrimaryConfigs {
		if cfg == "go.mod" {
			found = true
			break
		}
	}
	if !found {
		t.Error("go.mod should be in PrimaryConfigs")
	}
	
	t.Logf("File Tree:\n%s", cache.FileTree)
	t.Logf("README (first 50 chars): %s", cache.ReadmeContent[:min(50, len(cache.ReadmeContent))])
	t.Logf("Configs: %v", cache.PrimaryConfigs)
}

func TestGitignoreParser(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a .gitignore
	gitignore := `# Test gitignore
node_modules
*.log
dist/
`
	_ = os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignore), 0644)
	
	parser := NewGitignoreParser(tmpDir)
	err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	// Test pattern matching
	tests := []struct {
		path     string
		ignored  bool
	}{
		{"node_modules", true},
		{"node_modules/pkg", true},
		{"src/node_modules/lib", true},
		{"test.log", true},
		{"dist", true},
		{"dist/output.js", true},
		{"src/main.go", false},
		{"README.md", false},
	}
	
	for _, tt := range tests {
		result := parser.IsIgnored(tt.path)
		if result != tt.ignored {
			t.Errorf("IsIgnored(%q) = %v, want %v", tt.path, result, tt.ignored)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
