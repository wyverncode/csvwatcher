# Getting Started

Install Node.js 20 or newer, then install the package from GitHub Packages:

```bash
npm login --scope=@wyverncode --registry=https://npm.pkg.github.com
npm install -g @wyverncode/csvwatcher --registry=https://npm.pkg.github.com
```

Create two separate folders, such as `incoming` and `converted`, then run:

```bash
csvwatcher --input ./incoming --output ./converted
```

Create `incoming/people.csv`:

```csv
name,email
Ada,ada@example.com
Grace,grace@example.com
```

The tool writes `converted/people.json`. Use `--once` for scripts that should process current
files and exit. Use `--gui` to start the browser interface at `http://127.0.0.1:8080`.
