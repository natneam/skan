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

	// Buffering Variables
	var beforeBuffer []ContextLine

	// Matched Lines
	var matchedLines []Match

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

			matchedLines = append(matchedLines, Match{FileName: args.File, LineNumber: lineNumber, LineText: string(lineContext.CurrentLine), BeforeContext: append([]ContextLine{}, beforeBuffer...)})

			// reset buffers
			beforeBuffer = nil

		} else {
			lineText := string(lineContext.CurrentLine)
			remove := 0

			for i := range matchedLines {
				if len(matchedLines[i].AfterContext) < args.ContextLines.After {
					matchedLines[i].AfterContext = append(matchedLines[i].AfterContext, ContextLine{FileName: args.File, LineNumber: lineNumber, LineText: lineText})
				}

				if len(matchedLines[i].AfterContext) == args.ContextLines.After {
					args.Output <- matchedLines[i]
					remove++
				}
			}

			matchedLines = matchedLines[remove:]

			// Record before context
			beforeBuffer = append(beforeBuffer, ContextLine{FileName: args.File, LineNumber: lineNumber, LineText: lineText})
			if len(beforeBuffer) > args.ContextLines.Before {
				beforeBuffer = beforeBuffer[1:]
			}
		}

		lineNumber++
	}

	// flush out the matched lines one last time
	for _, m := range matchedLines {
		args.Output <- m
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}
