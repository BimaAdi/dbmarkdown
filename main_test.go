package main

import (
	"bytes"
	"context"
	"os"
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
