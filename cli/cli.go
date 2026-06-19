package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func Run() error {
	var searchString string
	var directories []string

	cmd := &cli.Command{
		Name:      "skan",
		Usage:     "Scan for a string in a directory",
		UsageText: "skan [options] DIRECTORIES...",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "query",
				Usage:       "string to search for",
				Aliases:     []string{"q"},
				Required:    true,
				Destination: &searchString,
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
			return nil
		},
	}

	return cmd.Run(context.Background(), os.Args)
}
