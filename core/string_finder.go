package core

import (
	"bufio"
	"bytes"
	"os"
)

type Handler interface {
	Excute(*LineContext) bool
	SetNext(Handler)
}

type BaseHandler struct {
	next Handler
}

func (b *BaseHandler) SetNext(h Handler) { b.next = h }
func (b *BaseHandler) Next(c *LineContext) bool {
	if b.next != nil {
		return b.next.Excute(c)
	}

	return true
}

type CaseInsenstitvityHandler struct{ BaseHandler }

func (h CaseInsenstitvityHandler) Excute(ctx *LineContext) bool {
	if ctx.Args.CaseInsensitive {
		ctx.CurrentLine = bytes.ToLower(ctx.CurrentLine)
	}
	return h.Next(ctx)
}

type SearchHandler struct{ BaseHandler }

func (h SearchHandler) Excute(ctx *LineContext) bool {
	contains := bytes.Contains(ctx.CurrentLine, ctx.Query)

	if ctx.Args.Invert {
		contains = !contains
	}

	if !contains {
		return false
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

	caseHandler.SetNext(searchHandler) // case -> search

	query := []byte(args.Query)
	if args.CaseInsensitive {
		query = bytes.ToLower(query)
	}

	scanner := bufio.NewScanner(f)
	lineNumber := 1

	for scanner.Scan() {
		originalLine := scanner.Bytes()
		lineContext := &LineContext{
			OriginalLine: originalLine,
			CurrentLine:  bytes.Clone(originalLine),
			Query:        query,
			Args:         args,
		}
		if caseHandler.Excute(lineContext) {
			args.Output <- Match{FileName: args.File, LineNumber: lineNumber, LineText: string(originalLine)}
		}

		lineNumber++
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}
