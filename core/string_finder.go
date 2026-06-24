package core

import (
	"bufio"
	"bytes"
	"os"
	"regexp"

	"natneam.com/skan/model"
)

func regexHandler(ctx *model.LineContext) (bool, [][]int) {
	contains := ctx.Regexp.Match(ctx.CurrentLine)
	var matchIndexes [][]int

	if !ctx.Args.Invert {
		matchIndexes = ctx.Regexp.FindAllIndex(ctx.CurrentLine, -1)
	}

	if ctx.Args.Invert {
		contains = !contains
	}

	if !contains {
		return false, matchIndexes
	}

	return true, matchIndexes
}

func Find(args model.FindArgs) error {
	f, err := os.Open(args.File)
	if err != nil {
		return err
	}

	defer f.Close()

	query := []byte(args.Query)

	// Buffering Variables
	var beforeBuffer []model.ContextLine

	// Matched Lines
	var matchedLines []model.Match

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
		lineContext := &model.LineContext{
			CurrentLine: bytes.Clone(line),
			Regexp:      regex,
			Args:        args,
		}

		match, matchIndexes := regexHandler(lineContext)
		if match {

			matchedLines = append(matchedLines, model.Match{FileName: args.File, LineNumber: lineNumber, LineText: string(lineContext.CurrentLine), BeforeContext: append([]model.ContextLine{}, beforeBuffer...), MatchIndexes: matchIndexes})

			// reset buffers
			beforeBuffer = nil

		} else {
			lineText := string(lineContext.CurrentLine)
			remove := 0

			for i := range matchedLines {
				if len(matchedLines[i].AfterContext) < args.ContextLines.After {
					matchedLines[i].AfterContext = append(matchedLines[i].AfterContext, model.ContextLine{FileName: args.File, LineNumber: lineNumber, LineText: lineText})
				}

				if len(matchedLines[i].AfterContext) == args.ContextLines.After {
					args.Output <- matchedLines[i]
					remove++
				}
			}

			matchedLines = matchedLines[remove:]

			// Record before context
			beforeBuffer = append(beforeBuffer, model.ContextLine{FileName: args.File, LineNumber: lineNumber, LineText: lineText})
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
