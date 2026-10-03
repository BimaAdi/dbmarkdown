package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestTutorialPrintsDocumentation(t *testing.T) {
	var output bytes.Buffer
	app := newCLI()
	app.Writer = &output
	app.ErrWriter = &output

	if err := app.Run(context.Background(), []string{"dbmarkdown", "tutorial"}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output.Bytes(), tutorial) {
		t.Fatalf("tutorial output does not match embedded documentation")
	}
}

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

func TestRunAppendFlag(t *testing.T) {
	runCommand := newCLI().Commands[0]
	flag, ok := runCommand.Flags[2].(*cli.BoolFlag)
	if !ok {
		t.Fatalf("run flag type = %T, want *cli.BoolFlag", runCommand.Flags[2])
	}
	if flag.Name != "append" {
		t.Fatalf("flag name = %q, want append", flag.Name)
	}
}

func TestCleanRemovesNamedResult(t *testing.T) {
	cleanCommand := newCLI().Commands[2]
	if cleanCommand.Name != "clean" {
		t.Fatalf("command name = %q, want clean", cleanCommand.Name)
	}

	path := t.TempDir() + "/result.md"
	markdown := "db=dev|name=items\n```sql\nSELECT 1\n```\n\nresult:\n---\n| old |\n|-----|\n---\n\ndb=dev|name=other\n```sql\nSELECT 2\n```\n\nresult:\n---\n| old |\n|-----|\n---\n"
	if err := os.WriteFile(path, []byte(markdown), 0644); err != nil {
		t.Fatal(err)
	}

	if err := newCLI().Run(context.Background(), []string{"dbmarkdown", "clean", "items", path}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "db=dev|name=items\n```sql\nSELECT 1\n```\n\ndb=dev|name=other\n```sql\nSELECT 2\n```\n\nresult:\n---\n| old |\n|-----|\n---\n"
	if string(got) != want {
		t.Fatalf("cleaned markdown = %q, want %q", got, want)
	}
}

func TestCleanAllRemovesEveryResult(t *testing.T) {
	cleanAllCommand := newCLI().Commands[3]
	if cleanAllCommand.Name != "cleanall" {
		t.Fatalf("command name = %q, want cleanall", cleanAllCommand.Name)
	}

	path := t.TempDir() + "/result.md"
	markdown := "db=dev|name=items\n```sql\nSELECT 1\n```\n\nresult:\n---\n| old |\n|-----|\n---\n\ndb=dev|name=other\n```sql\nSELECT 2\n```\n\nresult:\n---\n| old |\n|-----|\n---\n\nafter\n"
	if err := os.WriteFile(path, []byte(markdown), 0644); err != nil {
		t.Fatal(err)
	}

	if err := newCLI().Run(context.Background(), []string{"dbmarkdown", "cleanall", path}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "db=dev|name=items\n```sql\nSELECT 1\n```\n\ndb=dev|name=other\n```sql\nSELECT 2\n```\n\nafter\n"
	if string(got) != want {
		t.Fatalf("cleaned markdown = %q, want %q", got, want)
	}
}

func TestRunReadsConfigFromMarkdownConfSection(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	markdownPath := filepath.Join(dir, "todos.md")
	markdown := "conf:\n```json\n{\"db-sqlite\":\"sqlite://:memory:\"}\n```\n\ndb=db-sqlite|name=getall\n```sql\nSELECT 1 AS id\n```\n"
	if err := os.WriteFile(markdownPath, []byte(markdown), 0644); err != nil {
		t.Fatal(err)
	}

	if err := newCLI().Run(context.Background(), []string{"dbmarkdown", "run", "getall", markdownPath}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(markdownPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "result:\n---\n| id |\n|----|\n| 1  |\n---\n") {
		t.Fatalf("run result not written to markdown: %q", got)
	}
}

func TestExecRunsConfiguredQueryAndWritesToShell(t *testing.T) {
	configPath := t.TempDir() + "/db.json"
	if err := os.WriteFile(configPath, []byte(`{"blog":"sqlite://:memory:"}`), 0644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	app := newCLI()
	app.Writer = &output
	app.ErrWriter = &output
	if err := app.Run(context.Background(), []string{
		"dbmarkdown", "--conf", configPath, "exec", "blog", "SELECT 1 AS id",
	}); err != nil {
		t.Fatal(err)
	}

	want := "| id |\n|----|\n| 1  |\n"
	if output.String() != want {
		t.Fatalf("exec output = %q, want %q", output.String(), want)
	}
}

func TestExecRejectsUnknownConfig(t *testing.T) {
	configPath := t.TempDir() + "/db.json"
	if err := os.WriteFile(configPath, []byte(`{"blog":"sqlite://:memory:"}`), 0644); err != nil {
		t.Fatal(err)
	}

	err := newCLI().Run(context.Background(), []string{
		"dbmarkdown", "--conf", configPath, "exec", "missing", "SELECT 1",
	})
	if err == nil {
		t.Fatal("exec accepted an unknown config")
	}
	if err.Error() != `connection "missing" is not configured` {
		t.Fatalf("exec error = %q, want unknown config error", err)
	}
}
