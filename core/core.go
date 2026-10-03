package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/BimaAdi/dbmarkdown/db"
)

type QueryBlock struct {
	connection string
	name       string
	start      int
	end        int
	query      string
}

func (b QueryBlock) Connection() string { return b.connection }

func (b QueryBlock) Query() string { return b.query }

var markerPattern = regexp.MustCompile(`^\s*db=([^|\s]+)\|name=([^\s]+)\s*$`)
var resultPattern = regexp.MustCompile(`(?m)^[ \t]*result:[ \t]*(?:\r?\n|$)`)

func RunQuery(query string, cfg Config) (string, error) {
	if len(cfg) != 1 {
		return "", errors.New("runQuery requires exactly one configured connection")
	}
	for connection, dsn := range cfg {
		if strings.TrimSpace(dsn) == "" {
			return "", fmt.Errorf("connection %q is not configured", connection)
		}
		result, err := db.Query(dsn, query)
		if err != nil {
			return "", fmt.Errorf("query: %w", err)
		}
		return result, nil
	}
	return "", errors.New("runQuery requires exactly one configured connection")
}

// WriteToFile writes a query result into a Markdown file.
//
// path is the file to update.
// markdown is the file's current content.
// block identifies the SQL query and its position in markdown.
// result is the rendered database result.
// appendMode optionally enables inserting a new result instead of replacing
// the existing one.
func WriteToFile(path, markdown string, block QueryBlock, result string, appendMode ...bool) error {
	resultBlock := "result:\n---\n" + strings.TrimSuffix(result, "\n") + "\n---\n"
	updated := markdown
	if len(appendMode) > 0 && appendMode[0] {
		updated = markdown[:block.end] + "\n" + resultBlock + "\n" + markdown[block.end:]
	} else if start, end, ok := findResultBlock(markdown, block.end); ok {
		updated = markdown[:start] + resultBlock + markdown[end:]
	} else {
		updated = markdown[:block.end] + "\n" + resultBlock + markdown[block.end:]
	}
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}
	return nil
}

// findResultBlock finds the result immediately following a query block.
// markdown is the file content to search, and from is the byte offset after
// the query block. It returns the start and end byte offsets of the result to
// replace, plus false when no result is found. Both the current delimited
// format and the old undelimited format are supported.
func findResultBlock(markdown string, from int) (int, int, bool) {
	tail := markdown[from:]
	match := resultPattern.FindStringIndex(tail)
	if match == nil || strings.TrimSpace(tail[:match[0]]) != "" {
		return 0, 0, false
	}

	start := from + match[0]
	contentStart := from + match[1]
	lines := strings.SplitAfter(markdown[contentStart:], "\n")
	contentOffset := contentStart
	if len(lines) > 0 && strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(lines[0], "\n"), "\r")) == "---" {
		contentOffset += len(lines[0])
		for _, line := range strings.SplitAfter(markdown[contentOffset:], "\n") {
			contentOffset += len(line)
			if strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")) == "---" {
				return start, contentOffset - lineEnding(line), true
			}
		}
		return 0, 0, false
	}

	// Accept the old format so the first run upgrades an existing result in place.
	end := contentStart
	lastContentEnd := contentStart
	lastContentLineEnding := 0
	for _, line := range strings.SplitAfter(markdown[contentStart:], "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"))
		if end != contentStart && (strings.HasPrefix(trimmed, "db=") || trimmed == "result:") {
			break
		}
		end += len(line)
		if trimmed != "" {
			lastContentEnd = end
			lastContentLineEnding = lineEnding(line)
		}
	}
	return start, lastContentEnd - lastContentLineEnding, true
}

func lineEnding(line string) int {
	if strings.HasSuffix(line, "\r\n") {
		return 2
	}
	if strings.HasSuffix(line, "\n") || strings.HasSuffix(line, "\r") {
		return 1
	}
	return 0
}

// trailingNewline returns the length of the newline that starts at offset.
func trailingNewline(markdown string, offset int) int {
	if strings.HasPrefix(markdown[offset:], "\r\n") {
		return 2
	}
	if strings.HasPrefix(markdown[offset:], "\n") || strings.HasPrefix(markdown[offset:], "\r") {
		return 1
	}
	return 0
}

// cleanBlockBounds widens the result bounds by the whitespace separating the
// query block and the result, so cleaning leaves a single separator. end is
// widened by the newline that follows the result.
func cleanBlockBounds(markdown string, blockEnd, start, end int) (int, int) {
	if start > blockEnd && strings.TrimSpace(markdown[blockEnd:start]) == "" {
		start = blockEnd
	}
	return start, end + trailingNewline(markdown, end)
}

