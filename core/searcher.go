package core

import (
	"errors"
	"io/fs"
	"math"
	"path/filepath"
	"regexp"
	"strings"
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

	var includeRegex *regexp.Regexp
	if len(args.Include) != 0 {
		includeRegex, err = regexp.Compile("(" + strings.Join(args.Include, "|") + ")")
		if err != nil {
			return nil, err
		}
	}

	var excludeRegex *regexp.Regexp
	if len(args.Exclude) == 0 {
		excludeRegex = nil
	} else {
		excludeRegex, err = regexp.Compile("(" + strings.Join(args.Exclude, "|") + ")")
		if err != nil {
			return nil, err
		}
	}

	// Calculate Max Size
	var maxFileSize int64 = math.MaxInt64
	if args.MaxSize != "" {
		maxFileSize, err = utils.ParseSize(args.MaxSize, regexp.MustCompile(`(?i)^(\d+)([KMGT]?)B?$`))
		if err != nil {
			return nil, err
		}
	}

	if args.Workers <= 0 {
		return nil, errors.New("workers must be greater than 0")
	}

	for range args.Workers {
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
			traverse(dir, jobs, includeRegex, excludeRegex, args.AbsolutePaths, maxFileSize)
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

func traverse(directory string, jobs chan string, includeRegex, excludeRegex *regexp.Regexp, absolutePaths bool, maxSize int64) error {
	return filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(directory, path)
		if err != nil {
			return nil
		}

		// Match exclude
		if excludeRegex != nil {
			if utils.MatchAny(excludeRegex, rel) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if !d.IsDir() {
			// Match include
			if includeRegex == nil || utils.MatchAny(includeRegex, rel) {
				// read the first 512 bytes to see if it's a binary file, if so discard it
				binary, err := utils.IsBinary(path)
				if err != nil || binary {
					return nil
				}

				// Check max size
				info, err := d.Info()
				if err == nil && info.Size() > maxSize {
					return nil
				}

				abs, err := filepath.Abs(path)
				if err != nil {
					return nil
				}

				if absolutePaths {
					jobs <- abs
				} else {
					jobs <- rel
				}
			}
		}

		return nil
	})
}
