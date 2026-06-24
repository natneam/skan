package core

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"sync"

	"natneam.com/skan/utils"
)

func Searcher(args SearcherArgs) (chan Match, error) {
	output := make(chan Match, 100)
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

	for range runtime.NumCPU() {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for path := range jobs {
				Find(FindArgs{
					SearchOptions: SearchOptions{
						Query:           args.Query,
						CaseInsensitive: args.CaseInsensitive,
						Invert:          args.Invert,
						Regex:           args.Regex,
						WholeWordsOnly:  args.WholeWordsOnly,
						ContextLines:    ContextLineBuffer{Before: args.ContextLines.Before, After: args.ContextLines.After},
					},
					File:   path,
					Output: output,
				})
			}
		}()
	}

	for _, dir := range args.Directories {
		walkerWg.Add(1)
		go func() {
			defer walkerWg.Done()
			traverse(dir, jobs)
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

func traverse(directory string, jobs chan string) error {
	return filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// read the first 512 bytes to see if it's a binary file, if so discard it
			binary, err := utils.IsBinary(path)
			if err != nil || binary {
				return filepath.SkipDir
			}

			jobs <- path
		}

		return nil
	})
}
