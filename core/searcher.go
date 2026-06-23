package core

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"sync"
)

func Searcher(args SearcherArgs) error {
	output := make(chan Match)
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
					Query:           args.Query,
					CaseInsensitive: args.CaseInsensitive,
					Invert:          args.Invert,
					Regex:           args.Regex,
					WholeWordsOnly:  args.WholeWordsOnly,
					ContextLines:    args.ContextLines,
					File:            path,
					Output:          output,
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

	for res := range output {
		for _, bC := range res.BeforeContext {
			fmt.Printf("%s-%d-%s\n", bC.FileName, bC.LineNumber, bC.LineText)
		}
		fmt.Printf("%s:%d:%s\n", res.FileName, res.LineNumber, res.LineText)
		for _, aC := range res.AfterContext {
			fmt.Printf("%s-%d-%s\n", aC.FileName, aC.LineNumber, aC.LineText)
		}
		fmt.Println("---")
	}

	return nil
}

func traverse(directory string, jobs chan string) error {
	return filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			jobs <- path
		}

		return nil
	})
}
