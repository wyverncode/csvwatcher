# CSV Watcher

[![CI](https://github.com/wyvercode/csvwatcher/actions/workflows/ci.yml/badge.svg)](https://github.com/wyvercode/csvwatcher/actions/workflows/ci.yml)
[![npm](https://img.shields.io/npm/v/%40wyverncode%2Fcsvwatcher)](https://github.com/wyverncode/csvwatcher/packages)

Cross-platform Node.js tooling that watches a folder for CSV files and converts them into
validated, human-readable JSON. Use the CLI for automation or the local browser interface
for interactive workflows.

> **New to Node.js?** You only need Node.js 20+, an input folder, and an output folder. The
> [Wiki](https://github.com/wyverncode/csvwatcher/wiki) explains the concepts and gives
> copy-and-paste examples.

## Requirements

- Node.js 20 or newer
- npm 10 or newer

## Install

```bash
npm install -g @wyverncode/csvwatcher --registry=https://npm.pkg.github.com
```

GitHub Packages requires a GitHub personal access token with `read:packages`:

```bash
npm login --scope=@wyverncode --registry=https://npm.pkg.github.com
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

## What happens when it runs?

1. CSV files already in the input folder are found.
2. The first row of each file is treated as the header.
3. Each later row becomes a JSON object.
4. The JSON file is written to the output folder with the same base name.
5. The folder is scanned repeatedly for new or changed CSV files.

Example:

```text
incoming/orders.csv  ->  converted/orders.json
```

If a file is still being copied, CSV Watcher waits until its size and modification time stop
changing. If conversion fails, it is retried on a later scan.

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

The package is scoped as `@wyverncode/csvwatcher` and publishes to GitHub Packages. Maintainers
can publish from a clean `main` checkout with:

```bash
npm publish
```

The release workflow publishes when a `v*` tag is pushed. See [CONTRIBUTING.md](CONTRIBUTING.md)
and [SECURITY.md](SECURITY.md) for repository standards.

For a complete beginner guide, operational reference, architecture overview, and
troubleshooting steps, see the [project Wiki](https://github.com/wyverncode/csvwatcher/wiki)
or the [documentation source in this repository](docs/wiki/Home.md).

## License

MIT. See [LICENSE](LICENSE).
