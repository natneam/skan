package core

type Match struct {
	FileName   string
	LineNumber int
	LineText   string
}

type SearcherArgs struct {
	Query           string
	Invert          bool
	CaseInsensitive bool
	Directories     []string
}

type FindArgs struct {
	Query           string
	CaseInsensitive bool
	Invert          bool
	File            string
	Output          chan Match
}

type LineContext struct {
	OriginalLine []byte
	CurrentLine  []byte
	Query        []byte
	Args         FindArgs
}
