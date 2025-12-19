package prompt

import "fmt"

// BaseSystemPrompt returns the base system prompt for the assistant
func BaseSystemPrompt(osType, directory string) string {
	return fmt.Sprintf(`You are an AI assistant in the 'ask' CLI tool helping with projects via conversational queries.

CONTEXT:
- Stateful conversation with full history
- Current directory: %s
- Run 'ask --analyze <query>' for project structure
- Suggest analysis when needed: 'For more context, try: ask --analyze "your question"'
- Quote shell special characters

ENVIRONMENT:
- CLI in xterm-compatible shell
- No markdown formatting

STYLE:
- Concise, actionable answers
- Include code examples when relevant
- Reference prior conversation

PRUNING:
- Limited context window
- When asked to prune, identify least relevant exchanges

OS: %s`, directory, osType)
}

// AnalysisSystemPrompt returns additional context when directory analysis is available
func AnalysisSystemPrompt(fileTree, readme string, configs []string) string {
	prompt := "\n\nPROJECT ANALYSIS:\nThe following information has been gathered about this project:\n\n"

	if fileTree != "" {
		prompt += fmt.Sprintf("FILE TREE:\n%s\n\n", fileTree)
	}

	if readme != "" {
		prompt += fmt.Sprintf("README:\n%s\n\n", readme)
	}

	if len(configs) > 0 {
		prompt += "PRIMARY CONFIGURATION FILES:\n"
		for _, cfg := range configs {
			prompt += fmt.Sprintf("- %s\n", cfg)
		}
		prompt += "\n"
	}

	prompt += "Use this information to provide more accurate and project-specific responses."

	return prompt
}
