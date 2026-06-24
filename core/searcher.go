package core

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"

	"natneam.com/skan/model"
	"natneam.com/skan/utils"
)

func Searcher(args model.SearcherArgs) (chan model.Match, error) {
	output := make(chan model.Match, 100)
	jobs := make(chan string, 100)

	var walkerWg sync.WaitGroup
	var workerWg sync.WaitGroup

	if args.ContextLines.Context == -1 {
		args.ContextLines.Context = 0
	}

	if args.ContextLines.Before == -1 {
		args.ContextLines.Before = args.ContextLines.Context
	}
	if args.ContextLines.After == -1 {
		args.ContextLines.After = args.ContextLines.Context
	}

	query := []byte(args.Query)

	// Preprocess the Query
	if !args.Regex {
		query = []byte(regexp.QuoteMeta(string(query)))
	}
	if args.WholeWordsOnly {
		query = []byte("\\b" + string(query) + "\\b")
	}
	if args.CaseInsensitive {
		query = []byte("(?i)" + string(query))
	}
	regex, err := regexp.Compile(string(query))
	if err != nil {
		return nil, err
	}

	for range runtime.NumCPU() {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for path := range jobs {
				Find(model.FindArgs{
					Invert:        args.Invert,
					Regexp:        regex,
					ContextLines:  model.ContextLineBuffer{Before: args.ContextLines.Before, After: args.ContextLines.After},
					File:          path,
					Output:        output,
					AbsolutePaths: args.AbsolutePaths,
				})
			}
		}()
	}

	for _, dir := range args.Directories {
		walkerWg.Add(1)
		go func() {
			defer walkerWg.Done()
			traverse(dir, jobs, args.AbsolutePaths)
		}()
	}

	go func() {
		walkerWg.Wait()
		close(jobs)
		workerWg.Wait()
		close(output)
	}()

	return output, nil
}

func traverse(directory string, jobs chan string, absolutePaths bool) error {
	return filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// read the first 512 bytes to see if it's a binary file, if so discard it
			binary, err := utils.IsBinary(path)
			if err != nil || binary {
				return nil
			}

			if absolutePaths {
				abs, err := filepath.Abs(path)
				if err != nil {
					return nil
				}
				jobs <- abs
			} else {
				jobs <- path
			}
		}

		return nil
	})
}
