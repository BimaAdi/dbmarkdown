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

func WriteToFile(path, markdown string, block QueryBlock, result string) error {
	updated := markdown[:block.end] + "\nresult:\n" + result + markdown[block.end:]
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}
	return nil
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
