package db

import (
	"strings"
	"testing"
)

func TestQueryRejectsUnsupportedDSN(t *testing.T) {
	_, err := Query("oracle://localhost/database", "SELECT 1")
	if err == nil {
		t.Fatal("Query returned nil error for an unsupported DSN")
	}
	if !strings.Contains(err.Error(), "unsupported database DSN") {
		t.Fatalf("Query error = %q, want unsupported database DSN error", err)
	}
}
