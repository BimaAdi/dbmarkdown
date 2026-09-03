package main

import (
	"testing"

	"github.com/urfave/cli/v3"
)

func TestRunOutputFileFlag(t *testing.T) {
	runCommand := newCLI().Commands[0]
	flag, ok := runCommand.Flags[0].(*cli.StringFlag)
	if !ok {
		t.Fatalf("run flag type = %T, want *cli.StringFlag", runCommand.Flags[0])
	}
	if flag.Name != "output-file" {
		t.Fatalf("flag name = %q, want output-file", flag.Name)
	}
	if len(flag.Aliases) != 1 || flag.Aliases[0] != "of" {
		t.Fatalf("flag aliases = %v, want [of]", flag.Aliases)
	}
}

func TestRunOutputToFlag(t *testing.T) {
	runCommand := newCLI().Commands[0]
	flag, ok := runCommand.Flags[1].(*cli.StringFlag)
	if !ok {
		t.Fatalf("run flag type = %T, want *cli.StringFlag", runCommand.Flags[1])
	}
	if flag.Name != "output-to" {
		t.Fatalf("flag name = %q, want output-to", flag.Name)
	}
	if len(flag.Aliases) != 1 || flag.Aliases[0] != "ot" {
		t.Fatalf("flag aliases = %v, want [ot]", flag.Aliases)
	}
	if flag.Value != "file" {
		t.Fatalf("flag default = %q, want file", flag.Value)
	}
	if err := flag.Validator("file"); err != nil {
		t.Fatalf("file destination rejected: %v", err)
	}
	if err := flag.Validator("shell"); err != nil {
		t.Fatalf("shell destination rejected: %v", err)
	}
	if err := flag.Validator("invalid"); err == nil {
		t.Fatal("invalid destination was accepted")
	}
}

func TestFindQuery(t *testing.T) {
	markdown := "intro\n\ndb=dev|name=unfinished\n```sql\nSELECT id FROM todos WHERE done = false;\n```\n"

	block, err := findQuery(markdown, "unfinished")
	if err != nil {
		t.Fatal(err)
	}
	if block.connection != "dev" {
		t.Fatalf("connection = %q, want dev", block.connection)
	}
	if block.query != "SELECT id FROM todos WHERE done = false;" {
		t.Fatalf("query = %q", block.query)
	}
	if markdown[block.end:] != "" {
		t.Fatalf("end offset did not preserve trailing content: %q", markdown[block.end:])
	}
}

func TestFindQueryErrorsWhenMissing(t *testing.T) {
	if _, err := findQuery("db=dev|name=other\n```sql\nSELECT 1\n```\n", "missing"); err == nil {
		t.Fatal("findQuery returned nil error for a missing query")
	}
}
