# skan

> A fast, parallel file scanner for searching strings and patterns across directories.

## Description

`skan` recursively walks one or more directories and searches every file for a query string, printing each match with its file path and line number. Scanning is parallelized across all available CPU cores for speed.

## Usage

```bash
skan [options] DIRECTORIES...
```

## Global Options

| Option | Description |
| :--- | :--- |
| `--query string`, `-q string` | The string (or pattern) to search for in file contents (**required**) |
| `-i` | Perform a case-insensitive match (e.g. "Foo" matches "foo", "FOO") |
| `-v` | Invert results — print lines that do NOT contain the query |
| `-r` | Treat the query as a regular expression instead of a literal string |
| `-w` | Match whole words only (e.g. "cat" matches "cat" but not "cats" or "location") |
| `-B int` | Print `N` lines of leading context before matching lines (default: 0) |
| `-A int` | Print `N` lines of trailing context after matching lines (default: 0) |
| `-C int` | Print `N` lines of context before and after matching lines (default: 0) |
| `--color` | Colorize matching text in text output, doesn't affect JSON output |
| `--json` | Output results as newline-delimited JSON (one JSON object per match) |

## Examples

Search for a literal string in a directory:
```bash
skan -q "TODO" ./src
```

Case-insensitive search across multiple directories:
```bash
skan -q "error" -i ./logs ./tmp
```

Search using a regular expression (e.g., lines starting with "func "):
```bash
skan -q "^func " -r ./pkg
```

Inverted, case-insensitive search (find lines that do *not* contain "debug"):
```bash
skan -q "debug" -v -i ./src
```