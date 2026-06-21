```
NAME:
   skan - A fast, parallel file scanner for searching strings and patterns across directories

USAGE:
   skan [options] DIRECTORIES...

DESCRIPTION:
   skan recursively walks one or more directories and searches every file for a
   query string, printing each match with its file path and line number. Scanning
   is parallelized across all available CPU cores for speed.

GLOBAL OPTIONS:
   --query string, -q string  The string (or pattern) to search for in file contents (required)
   -i                         Perform a case-insensitive match (e.g. "Foo" matches "foo", "FOO")
   -v                         Invert results — print lines that do NOT contain the query
   -r                         Treat the query as a regular expression instead of a literal string
   --help, -h                 Show this help message

EXAMPLES:
   skan -q "TODO" ./src
   skan -q "error" -i ./logs ./tmp
   skan -q "^func " -r ./pkg
   skan -q "debug" -v -i ./src

```