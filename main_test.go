package main

import "testing"

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
