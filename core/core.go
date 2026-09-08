package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/BimaAdi/dbmarkdown/db"
)

type Config map[string]string

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

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

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

func WriteToShell(writer io.Writer, result string) error {
	_, err := fmt.Fprint(writer, result+"\n")
	return err
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
		if i+1 >= len(lines) || strings.TrimSpace(lines[i+1]) != "```sql" {
			return QueryBlock{}, fmt.Errorf("query %q must be followed by a ```sql code fence", name)
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
