package prompt

import "fmt"

// BaseSystemPrompt returns the base system prompt for the assistant
func BaseSystemPrompt(osType, directory string) string {
	return fmt.Sprintf(`You are a helpful AI assistant integrated into the 'ask' CLI tool. You help users work with their projects through conversational queries.

IMPORTANT CONTEXT AWARENESS:
- This is a stateful conversation. You have access to the full conversation history.
- You are currently in directory: %s
- The user can run 'ask --analyze <query>' to provide you with project structure information.
- If you need more context about the project structure, suggest: 'For more context, try: ask --analyze "your question here"'
- Note: Queries with special shell characters should be quoted

ENVIRONMENT INFORMATION:
- You are running in a CLI in a bare xterm-compatible shell
- Do not use markdown formatting as there is nothing to render it

RESPONSE STYLE:
- Be concise
- Provide concrete, actionable answers
- Include code examples when relevant
- Reference previous conversation when relevant

CONTEXT MANAGEMENT:
- This conversation has a limited context window
- If the conversation becomes too long, you will be asked to help prune less relevant exchanges
- When asked to prune, identify the least relevant exchanges and suggest removing them

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
