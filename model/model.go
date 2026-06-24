package model

import "regexp"

type Match struct {
	FileName      string        `json:"file_name"`
	LineNumber    int           `json:"line_number"`
	LineText      string        `json:"line_text"`
	BeforeContext []ContextLine `json:"before_context"`
	AfterContext  []ContextLine `json:"after_context"`
	MatchIndexes  [][]int       `json:"-"`
}

type ContextLine struct {
	FileName   string `json:"-"`
	LineNumber int    `json:"line_number"`
	LineText   string `json:"line_text"`
}

type SearcherArgs struct {
	CaseInsensitive bool
	ContextLines    ContextLineBuffer
	Invert          bool
	Query           string
	Regex           bool
	WholeWordsOnly  bool
	Directories     []string
}

type FindArgs struct {
	ContextLines ContextLineBuffer
	Invert       bool
	File         string
	Regexp       *regexp.Regexp
	Output       chan Match
}

type LineContext struct {
	CurrentLine []byte
	Args        FindArgs
}

type ContextLineBuffer struct {
	Before  int
	After   int
	Context int
}
