package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/BimaAdi/dbmarkdown/core"
	"github.com/urfave/cli/v3"
)

//go:embed docs/How-to-use.md
var tutorial []byte

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
		Name:  "dbmd",
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
				Usage:     "execute a named query on markdown file",
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
					&cli.BoolFlag{
						Name:  "append",
						Usage: "append a new result instead of replacing the existing result",
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 2 {
						return errors.New("usage: dbmd [--conf db.json] run <name> <markdown path> [--output-file path] [--output-to file|shell] [--append]")
					}

					name, markdownPath := cmd.Args().Get(0), cmd.Args().Get(1)
					outputPath := cmd.String("output-file")
					if outputPath == "" {
						outputPath = markdownPath
					}

					cfg, err := core.LoadConfig(cmd.String("conf"))
					if err != nil {
						return err
					}

					input, err := os.ReadFile(markdownPath)
					if err != nil {
						return fmt.Errorf("read markdown: %w", err)
					}

					block, err := core.FindQuery(string(input), name, cfg)
					if err != nil {
						return err
					}

					queryCfg := core.Config{block.Connection(): cfg[block.Connection()]}
					result, err := core.RunQuery(block.Query(), queryCfg)
					if err != nil {
						return err
					}

					if cmd.String("output-to") == "shell" {
						return core.WriteToShell(cmd.Writer, result)
					}
					return core.WriteToFile(outputPath, string(input), block, result, cmd.Bool("append"))
				},
			},
			{
				Name:      "exec",
				Usage:     "execute a db query directly using existing config",
				ArgsUsage: "<config name> <query>",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.NArg() != 2 {
						return errors.New("usage: dbmd [--conf db.json] exec <config name> <query>")
					}

					configName, query := cmd.Args().Get(0), cmd.Args().Get(1)
					cfg, err := core.LoadConfig(cmd.String("conf"))
					if err != nil {
						return err
					}

					dsn, ok := cfg[configName]
					if !ok || strings.TrimSpace(dsn) == "" {
						return fmt.Errorf("connection %q is not configured", configName)
					}

					result, err := core.RunQuery(query, core.Config{configName: dsn})
					if err != nil {
						return err
					}
					return core.WriteToShell(cmd.Writer, result)
				},
			},
			{
				Name:  "tutorial",
				Usage: "show the usage tutorial",
				Action: func(_ context.Context, cmd *cli.Command) error {
					_, err := cmd.Writer.Write(tutorial)
					return err
				},
			},
		},
	}
}
