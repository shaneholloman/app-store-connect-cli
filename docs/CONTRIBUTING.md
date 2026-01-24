# Contributing

## Git Workflow

- Branch from `main` and keep one logical change per branch
- Do not commit directly to `main` unless explicitly instructed; prefer PRs
- Prefer `git worktree add` for parallel tasks; remove with `git worktree remove` when done
- Rebase on `main` before merging; avoid merge commits
- Commit small, coherent changes; no WIP commits on shared branches
- Use concise, present-tense commit messages that match repo style
- Never commit secrets or local config files (keys, `.env`, `config.json`)

## Before Committing

```bash
make format     # Format code
make lint       # Check for issues
make test       # Run all tests
git diff        # Review changes before staging
```

## Adding a New Command

1. Create a command factory function (e.g., `MyCommand() *ffcli.Command`)
2. Follow the ffcli pattern from existing commands
3. Add to `RootCommand` subcommands list in `cmd/commands.go`
4. Write tests for validation and execution
5. Update README.md with usage examples

## Adding a New API Endpoint

1. Add method to `internal/asc/client.go`
2. Add types for request/response structs
3. Add helper functions for table/markdown output
4. Create command in `cmd/` to expose the endpoint
5. Write HTTP client tests with mocked responses

## Releases

Tag releases with plain semver like `0.1.0` (no `v` prefix).
