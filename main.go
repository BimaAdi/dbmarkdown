package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/BimaAdi/dbmarkdown/db"
)

type config map[string]string

type queryBlock struct {
	connection string
	name       string
	start      int
	end        int
	query      string
}

var markerPattern = regexp.MustCompile(`^\s*db=([^|\s]+)\|name=([^\s]+)\s*$`)

func main() {
	if err := runCLI(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "dbmarkdown:", err)
		os.Exit(1)
	}
}

func runCLI(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: dbmarkdown [--conf db.json] run <name> <markdown path> [--output path]")
	}

	confPath := "db.json"
	outputPath := ""
	for len(args) > 0 && strings.HasPrefix(args[0], "--") {
		switch args[0] {
		case "--conf":
			if len(args) < 2 {
				return errors.New("--conf requires a path")
			}
			confPath, args = args[1], args[2:]
		case "--output":
			if len(args) < 2 {
				return errors.New("--output requires a path")
			}
			outputPath, args = args[1], args[2:]
		default:
			return fmt.Errorf("unknown option %q", args[0])
		}
	}
	if len(args) == 0 || args[0] != "run" {
		return errors.New("usage: dbmarkdown [--conf db.json] run <name> <markdown path> [--output path]")
	}

	commandArgs := args[1:]
	// Accept options after the command as well, which is convenient with go run.
	positional := make([]string, 0, 2)
	for i := 0; i < len(commandArgs); i++ {
		switch commandArgs[i] {
		case "--conf", "--output":
			if i+1 >= len(commandArgs) {
				return fmt.Errorf("%s requires a path", commandArgs[i])
			}
			if commandArgs[i] == "--conf" {
				confPath = commandArgs[i+1]
			} else {
				outputPath = commandArgs[i+1]
			}
			i++
		default:
			positional = append(positional, commandArgs[i])
		}
	}
	if len(positional) != 2 {
		return errors.New("usage: dbmarkdown [--conf db.json] run <name> <markdown path> [--output path]")
	}
	if outputPath == "" {
		outputPath = positional[1]
	}

	cfg, err := loadConfig(confPath)
	if err != nil {
		return err
	}
	input, err := os.ReadFile(positional[1])
	if err != nil {
		return fmt.Errorf("read markdown: %w", err)
	}
	result, err := executeNamedQuery(string(input), positional[0], cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, []byte(result), 0644); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}
	return nil
}

func loadConfig(path string) (config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func executeNamedQuery(markdown, name string, cfg config) (string, error) {
	block, err := findQuery(markdown, name)
	if err != nil {
		return "", err
	}
	dsn, ok := cfg[block.connection]
	if !ok || strings.TrimSpace(dsn) == "" {
		return "", fmt.Errorf("connection %q is not configured", block.connection)
	}
	table, err := db.Query(dsn, block.query)
	if err != nil {
		return "", fmt.Errorf("query %q: %w", name, err)
	}
	return markdown[:block.end] + "\nresult:\n" + table + markdown[block.end:], nil
}

func findQuery(markdown, name string) (queryBlock, error) {
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
		if i+1 >= len(lines) || strings.TrimSpace(lines[i+1]) != "```sql" {
			return queryBlock{}, fmt.Errorf("query %q must be followed by a ```sql code fence", name)
		}
		queryStart := offsets[i+2]
		for j := i + 2; j < len(lines); j++ {
			if strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(lines[j], "\n"), "\r")) == "```" {
				query := markdown[queryStart:offsets[j]]
				query = strings.TrimSuffix(strings.TrimSuffix(query, "\n"), "\r")
				return queryBlock{connection: match[1], name: name, start: offsets[i], end: offsets[j] + len(lines[j]), query: query}, nil
			}
		}
		return queryBlock{}, fmt.Errorf("query %q has an unclosed code fence", name)
	}
	return queryBlock{}, fmt.Errorf("query %q was not found", name)
}
