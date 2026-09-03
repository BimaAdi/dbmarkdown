package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/BimaAdi/dbmarkdown/db"
	"github.com/urfave/cli/v3"
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
	return newCLI().Run(context.Background(), append([]string{"dbmarkdown"}, args...))
}

func newCLI() *cli.Command {
	return &cli.Command{
		Name:  "dbmarkdown",
		Usage: "execute a named SQL query from a Markdown file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "conf",
				Usage: "path to the database configuration file",
				Value: "db.json",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "run",
				Usage:     "execute a named query",
				ArgsUsage: "<name> <markdown path>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output-file",
						Aliases: []string{"of"},
						Usage:   "path to write the resulting Markdown, default: <markdown path>",
					},
					&cli.StringFlag{
						Name:    "output-to",
						Aliases: []string{"ot"},
						Usage:   "write the resulting output to (options: file/shell)",
						Value:   "file",
						Validator: func(value string) error {
							if value != "file" && value != "shell" {
								return fmt.Errorf("invalid output destination %q: must be file or shell", value)
							}
							return nil
						},
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 2 {
						return errors.New("usage: dbmarkdown [--conf db.json] run <name> <markdown path> [--output-file path] [--output-to file|shell]")
					}

					name, markdownPath := cmd.Args().Get(0), cmd.Args().Get(1)
					outputPath := cmd.String("output-file")
					if outputPath == "" {
						outputPath = markdownPath
					}

					cfg, err := loadConfig(cmd.String("conf"))
					if err != nil {
						return err
					}

					input, err := os.ReadFile(markdownPath)
					if err != nil {
						return fmt.Errorf("read markdown: %w", err)
					}

					block, err := findQuery(string(input), name, cfg)
					if err != nil {
						return err
					}

					queryCfg := config{block.connection: cfg[block.connection]}
					result, err := runQuery(block.query, queryCfg)
					if err != nil {
						return err
					}

					if cmd.String("output-to") == "shell" {
						return writeToShell(cmd.Writer, result)
					}
					return writeToFile(outputPath, string(input), block, result)
				},
			},
		},
	}
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

func runQuery(query string, cfg config) (string, error) {
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

func writeToFile(path, markdown string, block queryBlock, result string) error {
	updated := markdown[:block.end] + "\nresult:\n" + result + markdown[block.end:]
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}
	return nil
}

func writeToShell(writer io.Writer, result string) error {
	_, err := fmt.Fprint(writer, result + "\n")
	return err
}

func findQuery(markdown, name string, configs ...config) (queryBlock, error) {
	if len(configs) > 1 {
		return queryBlock{}, errors.New("findQuery accepts at most one config")
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
				return queryBlock{}, fmt.Errorf("connection %q is not configured", match[1])
			}
		}
		if i+1 >= len(lines) || strings.TrimSpace(lines[i+1]) != "```sql" {
			return queryBlock{}, fmt.Errorf("query %q must be followed by a ```sql code fence", name)
		}
		if i+2 >= len(lines) {
			return queryBlock{}, fmt.Errorf("query %q has an unclosed code fence", name)
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
