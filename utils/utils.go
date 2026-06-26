package utils

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
)

const (
	AnsiReset = "\033[0m"
	AnsiRed   = "\033[31m" // red color
)

func IsBinary(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}

	defer file.Close()
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}
	return bytes.IndexByte(buffer[:n], 0) != -1, nil
}

// IsTTY returns true if standard output is an interactive terminal.
func IsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (fi.Mode() & os.ModeCharDevice) != 0
}

func HighlightLine(line string, spans [][]int) string {
	var buf strings.Builder
	lastIndex := 0

	for _, span := range spans {
		start, end := span[0], span[1]
		buf.WriteString(line[lastIndex:start])

		// Append colored match
		buf.WriteString(AnsiRed)
		buf.WriteString(line[start:end])
		buf.WriteString(AnsiReset)

		lastIndex = end
	}

	buf.WriteString(line[lastIndex:])
	return buf.String()
}

func MatchAny(regexp *regexp.Regexp, str string) bool {
	if regexp != nil {
		return regexp.MatchString(str)
	}
	return false
}
