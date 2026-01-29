// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yourorg/arc-sdk/ai"
)

// NewRootCmd creates the root command for arc-commit.
func NewRootCmd(aiCfg *ai.Config) *cobra.Command {
	root := &cobra.Command{
		Use:   "arc-commit",
		Short: "Git commit with AI-generated messages",
		Long: `Interactive commit workflow with AI-generated messages.

This command provides a guided workflow for creating git commits
with AI-generated commit messages based on staged changes.

The workflow:
  1. Checks for staged changes
  2. Generates commit message with AI
  3. Presents for approval/editing/regeneration
  4. Creates the commit`,
		Example: `  # Run the guided workflow with iterative approvals
  arc-commit

  # Commit immediately once the first suggestion looks good
  arc-commit --yes

  # Preview the generated message without writing the commit
  arc-commit --dry-run

  # Override the default model
  arc-commit --model claude-sonnet-4-5-20250929`,
	}

	root.AddCommand(
		newCommitCmd(aiCfg),
	)

	return root
}
