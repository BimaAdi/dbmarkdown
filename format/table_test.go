package format

import (
	"testing"
)

func TestMarkdownTableAlignsColumns(t *testing.T) {
	header := []string{"id", "name", "status"}
	data := [][]string{{"1", "a longer name", "ok"}, {"20", "x", "not ok"}}
	got := MarkdownTable(header, data)
	want := "| id | name          | status |\n|----|---------------|--------|\n| 1  | a longer name | ok     |\n| 20 | x             | not ok |"
	if got != want {
		t.Fatalf("MarkdownTable() = %q, want %q", got, want)
	}
}

func TestMarkdownTableEscapesCells(t *testing.T) {
	got := MarkdownTable([]string{"value"}, [][]string{{"a|b\nc"}})
	want := "| value     |\n|-----------|\n| a\\|b<br>c |"
	if got != want {
		t.Fatalf("MarkdownTable() = %q, want %q", got, want)
	}
}
