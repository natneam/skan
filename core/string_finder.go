package core

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
)

func regexHandler(ctx *LineContext) bool {
	contains := ctx.Regexp.Match(ctx.CurrentLine)

	if ctx.Args.Invert {
		contains = !contains
	}

	if !contains {
		return false
	}

	return true
}

func Find(args FindArgs) error {
	f, err := os.Open(args.File)
	if err != nil {
		return err
	}

	defer f.Close()

	query := []byte(args.Query)

	// Preprocess the Query
	if !args.Regex {
		query = []byte(regexp.QuoteMeta(string(query)))
	}
	if args.WholeWordsOnly {
		query = []byte("\\b" + string(query) + "\\b")
	}
	if args.CaseInsensitive {
		query = []byte("(?i)" + string(query))
	}
	regex, _ := regexp.Compile(string(query))

	// Read line by line and match
	scanner := bufio.NewScanner(f)
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Bytes()
		lineContext := &LineContext{
			CurrentLine: bytes.Clone(line),
			Regexp:      regex,
			Args:        args,
		}

		if regexHandler(lineContext) {
			args.Output <- Match{FileName: args.File, LineNumber: lineNumber, LineText: string(lineContext.CurrentLine)}
		}

		lineNumber++
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}
