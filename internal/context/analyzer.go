package context

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConfigFiles are common configuration files to detect
var ConfigFiles = []string{
	"go.mod",
	"package.json",
	"Cargo.toml",
	"pyproject.toml",
	"requirements.txt",
	"pom.xml",
	"build.gradle",
	"Makefile",
	"docker-compose.yml",
	"Dockerfile",
}

// ReadmeFiles are common README file names
var ReadmeFiles = []string{
	"README.md",
	"README.txt",
	"README",
	"readme.md",
	"Readme.md",
}

// Analyzer handles directory analysis
type Analyzer struct {
	rootDir      string
	gitignore    *GitignoreParser
	maxDepth     int
	maxFileSize  int64
	maxReadmeLen int
}

// NewAnalyzer creates a new directory analyzer
func NewAnalyzer(rootDir string) *Analyzer {
	return &Analyzer{
		rootDir:      rootDir,
		maxDepth:     2,          // Only descend 2 levels (reduced from 3)
		maxFileSize:  1024 * 50,  // Skip files > 50KB for tree
		maxReadmeLen: 5000,       // Max 5KB of README content
	}
}

// Analyze performs directory analysis and returns the cache
func (a *Analyzer) Analyze() (*AnalysisCache, error) {
	// Parse .gitignore if it exists
	a.gitignore = NewGitignoreParser(a.rootDir)
	if err := a.gitignore.Parse(); err != nil {
		// .gitignore is optional, continue without it
	}

	// Generate file tree
	tree, err := a.generateFileTree()
	if err != nil {
		return nil, fmt.Errorf("failed to generate file tree: %w", err)
	}

	// Find and read README
	readme := a.findReadme()

	// Detect config files
	configs := a.detectConfigFiles()

	return &AnalysisCache{
		FileTree:       tree,
		ReadmeContent:  readme,
		PrimaryConfigs: configs,
	}, nil
}

// generateFileTree creates a tree representation of the directory
func (a *Analyzer) generateFileTree() (string, error) {
	var builder strings.Builder
	builder.WriteString(filepath.Base(a.rootDir) + "/\n")

	if err := a.walkDirectory("", 0, &builder); err != nil {
		return "", err
	}

	tree := builder.String()

	// Aggressive truncation - max 10KB for file tree
	const maxTreeSize = 10000
	if len(tree) > maxTreeSize {
		tree = tree[:maxTreeSize] + "\n\n[File tree truncated - project too large]\n[Tip: Use 'ask' without --analyze for less context]"
	}

	return tree, nil
}

// walkDirectory recursively walks the directory structure
func (a *Analyzer) walkDirectory(relPath string, depth int, builder *strings.Builder) error {
	if depth > a.maxDepth {
		return nil
	}

	fullPath := filepath.Join(a.rootDir, relPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil // Skip directories we can't read
	}

	for _, entry := range entries {
		name := entry.Name()
		entryPath := filepath.Join(relPath, name)

		// Skip hidden files and gitignored paths
		if strings.HasPrefix(name, ".") && name != ".env.example" {
			continue
		}

		if a.gitignore.IsIgnored(entryPath) {
			continue
		}

		// Add indentation
		indent := strings.Repeat("  ", depth+1)
		if entry.IsDir() {
			builder.WriteString(fmt.Sprintf("%s%s/\n", indent, name))
			// Recurse into directory
			a.walkDirectory(entryPath, depth+1, builder)
		} else {
			// Check file size
			info, err := entry.Info()
			if err == nil && info.Size() < a.maxFileSize {
				builder.WriteString(fmt.Sprintf("%s%s\n", indent, name))
			}
		}
	}

	return nil
}

// findReadme looks for and reads a README file
func (a *Analyzer) findReadme() string {
	for _, filename := range ReadmeFiles {
		path := filepath.Join(a.rootDir, filename)
		if data, err := os.ReadFile(path); err == nil {
			content := string(data)

			// Aggressive truncation - max 5KB for README
			maxLen := 5000
			if len(content) > maxLen {
				content = content[:maxLen] + "\n\n[README truncated - too large]"
			}
			return content
		}
	}
	return ""
}

// detectConfigFiles finds common configuration files
func (a *Analyzer) detectConfigFiles() []string {
	var found []string
	for _, filename := range ConfigFiles {
		path := filepath.Join(a.rootDir, filename)
		if _, err := os.Stat(path); err == nil {
			found = append(found, filename)
		}
	}
	return found
}

// GitignoreParser handles .gitignore pattern matching
type GitignoreParser struct {
	rootDir  string
	patterns []string
}

// NewGitignoreParser creates a new gitignore parser
func NewGitignoreParser(rootDir string) *GitignoreParser {
	return &GitignoreParser{
		rootDir:  rootDir,
		patterns: []string{},
	}
}

// Parse reads and parses the .gitignore file
func (g *GitignoreParser) Parse() error {
	gitignorePath := filepath.Join(g.rootDir, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		g.patterns = append(g.patterns, line)
	}

	return scanner.Err()
}

// IsIgnored checks if a path matches any gitignore pattern
func (g *GitignoreParser) IsIgnored(path string) bool {
	// Common patterns to always ignore
	commonIgnores := []string{
		"node_modules",
		".git",
		"vendor",
		"target",
		"dist",
		"build",
		"__pycache__",
		".pytest_cache",
		".mypy_cache",
	}

	for _, pattern := range commonIgnores {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	// Check custom patterns (basic matching)
	for _, pattern := range g.patterns {
		if matchPattern(path, pattern) {
			return true
		}
	}

	return false
}

// matchPattern does basic glob pattern matching
func matchPattern(path, pattern string) bool {
	// Remove leading/trailing slashes
	pattern = strings.Trim(pattern, "/")
	path = strings.Trim(path, "/")

	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		pattern = strings.TrimSuffix(pattern, "/")
		return strings.HasPrefix(path, pattern+"/") || path == pattern
	}

	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		// Simple wildcard matching
		if pattern == "*" {
			return true
		}
		// *.ext pattern
		if strings.HasPrefix(pattern, "*.") {
			ext := strings.TrimPrefix(pattern, "*")
			return strings.HasSuffix(path, ext)
		}
	}

	// Exact match or contains
	return path == pattern || strings.Contains(path, "/"+pattern) || strings.HasPrefix(path, pattern+"/")
}

// AnalyzeDirectory is a convenience function to analyze the current directory
func AnalyzeDirectory(store *Store) error {
	analyzer := NewAnalyzer(store.Directory)
	cache, err := analyzer.Analyze()
	if err != nil {
		return err
	}

	store.AnalysisCache = cache
	now := time.Now()
	store.LastAnalysisAt = &now

	return nil
}
