# Contributing

## Before you start

Check existing issues before opening a new one. For security vulnerabilities, follow
[SECURITY.md](SECURITY.md) instead of filing a public issue.

## Local workflow

```powershell
gofmt -w .
go vet ./...
go test -race ./...
go build ./...
```

Keep changes focused, preserve the CLI and conversion contract, and add tests for behavior
that changes. Avoid committing generated JSON, local folders, binaries, or credentials.

## Pull requests

Use a descriptive title and explain:

- What changed and why.
- How the change was tested.
- Any compatibility or operational impact.

Pull requests must pass GitHub Actions CI and should include documentation updates when
user-facing behavior changes. Squash-merge is preferred for focused changes.
