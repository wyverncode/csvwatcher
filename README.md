# CSV Watcher

[![CI](https://github.com/wyverncode/csvwatcher/actions/workflows/ci.yml/badge.svg)](https://github.com/wyverncode/csvwatcher/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/wyverncode/csvwatcher.svg)](https://pkg.go.dev/github.com/wyverncode/csvwatcher)
[![Latest Release](https://img.shields.io/github/v/release/wyverncode/csvwatcher)](https://github.com/wyverncode/csvwatcher/releases)

Cross-platform Go tooling that watches a folder for CSV files and converts them into
validated, human-readable JSON. Use the command line for automation or the local browser
interface for interactive workflows.

## Features

- Processes CSV files already present at startup.
- Watches for new and changed `.csv` files.
- Uses the first row as JSON property names.
- Preserves quoted values and safely escapes JSON characters.
- Rejects empty or duplicate headers and mismatched row lengths.
- Waits for files to stop changing before conversion.
- Retries failed conversions on a later scan.
- Publishes output atomically to prevent partial JSON files.
- Supports graceful shutdown with `Ctrl+C` or `SIGTERM`.
- Includes a standard-library-only local web interface.

## Quick start

### CLI

Requirements: Go 1.22 or newer.

```powershell
go install github.com/wyverncode/csvwatcher@latest
csvwatcher --input .\incoming --output .\converted
```

For a one-time conversion:

```powershell
csvwatcher --input .\incoming --output .\converted --once
```

### Browser interface

```powershell
csvwatcher --gui
```

Open <http://127.0.0.1:8080>. The interface provides folder configuration, start/stop
controls, status, and live activity logs. The server binds to localhost only.

## Configuration

| Flag | Default | Purpose |
| --- | --- | --- |
| `--input` | — | Input folder containing CSV files; required for CLI mode |
| `--output` | — | Output folder for JSON files; required for CLI mode |
| `--once` | `false` | Convert current files and exit |
| `--interval` | `1s` | Scan interval for changed files |
| `--stability-timeout` | `10s` | Maximum wait for a file to stop changing |
| `--gui` | `false` | Start the local browser interface |

Run `csvwatcher --help` for the authoritative option list.

## Conversion contract

Each CSV file produces a JSON array with the same base name:

```text
incoming/customers.csv
converted/customers.json
```

The first CSV row is the header. Every subsequent row must contain exactly the same number
of fields. Empty headers and duplicate headers are rejected. JSON output is indented with
two spaces and ends with a newline.

## Build from source

```powershell
git clone https://github.com/wyverncode/csvwatcher.git
cd csvwatcher
go test ./...
go build -o csvwatcher.exe .
.\csvwatcher.exe --input .\incoming --output .\converted
```

For a local GUI build:

```powershell
.\csvwatcher.exe --gui
```

## Development

The repository uses standard Go tooling and GitHub Actions CI:

```powershell
gofmt -w .
go vet ./...
go test -race ./...
go build ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the pull request workflow and quality
expectations. Generated JSON and local input/output folders are ignored by default.

## Security and support

The GUI is intentionally bound to `127.0.0.1`; do not expose it directly to an untrusted
network. See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## License

Released under the [MIT License](LICENSE).
