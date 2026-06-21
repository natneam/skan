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

	cmd := &cli.Command{
		Name:      "skan",
		Usage:     "Scan for a string in a directory",
		UsageText: "skan [options] DIRECTORIES...",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "query",
				Usage:       "String to search for",
				Aliases:     []string{"q"},
				Required:    true,
				Destination: &searchString,
			},
			&cli.BoolFlag{
				Name:        "i",
				Usage:       "Case insensitive search",
				Destination: &caseInsensitive,
			},
			&cli.BoolFlag{
				Name:        "v",
				Usage:       "Invert search results",
				Destination: &invertResults,
			},
			&cli.BoolFlag{
				Name:        "r",
				Usage:       "Regular expression search",
				Destination: &regexpSearch,
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
				Directories:     directories,
			})
		},
	}

	return cmd.Run(context.Background(), os.Args)
}
