package core

import "regexp"

type Match struct {
	FileName      string
	LineNumber    int
	LineText      string
	BeforeContext []ContextLine
	AfterContext  []ContextLine
	MatchIndexes  [][]int
}

type ContextLine struct {
	FileName   string
	LineNumber int
	LineText   string
}

type SearchOptions struct {
	CaseInsensitive bool
	ContextLines    ContextLineBuffer
	Invert          bool
	Query           string
	Regex           bool
	WholeWordsOnly  bool
}

type SearcherArgs struct {
	SearchOptions
	Directories []string
}

type FindArgs struct {
	SearchOptions
	File   string
	Output chan Match
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
