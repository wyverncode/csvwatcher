# Maintainer Guide

Run `npm ci`, `npm run check`, and `npm pack --dry-run` before a release. Update the package
version, commit to `main`, and push an annotated `v*` tag. GitHub Actions runs checks and the
tag workflow publishes `@wyverncode/csvwatcher` to GitHub Packages.

Never commit tokens, generated JSON, local data, or package credentials. Update this
documentation when user-facing behavior changes.
