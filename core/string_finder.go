package core

import (
	"bufio"
	"io"
	"os"

	"natneam.com/skan/model"
)

func regexHandler(ctx *model.LineContext) (bool, [][]int) {
	contains := ctx.Args.Regexp.Match(ctx.CurrentLine)
	var matchIndexes [][]int

	if !ctx.Args.Invert {
		matchIndexes = ctx.Args.Regexp.FindAllIndex(ctx.CurrentLine, -1)
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
	f, err := os.Open(args.Job.AbsolutePath)
	if err != nil {
		return err
	}

	defer f.Close()

	fileDisplayName := args.Job.RelativePath
	if !args.Job.IsRelative {
		fileDisplayName = args.Job.AbsolutePath
	}

	// Buffering Variables
	var beforeBuffer []model.ContextLine

	// Matched Lines
	var matchedLines []model.Match

	// Read line by line and match
	reader := bufio.NewReader(f)
	lineNumber := 0

	for {
		lineNumber++
		line, err := reader.ReadBytes('\n')

		if err != io.EOF && err != nil {
			break
		}

		lineContext := &model.LineContext{
			CurrentLine: line,
			Args:        args,
		}

		match, matchIndexes := regexHandler(lineContext)
		lineText := string(lineContext.CurrentLine)
		if match {

			matchedLines = append(matchedLines, model.Match{FileName: fileDisplayName, LineNumber: lineNumber, LineText: lineText, BeforeContext: append([]model.ContextLine{}, beforeBuffer...), AfterContext: []model.ContextLine{}, MatchIndexes: matchIndexes})

		}

		remove := 0

		for i := range matchedLines {
			if len(matchedLines[i].AfterContext) < args.ContextLines.After {
				if !(match && i == len(matchedLines)-1) {
					matchedLines[i].AfterContext = append(matchedLines[i].AfterContext, model.ContextLine{FileName: fileDisplayName, LineNumber: lineNumber, LineText: lineText})
				}
			}

			if len(matchedLines[i].AfterContext) == args.ContextLines.After {
				args.Output <- matchedLines[i]
				remove++
			}
		}

		matchedLines = matchedLines[remove:]

		// Record before context
		beforeBuffer = append(beforeBuffer, model.ContextLine{FileName: fileDisplayName, LineNumber: lineNumber, LineText: lineText})
		if len(beforeBuffer) > args.ContextLines.Before {
			beforeBuffer = beforeBuffer[1:]
		}

		if err == io.EOF {
			break
		}

	}

	// flush out the matched lines one last time
	for _, m := range matchedLines {
		args.Output <- m
	}

	return nil
}
