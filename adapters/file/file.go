package file

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

// FileContent represents the content of a file.
type FileContent struct {
	Path    string
	Content []byte
	Error   error
}

// Convert file pattern to regex (support * wildcard)
func patternToRegex(pattern string) string {
	// Escape special regex characters except *
	escaped := regexp.QuoteMeta(pattern)
	// Replace escaped * with .*
	wildcard := strings.ReplaceAll(escaped, `\*`, ".*")
	return "^" + wildcard + "$"
}

// Find files matching any of the patterns and return both path and filename with spaces replaced by underscores
func findMatchingFiles(root string, patterns []string) ([]string, error) {
	var matches []string
	compiledPatterns := make([]*regexp.Regexp, len(patterns))

	// Compile all patterns to regex
	for i, pattern := range patterns {
		re, err := regexp.Compile(patternToRegex(pattern))
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}
		compiledPatterns[i] = re
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		filename := filepath.Base(path)
		for _, re := range compiledPatterns {
			if re.MatchString(filename) {
				// Add both the original path and the filename with spaces replaced by underscores
				matches = append(matches, path, strings.ReplaceAll(filename, " ", "_"))
				break
			}
		}
		return nil
	})

	return matches, err
}

// Read file contents with concurrency
func readFilesConcurrently(paths []string) <-chan FileContent {
	resultChan := make(chan FileContent)
	var wg sync.WaitGroup

	for _, path := range paths {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			file, err := os.Open(filepath.Clean(filePath))
			if err != nil {
				resultChan <- FileContent{Path: filePath, Error: err}
				return
			}
			defer func() {
				if err := file.Close(); err != nil {
					helpers.Println(constant.ERROR, fmt.Sprintf("Error closing file: %v", err))
				}
			}()

			var content []byte
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				content = append(content, scanner.Bytes()...)
				content = append(content, '\n')
			}

			if err := scanner.Err(); err != nil {
				resultChan <- FileContent{Path: filePath, Error: err}
				return
			}

			resultChan <- FileContent{
				Path:    filePath,
				Content: content,
			}
		}(path)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}

// GetFilesContent locates files under rootDir that match any of the provided patterns and returns a channel streaming their contents.
// 
// It returns a receive-only channel that yields a FileContent for each matched file (the Path, Content, and any Error encountered while reading).
// An error is returned if pattern compilation or file discovery fails, or if no files match the supplied patterns.
func GetFilesContent(rootDir string, patterns []string) (<-chan FileContent, error) {
	matchedFiles, err := findMatchingFiles(rootDir, patterns)
	if err != nil {
		return nil, fmt.Errorf("error finding files: %w", err)
	}

	if len(matchedFiles) == 0 {
		return nil, fmt.Errorf("no files matching patterns: %v", patterns)
	}

	return readFilesConcurrently(matchedFiles), nil
}

// NormalizeFileName trims leading and trailing whitespace from name and replaces all spaces with underscores.
// It returns the normalized filename.
func NormalizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	return name
}