// CleanResult removes the result section that follows a named query block.
// path is the file to update and markdown is its current content.
// block identifies the query whose result is removed.
func CleanResult(path, markdown string, block QueryBlock) error {
	start, end, ok := findResultBlock(markdown, block.end)
	if !ok {
		return fmt.Errorf("query %q has no result section", block.name)
	}
	start, end = cleanBlockBounds(markdown, block.end, start, end)
	if err := os.WriteFile(path, []byte(markdown[:start]+markdown[end:]), 0644); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}
	return nil
}

// CleanAllResults removes every result section from the markdown file.
// path is the file to update and markdown is its current content.
func CleanAllResults(path, markdown string) error {
	lines := strings.SplitAfter(markdown, "\n")
	offsets := make([]int, len(lines))
	for i := 1; i < len(lines); i++ {
		offsets[i] = offsets[i-1] + len(lines[i-1])
	}

	// Walk in reverse so removing a result keeps the offsets of the earlier
	// query blocks valid.
	updated := markdown
	for i, line := range slices.Backward(lines) {
		trimmed := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if markerPattern.FindStringSubmatch(trimmed) == nil {
			continue
		}
		blockEnd := -1
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSuffix(strings.TrimSuffix(lines[j], "\n"), "\r") == "```" {
				blockEnd = offsets[j] + len(lines[j])
				break
			}
		}
		if blockEnd == -1 {
			continue
		}
		if start, end, ok := findResultBlock(updated, blockEnd); ok {
			start, end = cleanBlockBounds(updated, blockEnd, start, end)
			updated = updated[:start] + updated[end:]
		}
	}
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}
	return nil
}

func WriteToShell(writer io.Writer, result string) error {
	_, err := fmt.Fprint(writer, result+"\n")
	return err
}

// FindDuplicateName returns true when markdown contains more than one query
// block with the given name.
func FindDuplicateName(markdown, name string) bool {
	seen := false
	for _, line := range strings.SplitAfter(markdown, "\n") {
		trimmed := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		match := markerPattern.FindStringSubmatch(trimmed)
		if match == nil || match[2] != name {
			continue
		}
		if seen {
			return true
		}
		seen = true
	}
	return false
}

func FindQuery(markdown, name string, configs ...Config) (QueryBlock, error) {
	if len(configs) > 1 {
		return QueryBlock{}, errors.New("findQuery accepts at most one config")
	}
	lines := strings.SplitAfter(markdown, "\n")
	offsets := make([]int, len(lines))
	for i := 1; i < len(lines); i++ {
		offsets[i] = offsets[i-1] + len(lines[i-1])
	}
	for i, line := range lines {
		trimmed := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		match := markerPattern.FindStringSubmatch(trimmed)
		if match == nil {
			continue
		}
		if match[2] != name {
			continue
		}
		if len(configs) == 1 {
			dsn, ok := configs[0][match[1]]
			if !ok || strings.TrimSpace(dsn) == "" {
				return QueryBlock{}, fmt.Errorf("connection %q is not configured", match[1])
			}
		}

		fence := "sql"
		if len(configs) == 1 && db.IsRedisDSN(configs[0][match[1]]) {
			fence = "redis"
		} else if len(configs) == 0 && i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "```redis" {
			fence = "redis"
		}
		if i+1 >= len(lines) || strings.TrimSpace(lines[i+1]) != "```"+fence {
			return QueryBlock{}, fmt.Errorf("query %q must be followed by a ```%s code fence", name, fence)
		}
		if i+2 >= len(lines) {
			return QueryBlock{}, fmt.Errorf("query %q has an unclosed code fence", name)
		}
		queryStart := offsets[i+2]
		for j := i + 2; j < len(lines); j++ {
			if strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(lines[j], "\n"), "\r")) == "```" {
				query := markdown[queryStart:offsets[j]]
				query = strings.TrimSuffix(strings.TrimSuffix(query, "\n"), "\r")
				return QueryBlock{connection: match[1], name: name, start: offsets[i], end: offsets[j] + len(lines[j]), query: query}, nil
			}
		}
		return QueryBlock{}, fmt.Errorf("query %q has an unclosed code fence", name)
	}
	return QueryBlock{}, fmt.Errorf("query %q was not found", name)
}
