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

type Job struct {
	RelativePath string
	AbsolutePath string
	IsRelative   bool
}

type SearcherArgs struct {
	CaseInsensitive bool
	ContextLines    ContextLineBuffer
	Invert          bool
	Query           string
	Regex           bool
	WholeWordsOnly  bool
	AbsolutePaths   bool
	Directories     []string
	Include         []string
	Exclude         []string
	MaxSize         string
	Workers         int
	Depth           *int
	OutputChan      chan Match
	ErrorChan       chan error
}

type FindArgs struct {
	ContextLines ContextLineBuffer
	Invert       bool
	Job          Job
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
