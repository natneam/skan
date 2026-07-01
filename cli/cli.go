package cli

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/urfave/cli/v3"
	"natneam.com/skan/cli/output"
	"natneam.com/skan/core"
	"natneam.com/skan/model"
	"natneam.com/skan/utils"
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
	var absolutePaths bool
	var include []string
	var exclude []string
	var maxSize string
	var workers int
	var depth int

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
			&cli.BoolFlag{
				Name:        "absolute",
				Usage:       "Output absolute paths for matching files",
				DefaultText: "relative",
				Destination: &absolutePaths,
			},
			&cli.StringSliceFlag{
				Name:        "include",
				Usage:       "Only include files and directories with relative paths that match the given regex pattern",
				Destination: &include,
			},
			&cli.StringSliceFlag{
				Name:        "exclude",
				Usage:       "Exclude files and directories with relative paths that match the given regex pattern",
				Destination: &exclude,
			},
			&cli.StringFlag{
				Name:        "max-size",
				Usage:       "Maximum file size to include in the search (e.g. 100MB)",
				Destination: &maxSize,
			},
			&cli.IntFlag{
				Name:        "workers",
				Usage:       "Number of workers to use for parallel search",
				DefaultText: "The number of CPUs",
				Value:       runtime.NumCPU(),
				Destination: &workers,
			},
			&cli.IntFlag{
				Name:        "max-depth",
				Usage:       "Maximum depth of directories to scan, negative values are unlimited",
				DefaultText: "-1 (unlimited)",
				Value:       -1,
				Destination: &depth,
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
			var errChan = make(chan error)
			var outputChan = make(chan model.Match)
			var errorCount int
			var errorOutputfile string
			var wg sync.WaitGroup
			var mu = &sync.Mutex{}
			mu.Lock()

			// Write errors to the local file
			wg.Add(1)
			go func() {
				var fileCreationAttempted bool
				var outputFile *os.File
				var err error

				defer wg.Done()
				defer mu.Unlock()

				for val := range errChan {
					if !fileCreationAttempted {
						errorOutputfile, outputFile, err = utils.CreateLogFile("")
						if err == nil {
							defer outputFile.Close()
						}
						fileCreationAttempted = true
					}

					if err == nil && val != nil {
						fmt.Fprintln(outputFile, val)
					}

					errorCount++
				}
			}()

			// Write output to stdout
			wg.Add(1)
			go func() {
				if countMode {
					output.EmitCount(outputChan)
				} else if jsonOutput {
					output.EmitJSON(outputChan)
				} else {
					output.EmitText(outputChan, colorOutput)
				}
				mu.Lock()
				if errorCount > 0 {
					output.EmitErrorInfo(errorCount, errorOutputfile)
				}
				mu.Unlock()
				wg.Done()
			}()

			err := core.Searcher(model.SearcherArgs{
				Query:           searchString,
				CaseInsensitive: caseInsensitive,
				Invert:          invertResults,
				Regex:           regexpSearch,
				WholeWordsOnly:  wholeWordsOnly,
				AbsolutePaths:   absolutePaths,
				ContextLines:    contextLinesInput,
				Directories:     directories,
				Exclude:         exclude,
				Include:         include,
				MaxSize:         maxSize,
				Workers:         workers,
				Depth:           &depth,
				OutputChan:      outputChan,
				ErrorChan:       errChan,
			})

			wg.Wait()
			return err
		},
	}

	return cmd.Run(context.Background(), os.Args)
}
