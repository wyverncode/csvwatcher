# CSV Watcher

A small Go command-line tool that watches a directory for CSV files and converts them to formatted JSON.

## Requirements

- Go 1.22 or newer

## Usage

```powershell
go run . --input .\incoming --output .\converted
```

The program:

- Converts CSV files already present in the input directory at startup.
- Watches for new or changed `.csv` files.
- Uses the first CSV row as JSON property names.
- Preserves quoted values and escapes JSON safely.
- Writes one indented JSON array per CSV file.
- Waits for a file to stop changing before reading it.
- Logs successful conversions and errors.

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
