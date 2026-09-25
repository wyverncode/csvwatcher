# User Guide

The command accepts `--input`, `--output`, `--once`, `--interval`, `--stability-timeout`,
`--gui`, and `--help`. Durations use `ms`, `s`, or `m`, for example `250ms` or `30s`.

The first CSV row is the header. Headers must be non-empty and unique. Every later row must
have the same number of fields. Standard quoted CSV values are supported.

Output uses the same base filename with a `.json` extension, is formatted with two-space
indentation, and is published atomically. Failed conversions are retried on a later scan.
