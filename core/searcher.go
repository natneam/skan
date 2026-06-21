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

	for range runtime.NumCPU() {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for path := range jobs {
				Find(FindArgs{
					Query:           args.Query,
					CaseInsensitive: args.CaseInsensitive,
					Invert:          args.Invert,
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
		fmt.Println(res.FileName, ": ", res.LineNumber)
		fmt.Println("  ", res.LineText)
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
