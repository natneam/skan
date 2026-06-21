package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"natneam.com/skan/core"
)

func Run() error {
	var searchString string
	var directories []string
	var caseInsensitive bool
	var invertResults bool
	var regexpSearch bool
	var wholeWordsOnly bool

	cmd := &cli.Command{
		Name:        "skan",
		Usage:       "A fast, parallel file scanner for searching strings and patterns across directories",
		UsageText:   "skan [options] DIRECTORIES...",
		Description: "skan recursively walks one or more directories and searches every file for a\nquery string, printing each match with its file path and line number. Scanning \nis parallelized across all available CPU cores for speed.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "query",
				Usage:       "The string (or pattern) to search for in file contents (required)",
				Aliases:     []string{"q"},
				Required:    true,
				Destination: &searchString,
			},
			&cli.BoolFlag{
				Name:        "i",
				Usage:       "Perform a case-insensitive match (e.g. \"Foo\" matches \"foo\", \"FOO\")",
				Destination: &caseInsensitive,
			},
			&cli.BoolFlag{
				Name:        "v",
				Usage:       "Invert results — print lines that do NOT contain the query",
				Destination: &invertResults,
			},
			&cli.BoolFlag{
				Name:        "r",
				Usage:       "Treat the query as a regular expression instead of a literal string",
				Destination: &regexpSearch,
			},
			&cli.BoolFlag{
				Name:        "w",
				Usage:       "Match whole words only (e.g. \"cat\" matches \"cat\" but not \"cats\" or \"location\")",
				Destination: &wholeWordsOnly,
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "directories",
				UsageText:   "directories to scan",
				Max:         8,
				Min:         1,
				Destination: &directories,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Printf("Initializing skan for query: %q inside [%s]...\n", searchString, strings.Join(directories, ", "))
			fmt.Println("====================================== Result ======================================")
			return core.Searcher(core.SearcherArgs{
				Query:           searchString,
				CaseInsensitive: caseInsensitive,
				Invert:          invertResults,
				Regex:           regexpSearch,
				WholeWordsOnly:  wholeWordsOnly,
				Directories:     directories,
			})
		},
	}

	return cmd.Run(context.Background(), os.Args)
}
