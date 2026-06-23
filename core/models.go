package core

import "regexp"

type Match struct {
	FileName      string
	LineNumber    int
	LineText      string
	BeforeContext []ContextLine
	AfterContext  []ContextLine
}

type ContextLine struct {
	FileName   string
	LineNumber int
	LineText   string
}

type SearcherArgs struct {
	Query           string
	Invert          bool
	CaseInsensitive bool
	Regex           bool
	WholeWordsOnly  bool
	ContextLines    ContextLineBuffer
	Directories     []string
}

type FindArgs struct {
	Query           string
	CaseInsensitive bool
	Invert          bool
	Regex           bool
	WholeWordsOnly  bool
	ContextLines    ContextLineBuffer
	File            string
	Output          chan Match
}

type LineContext struct {
	CurrentLine []byte
	Regexp      *regexp.Regexp
	Args        FindArgs
	LineNumber  int
}

type ContextLineBuffer struct {
	Before  int
	After   int
	Context int
}
