package format

import "testing"

func TestText(t *testing.T) {
	got := Text([]string{"OK", "hello", "foo"})
	if got != "```\nOK\nhello\nfoo\n```\n" {
		t.Fatalf("Text() = %q, want %q", got, "OK\nhello\nfoo")
	}
}

func TestTextEmpty(t *testing.T) {
	if got := Text(nil); got != "```\n\n```\n" {
		t.Fatalf("Text(nil) = %q, want empty text", got)
	}
}
