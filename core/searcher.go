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

func Searcher(query string, directories ...string) error {
	output := make(chan Match)
	jobs := make(chan string, 100)

	var walkerWg sync.WaitGroup
	var workerWg sync.WaitGroup

	for range runtime.NumCPU() {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for path := range jobs {
				find(query, path, output)
			}
		}()
	}

	for _, dir := range directories {
		walkerWg.Add(1)
		go func() {
			defer walkerWg.Done()
			Traverse(dir, jobs)
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

func Traverse(directory string, jobs chan string) error {
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

func find(query string, file string, output chan Match) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.Contains(line, []byte(query)) {
			output <- Match{FileName: file, LineNumber: lineNumber, LineText: string(line)}
		}
		lineNumber++
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}
