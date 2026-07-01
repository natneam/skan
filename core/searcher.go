package core

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"natneam.com/skan/model"
	"natneam.com/skan/utils"
)

func Searcher(args model.SearcherArgs) error {
	jobs := make(chan model.Job, 100)
	var walkerWg, workerWg sync.WaitGroup

	defer func() {
		walkerWg.Wait()
		close(jobs)
		workerWg.Wait()
		close(args.OutputChan)
		close(args.ErrorChan)
	}()

	if args.ContextLines.Context == -1 {
		args.ContextLines.Context = 0
	}

	if args.ContextLines.Before == -1 {
		args.ContextLines.Before = args.ContextLines.Context
	}
	if args.ContextLines.After == -1 {
		args.ContextLines.After = args.ContextLines.Context
	}

	if args.Depth == nil {
		depth := -1
		args.Depth = &depth
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
		return err
	}

	var includeRegex *regexp.Regexp
	if len(args.Include) != 0 {
		includeRegex, err = regexp.Compile("(" + strings.Join(args.Include, "|") + ")")
		if err != nil {
			return err
		}
	}

	var excludeRegex *regexp.Regexp
	if len(args.Exclude) == 0 {
		excludeRegex = nil
	} else {
		excludeRegex, err = regexp.Compile("(" + strings.Join(args.Exclude, "|") + ")")
		if err != nil {
			return err
		}
	}

	// Calculate Max Size
	var maxFileSize int64 = math.MaxInt64
	if args.MaxSize != "" {
		maxFileSize, err = utils.ParseSize(args.MaxSize, regexp.MustCompile(`(?i)^(\d+)([KMGT]?)B?$`))
		if err != nil {
			return err
		}
	}

	if args.Workers <= 0 {
		return errors.New("workers must be greater than 0")
	}

	for range args.Workers {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for job := range jobs {
				err := Find(model.FindArgs{
					Invert:       args.Invert,
					Regexp:       regex,
					ContextLines: model.ContextLineBuffer{Before: args.ContextLines.Before, After: args.ContextLines.After},
					Job:          job,
					Output:       args.OutputChan,
				})

				if err != nil {
					if args.AbsolutePaths {
						args.ErrorChan <- fmt.Errorf("[%s] %s", job.AbsolutePath, err)
					} else {
						args.ErrorChan <- fmt.Errorf("[%s] %s", job.RelativePath, err)
					}
				}
			}
		}()
	}

	for _, dir := range args.Directories {
		walkerWg.Add(1)
		go func() {
			defer walkerWg.Done()
			traverse(dir, jobs, includeRegex, excludeRegex, args.AbsolutePaths, maxFileSize, *args.Depth, args.ErrorChan)
		}()
	}

	return nil
}

func traverse(directory string, jobs chan model.Job, includeRegex, excludeRegex *regexp.Regexp, absolutePaths bool, maxSize int64, depth int, errChan chan error) {
	filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			errChan <- fmt.Errorf("[%s] %w", path, err)
			return nil
		}

		rel, err := filepath.Rel(directory, path)
		if err != nil {
			errChan <- fmt.Errorf("[%s] %w", path, err)
			return nil
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			errChan <- fmt.Errorf("[%s] %w", path, err)
			return nil
		}

		// File Output
		fileOutput := rel
		if fileOutput == "." {
			fileOutput = path
		}

		if absolutePaths {
			fileOutput = abs
		}

		// Depth check
		dpt := strings.Count(rel, string(filepath.Separator))
		if depth >= 0 {
			if d.IsDir() && rel != "." && dpt >= depth {
				return filepath.SkipDir
			}
			if !d.IsDir() && dpt > depth {
				return nil
			}
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
				binary, err := utils.IsBinary(abs)
				if err != nil {
					errChan <- fmt.Errorf("[%s] %w", fileOutput, err)
					return nil
				}

				if binary {
					return nil
				}

				// Check max size
				info, err := d.Info()
				if err != nil {
					errChan <- fmt.Errorf("[%s] %w", fileOutput, err)
					return nil
				}

				if info.Size() > maxSize {
					return nil
				}

				if absolutePaths {
					jobs <- model.Job{RelativePath: fileOutput, AbsolutePath: abs, IsRelative: false}
				} else {
					jobs <- model.Job{RelativePath: fileOutput, AbsolutePath: abs, IsRelative: true}
				}
			}
		}

		return nil
	})
}
