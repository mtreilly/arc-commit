// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-commit/internal/prompt"
	"github.com/yourorg/arc-sdk/ai"
	"github.com/yourorg/arc-sdk/errors"
)

// newCommitCmd creates the commit subcommand.
func newCommitCmd(aiCfg *ai.Config) *cobra.Command {
	var (
		autoYes bool
		dryRun  bool
		model   string
	)

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Create commit with AI-generated message",
		Long: `Interactive commit workflow with AI-generated message.

This command provides a guided workflow:
  1. Checks for staged changes
  2. Generates commit message with AI
  3. Presents for approval/editing/regeneration
  4. Creates the commit`,
		Example: `  # Run the guided workflow with iterative approvals
  arc-commit commit

  # Commit immediately once the first suggestion looks good
  arc-commit commit --yes

  # Preview the generated message without writing the commit
  arc-commit commit --dry-run

  # Override the default model
  arc-commit commit --model claude-sonnet-4-5-20250929`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build effective config with flag overrides
			cfg := *aiCfg
			if model != "" {
				cfg.DefaultModel = model
			}

			return runInteractiveCommit(&cfg, autoYes, dryRun)
		},
	}

	cmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Generate message but don't commit")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model to use (default: "+prompt.CommitMessageModel+")")

	return cmd
}

// runInteractiveCommit implements the interactive commit workflow.
func runInteractiveCommit(cfg *ai.Config, autoYes, dryRun bool) error {
	// 1. Check for staged changes
	fmt.Println("Checking for staged changes...")
	if err := checkStagedChanges(); err != nil {
		return errors.NewCLIError("no staged changes found").
			WithHint("Stage changes first: git add <files>")
	}

	// 2. Get diff
	fmt.Println("Generating diff...")
	diff, err := getStagedDiff()
	if err != nil {
		return errors.NewCLIError("failed to get diff").WithCause(err)
	}

	if len(diff) == 0 {
		return errors.NewCLIError("no changes to commit").
			WithHint("Stage changes first: git add <files>")
	}

	// 3. Create AI client and service
	client, err := ai.NewClient(*cfg)
	if err != nil {
		return errors.NewCLIError("failed to create AI client").WithCause(err)
	}

	// Set default model if not specified
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = prompt.CommitMessageModel
	}

	service := ai.NewService(client, *cfg)

	// 4. Initial message generation
	fmt.Println("Generating commit message with AI...")
	message, err := generateCommitMessage(service, diff, "")
	if err != nil {
		return errors.NewCLIError("failed to generate commit message").WithCause(err)
	}

	// 5. Interactive loop
	reader := bufio.NewReader(os.Stdin)
	for {
		// Display message
		fmt.Println("\n" + strings.Repeat("=", 70))
		fmt.Println(message)
		fmt.Println(strings.Repeat("=", 70))

		// Dry run: show and exit
		if dryRun {
			fmt.Println("\n(Dry run - no commit created)")
			return nil
		}

		// Auto-yes: commit without prompting
		if autoYes {
			fmt.Println("\nAuto-committing...")
			return createCommit(message)
		}

		// Prompt user
		fmt.Print("\n[y]es, [n]o (regenerate), [e]dit, [c]ancel: ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			return errors.NewCLIError("failed to read input").WithCause(err)
		}

		choice = strings.ToLower(strings.TrimSpace(choice))

		switch choice {
		case "y", "yes":
			return createCommit(message)

		case "n", "no":
			fmt.Print("\nWhat would you like improved? (or press Enter for generic): ")
			feedback, _ := reader.ReadString('\n')
			feedback = strings.TrimSpace(feedback)

			fmt.Println("\nRegenerating...")
			message, err = generateCommitMessage(service, diff, feedback)
			if err != nil {
				return errors.NewCLIError("failed to regenerate message").WithCause(err)
			}

		case "e", "edit":
			edited, err := editInEditor(message)
			if err != nil {
				return errors.NewCLIError("failed to open editor").WithCause(err)
			}
			return createCommit(edited)

		case "c", "cancel":
			fmt.Println("\nCommit cancelled.")
			return nil

		default:
			fmt.Println("\nInvalid choice. Please enter y/n/e/c.")
		}
	}
}

// generateCommitMessage generates a commit message from diff and optional feedback.
func generateCommitMessage(service *ai.Service, diff, feedback string) (string, error) {
	systemPrompt, userPrompt := prompt.CommitMessage(diff, feedback)

	ctx := context.Background()
	resp, err := service.Run(ctx, ai.RunOptions{
		System: systemPrompt,
		Prompt: userPrompt,
	})
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}

	return strings.TrimSpace(resp.Text), nil
}

// checkStagedChanges checks if there are staged changes in git.
func checkStagedChanges() error {
	cmd := exec.Command("git", "diff", "--staged", "--quiet")
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Exit code 1 means there are differences (staged changes exist)
			return nil
		}
		return fmt.Errorf("failed to check staged changes: %w", err)
	}
	// Exit code 0 means no differences (no staged changes)
	return fmt.Errorf("no staged changes")
}

// getStagedDiff gets the diff of staged changes.
func getStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	return string(output), nil
}

// editInEditor opens the message in the user's editor.
func editInEditor(message string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "arc-commit-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write message to temp file
	if _, err := tmpFile.WriteString(message); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Open editor
	editCmd := exec.Command(editor, tmpFile.Name())
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return "", fmt.Errorf("editor failed: %w", err)
	}

	// Read edited content
	edited, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}

	return string(edited), nil
}

// createCommit creates a git commit with the given message.
func createCommit(message string) error {
	cmd := exec.Command("git", "commit", "-F", "-")
	cmd.Stdin = strings.NewReader(message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}
