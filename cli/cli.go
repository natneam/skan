package cli

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
	"natneam.com/skan/cli/output"
	"natneam.com/skan/core"
	"natneam.com/skan/model"
)

func Run() error {
	var searchString string
	var directories []string
	var caseInsensitive bool
	var invertResults bool
	var regexpSearch bool
	var wholeWordsOnly bool
	var contextLinesInput model.ContextLineBuffer
	var colorOutput bool
	var jsonOutput bool
	var countMode bool

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
			&cli.IntFlag{
				Name:        "B",
				Usage:       "Print N lines of leading context before matching lines",
				DefaultText: "0",
				Value:       -1,
				Destination: &contextLinesInput.Before,
			},
			&cli.IntFlag{
				Name:        "A",
				Usage:       "Print N lines of trailing context after matching lines",
				DefaultText: "0",
				Value:       -1,
				Destination: &contextLinesInput.After,
			},
			&cli.IntFlag{
				Name:        "C",
				Usage:       "Print N lines of context before and after matching lines",
				DefaultText: "0",
				Value:       -1,
				Destination: &contextLinesInput.Context,
			},
			&cli.BoolFlag{
				Name:        "color",
				Usage:       "Colorize matching text in text output, doesn't affect JSON output",
				Destination: &colorOutput,
			},
			&cli.BoolFlag{
				Name:        "json",
				Usage:       "Output results as newline-delimited JSON (one JSON object per match)",
				Destination: &jsonOutput,
			},
			&cli.BoolFlag{
				Name:        "count",
				Aliases:     []string{"c"},
				Usage:       "Output the number of matches instead of the matching lines",
				Destination: &countMode,
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
			outputData, err := core.Searcher(model.SearcherArgs{
				Query:           searchString,
				CaseInsensitive: caseInsensitive,
				Invert:          invertResults,
				Regex:           regexpSearch,
				WholeWordsOnly:  wholeWordsOnly,
				ContextLines:    contextLinesInput,
				Directories:     directories,
			})

			if err != nil {
				return err
			}

			if countMode {
				output.EmitCount(outputData)
			} else if jsonOutput {
				output.EmitJSON(outputData)
			} else {
				output.EmitText(outputData, colorOutput)
			}
			return nil
		},
	}

	return cmd.Run(context.Background(), os.Args)
}
