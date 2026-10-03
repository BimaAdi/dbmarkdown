package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigUsesGivenPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "other.json")
	if err := os.WriteFile(path, []byte(`{"db-sqlite":"sqlite://other.db"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path, "conf:\n```json\n{\"db-sqlite\":\"sqlite://markdown.db\"}\n```\n")
	if err != nil {
		t.Fatal(err)
	}
	if cfg["db-sqlite"] != "sqlite://other.db" {
		t.Fatalf("db-sqlite = %q, want sqlite://other.db", cfg["db-sqlite"])
	}
}

func TestLoadConfigReportsMissingGivenPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	_, err := LoadConfig(path, "conf:\n```json\n{\"db-sqlite\":\"sqlite://markdown.db\"}\n```\n")
	if err == nil {
		t.Fatal("LoadConfig accepted a missing config path")
	}
	if err.Error() != "config "+path+" not found" {
		t.Fatalf("error = %q, want config %s not found", err, path)
	}
}

func TestLoadConfigReadsMarkdownConfSection(t *testing.T) {
	markdown := "# Todos\n\nconf:\n```json\n{\"db-sqlite\":\"sqlite://markdown.db\"}\n```\n\ndb=db-sqlite|name=get\n```sql\nSELECT 1\n```\n"

	cfg, err := LoadConfig("", markdown)
	if err != nil {
		t.Fatal(err)
	}
	if cfg["db-sqlite"] != "sqlite://markdown.db" {
		t.Fatalf("db-sqlite = %q, want sqlite://markdown.db", cfg["db-sqlite"])
	}
}

func TestLoadConfigReadsMarkdownConfInsideExampleFences(t *testing.T) {
	markdown := "```markdown\nconf:\n```json\n{\"db-sqlite\":\"sqlite://ignored.db\"}\n```\n```\n"

	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	path := filepath.Join(dir, defaultConfigPath)
	if err := os.WriteFile(path, []byte(`{"db-sqlite":"sqlite://default.db"}`), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig("", markdown)
	if err != nil {
		t.Fatal(err)
	}
	if cfg["db-sqlite"] != "sqlite://default.db" {
		t.Fatalf("db-sqlite = %q, want sqlite://default.db", cfg["db-sqlite"])
	}
}

func TestLoadConfigMarkdownConfMissingJSONFence(t *testing.T) {
	_, err := LoadConfig("", "conf:\nSELECT 1\n")
	if err == nil {
		t.Fatal("LoadConfig accepted a conf section without a ```json fence")
	}
	if !strings.Contains(err.Error(), "```json") {
		t.Fatalf("error = %q, want missing ```json fence error", err)
	}
}

func TestLoadConfigMarkdownConfUnclosedJSONFence(t *testing.T) {
	_, err := LoadConfig("", "conf:\n```json\n{\"db-sqlite\":\"sqlite://markdown.db\"}\n")
	if err == nil {
		t.Fatal("LoadConfig accepted an unclosed ```json fence")
	}
	if !strings.Contains(err.Error(), "unclosed") {
		t.Fatalf("error = %q, want unclosed fence error", err)
	}
}

func TestLoadConfigFallsBackToDefaultConfigPath(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if err := os.WriteFile(defaultConfigPath, []byte(`{"db-sqlite":"sqlite://default.db"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig("", "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg["db-sqlite"] != "sqlite://default.db" {
		t.Fatalf("db-sqlite = %q, want sqlite://default.db", cfg["db-sqlite"])
	}
}

func TestLoadConfigReportsMissingDefaultConfigPath(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	_, err = LoadConfig("", "")
	if err == nil {
		t.Fatal("LoadConfig accepted a missing default config")
	}
	if err.Error() != "config "+defaultConfigPath+" not found" {
		t.Fatalf("error = %q, want config %s not found", err, defaultConfigPath)
	}
}
