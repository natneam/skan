package core

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
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
				find(FindArgs{
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

func find(args FindArgs) error {
	f, err := os.Open(args.File)
	if err != nil {
		return err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	lineNumber := 1

	for scanner.Scan() {
		originalLine := scanner.Bytes()
		line := originalLine
		query := []byte(args.Query)

		if args.CaseInsensitive {
			line = bytes.ToLower(line)
			query = bytes.ToLower(query)
		}

		if args.Invert && !bytes.Contains(line, query) {
			args.Output <- Match{FileName: args.File, LineNumber: lineNumber, LineText: string(originalLine)}
		} else if !args.Invert && bytes.Contains(line, query) {
			args.Output <- Match{FileName: args.File, LineNumber: lineNumber, LineText: string(originalLine)}
		}
		lineNumber++
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}
