# CSV Watcher

[![CI](https://github.com/wyvercode/csvwatcher/actions/workflows/ci.yml/badge.svg)](https://github.com/wyvercode/csvwatcher/actions/workflows/ci.yml)
[![npm](https://img.shields.io/npm/v/%40wyvercode%2Fcsvwatcher)](https://github.com/wyvercode/csvwatcher/packages)

Cross-platform Node.js tooling that watches a folder for CSV files and converts them into
validated, human-readable JSON. Use the CLI for automation or the local browser interface
for interactive workflows.

## Requirements

- Node.js 20 or newer
- npm 10 or newer

## Install

```bash
npm install -g @wyvercode/csvwatcher --registry=https://npm.pkg.github.com
```

GitHub Packages requires a GitHub personal access token with `read:packages`:

```bash
npm login --scope=@wyvercode --registry=https://npm.pkg.github.com
```

## Quick start

```bash
csvwatcher --input ./incoming --output ./converted
```

Convert existing files once and exit:

```bash
csvwatcher --input ./incoming --output ./converted --once
```

Launch the local browser interface:

```bash
csvwatcher --gui
```

Open <http://127.0.0.1:8080>. The GUI is localhost-only and includes folder configuration,
start/stop controls, status, and live activity logs.

## CLI options

| Option                         | Default | Description                         |
| ------------------------------ | ------- | ----------------------------------- |
| `--input DIR`                  | —       | Input folder containing CSV files   |
| `--output DIR`                 | —       | Output folder for JSON files        |
| `--once`                       | `false` | Convert current files and exit      |
| `--interval DURATION`          | `1s`    | Scan interval (`250ms`, `1s`, `1m`) |
| `--stability-timeout DURATION` | `10s`   | Maximum file-readiness wait         |
| `--gui`                        | `false` | Start the local browser interface   |
| `--help`                       | —       | Show usage                          |

## Conversion contract

The first CSV row becomes the JSON object keys. Every later row must have exactly the same
number of fields. Empty or duplicate headers and malformed quoted fields are rejected.

```text
incoming/customers.csv  ->  converted/customers.json
```

Output is a two-space-indented JSON array, written atomically so readers never see a partial
file. Failed files are retried on a later scan.

## Development

```bash
git clone https://github.com/wyvercode/csvwatcher.git
cd csvwatcher
npm ci
npm run check
npm start -- --input ./incoming --output ./converted
```

`npm run check` runs Prettier validation, ESLint, and the Node.js test suite. GitHub Actions
also runs the check suite on every push and pull request.

## Publishing

The package is scoped as `@wyvercode/csvwatcher` and publishes to GitHub Packages. Maintainers
can publish from a clean `main` checkout with:

```bash
npm publish
```

The release workflow publishes when a `v*` tag is pushed. See [CONTRIBUTING.md](CONTRIBUTING.md)
and [SECURITY.md](SECURITY.md) for repository standards.

## License

MIT. See [LICENSE](LICENSE).
