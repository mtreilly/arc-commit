# arc-commit

Git commit workflow with AI-generated messages.

## Features

- Interactive commit workflow
- AI-generated commit messages based on staged changes
- Approval, editing, and regeneration options
- Dry-run mode for previewing

## Installation

```bash
go install github.com/mtreilly/arc-commit@latest
```

## Usage

```bash
# Run the guided workflow
arc-commit

# Auto-approve the first suggestion
arc-commit --yes

# Preview without committing
arc-commit --dry-run

# Use a specific model
arc-commit --model claude-sonnet-4-5-20250929
```

## Workflow

1. Checks for staged changes
2. Generates commit message with AI
3. Presents for approval/editing/regeneration
4. Creates the commit

## License

MIT
