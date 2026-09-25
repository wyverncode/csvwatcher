# How It Works

CSV Watcher validates the input and output paths, processes existing CSV files, waits for
files to stop changing, parses rows into objects, and atomically replaces JSON outputs.
Polling is used instead of native filesystem dependencies so the same package behaves
consistently on Windows, macOS, and Linux.

The main modules are `src/cli.js` for arguments and lifecycle, `src/csv.js` for validation,
`src/converter.js` for JSON output, `src/watcher.js` for scanning and retries, and `src/web.js`
for the localhost browser interface.
