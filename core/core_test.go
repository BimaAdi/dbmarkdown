package core

import (
	"os"
	"testing"
)

func TestFindQuery(t *testing.T) {
	markdown := "intro\n\ndb=dev|name=unfinished\n```sql\nSELECT id FROM todos WHERE done = false;\n```\n"

	block, err := FindQuery(markdown, "unfinished")
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
	if _, err := FindQuery("db=dev|name=other\n```sql\nSELECT 1\n```\n", "missing"); err == nil {
		t.Fatal("findQuery returned nil error for a missing query")
	}
}

func TestWriteToFileReplacesDelimitedResult(t *testing.T) {
	path := t.TempDir() + "/result.md"
	markdown := "db=dev|name=items\n```sql\nSELECT 1\n```\nresult:\n---\n| old |\n|-----|\n---\n\nnext\n"
	block, err := FindQuery(markdown, "items")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteToFile(path, markdown, block, "| new |\n|-----|", false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "db=dev|name=items\n```sql\nSELECT 1\n```\nresult:\n---\n| new |\n|-----|\n---\n\n\nnext\n"
	if string(got) != want {
		t.Fatalf("updated markdown = %q, want %q", got, want)
	}
}

func TestWriteToFileAppendsDelimitedResult(t *testing.T) {
	path := t.TempDir() + "/result.md"
	markdown := "db=dev|name=items\n```sql\nSELECT 1\n```\nresult:\n---\n| old |\n|-----|\n---\n"
	block, err := FindQuery(markdown, "items")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteToFile(path, markdown, block, "| new |", true); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "db=dev|name=items\n```sql\nSELECT 1\n```\n\nresult:\n---\n| new |\n---\n\nresult:\n---\n| old |\n|-----|\n---\n"
	if string(got) != want {
		t.Fatalf("updated markdown = %q, want %q", got, want)
	}
}

func TestWriteToFileReplacesLegacyResult(t *testing.T) {
	path := t.TempDir() + "/result.md"
	markdown := "db=dev|name=items\n```sql\nSELECT 1\n```\nresult:\n| old |\n|-----|\n\ndb=dev|name=other\n"
	block, err := FindQuery(markdown, "items")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteToFile(path, markdown, block, "| new |", false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "db=dev|name=items\n```sql\nSELECT 1\n```\nresult:\n---\n| new |\n---\n\n\ndb=dev|name=other\n"
	if string(got) != want {
		t.Fatalf("updated markdown = %q, want %q", got, want)
	}
}
