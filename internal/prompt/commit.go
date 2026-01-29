// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package prompt

// CommitMessageModel is the default model for commit message generation.
const CommitMessageModel = "claude-haiku-4-5-20251001"

// CommitMessage returns the system and user prompts for generating a commit message.
func CommitMessage(diff, feedback string) (system, user string) {
	system = `You are an expert developer who writes clear, professional commit messages following conventional commits format.

Your task is to generate a commit message based on git diff output. Follow these principles:

1. **Format**: Use conventional commits (feat:, fix:, refactor:, docs:, test:, chore:)
2. **Subject line**: Concise summary (max 72 chars), imperative mood ("add" not "added")
3. **Body**: Explain WHY, not WHAT (the diff shows what changed)
4. **Scope**: Add scope when helpful (e.g., "feat(cli):", "fix(database):")
5. **Breaking changes**: Use "!" for breaking changes (e.g., "feat!:")

Style guidelines:
- Clear and professional tone
- No unnecessary words or filler
- Focus on user impact and intent
- Group related changes logically

Output ONLY the commit message, no additional commentary.`

	user = `Generate a conventional commit message for these changes:

` + diff

	if feedback != "" {
		user += `

User feedback for improvement: ` + feedback
	}

	return system, user
}
