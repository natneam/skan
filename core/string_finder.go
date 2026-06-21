package core

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
)

type Handler interface {
	Execute(*LineContext) bool
	SetNext(Handler)
}

type BaseHandler struct {
	next Handler
}

func (b *BaseHandler) SetNext(h Handler) { b.next = h }
func (b *BaseHandler) Next(c *LineContext) bool {
	if b.next != nil {
		return b.next.Execute(c)
	}

	return true
}

type CaseInsenstitvityHandler struct{ BaseHandler }

func (h *CaseInsenstitvityHandler) Execute(ctx *LineContext) bool {
	if ctx.Args.CaseInsensitive {
		ctx.CurrentLine = bytes.ToLower(ctx.CurrentLine)
	}
	return h.Next(ctx)
}

type SearchHandler struct{ BaseHandler }

func (h *SearchHandler) Execute(ctx *LineContext) bool {
	contains := bytes.Contains(ctx.CurrentLine, ctx.Query)

	if ctx.Args.Invert {
		contains = !contains
	}

	if !contains {
		return false
	}

	return h.Next(ctx)
}

type RegexHandler struct{ BaseHandler }

func (h *RegexHandler) Execute(ctx *LineContext) bool {
	if ctx.Regexp != nil {
		contains := ctx.Regexp.Match(ctx.CurrentLine)

		if ctx.Args.Invert {
			contains = !contains
		}

		if !contains {
			return false
		}
	}

	return h.Next(ctx)
}

func Find(args FindArgs) error {
	f, err := os.Open(args.File)
	if err != nil {
		return err
	}

	defer f.Close()

	caseHandler := &CaseInsenstitvityHandler{}
	searchHandler := &SearchHandler{}
	regexHandler := &RegexHandler{}

	var startingHandler Handler
	query := []byte(args.Query)
	var regex *regexp.Regexp

	if args.Regex {
		startingHandler = regexHandler
		if args.CaseInsensitive {
			query = append([]byte("(?i)"), query...)
		}
		regex, _ = regexp.Compile(string(query))
	} else {
		startingHandler = caseHandler
		startingHandler.SetNext(searchHandler) // case -> search
		if args.CaseInsensitive {
			query = bytes.ToLower(query)
		}
	}

	scanner := bufio.NewScanner(f)
	lineNumber := 1

	for scanner.Scan() {
		originalLine := scanner.Bytes()
		lineContext := &LineContext{
			OriginalLine: originalLine,
			CurrentLine:  bytes.Clone(originalLine),
			Query:        query,
			Regexp:       regex,
			Args:         args,
		}
		if startingHandler.Execute(lineContext) {
			args.Output <- Match{FileName: args.File, LineNumber: lineNumber, LineText: string(originalLine)}
		}

		lineNumber++
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}
