# Go Standards

Follow idiomatic Go so the code is predictable to anyone who reads Go.

## Formatting

- Always run `gofmt` (and `gofumpt` via `make format`)
- No manual formatting

## Naming

- Use mixedCaps (not snake_case)
- Keep common initialisms uppercase: `ID`, `URL`, `API`, `JSON`

## Error Handling

- Return errors, don't panic for expected failures
- Wrap with context using `%w`: `fmt.Errorf("operation failed: %w", err)`

## Context

- Pass `context.Context` into network operations
- Respect timeouts and cancellations

## Types

- Model request/response types with JSON tags
- Use pointers for optional fields, values for required fields
- Prefer typed `const` values (not raw strings) for API enums and resource types

## CLI Behavior

- If a flag is accepted, it must be implemented or error
- Never silently ignore flags
- Data goes to stdout, errors to stderr
- Keep JSON minified by default

## Dependencies

- Standard library first
- Avoid new deps unless necessary and justified

## CLI Help Output

- Use `UsageFunc` on ffcli commands for consistent help formatting
- When returning `flag.ErrHelp`, do **not** call `fs.Usage()` manually (avoids duplicate output)
- Help output is written to stderr by default
