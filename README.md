# CSV Watcher

A small Go command-line tool that watches a directory for CSV files and converts them to formatted JSON.

## Requirements

- Go 1.22 or newer

## Usage

```powershell
go run . --input .\incoming --output .\converted
```

Install the published CLI with:

```powershell
go install github.com/wyverncode/csvwatcher@latest
```

For a one-time conversion (useful in scripts and CI):

```powershell
go run . --input .\incoming --output .\converted --once
```

The scan interval and file readiness timeout can be tuned when needed:

```powershell
go run . --input .\incoming --output .\converted --interval 250ms --stability-timeout 30s
```

The program:

- Converts CSV files already present in the input directory at startup.
- Watches for new or changed `.csv` files.
- Uses the first CSV row as JSON property names.
- Preserves quoted values and escapes JSON safely.
- Writes one indented JSON array per CSV file.
- Waits for a file to stop changing before reading it.
- Logs successful conversions and errors.
- Stops cleanly when you press `Ctrl+C` or the process receives `SIGTERM`.
- Retries files whose conversion fails on a later scan.
- Rejects input and output directories that overlap.

Run `go run . --help` for the complete option list.

Build a local binary with:

```powershell
go build -o csvwatcher.exe .
.\csvwatcher.exe --input .\incoming --output .\converted
```

## Development

Run formatting, tests, and a build before opening a pull request:

```powershell
gofmt -w .
go test ./...
go build ./...
```

The repository includes GitHub Actions CI for formatting, tests, and builds. Keep generated
JSON and local input/output folders out of commits; the included `.gitignore` already covers
the standard local paths.
