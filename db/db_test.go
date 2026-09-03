package db

import (
	"strings"
	"testing"
)

func TestIsSQLiteDSN(t *testing.T) {
	tests := []struct {
		dsn  string
		want bool
	}{
		{dsn: "sqlite://:memory:", want: true},
		{dsn: "sqlite://database.db", want: true},
		{dsn: "postgres://database", want: false},
		{dsn: "SQLite://database.db", want: false},
		{dsn: "", want: false},
	}

	for _, test := range tests {
		if got := isSQLiteDSN(test.dsn); got != test.want {
			t.Errorf("isSQLiteDSN(%q) = %t, want %t", test.dsn, got, test.want)
		}
	}
}

func TestQueryRejectsUnsupportedDSN(t *testing.T) {
	_, err := Query("mysql://localhost/database", "SELECT 1")
	if err == nil {
		t.Fatal("Query returned nil error for an unsupported DSN")
	}
	if !strings.Contains(err.Error(), "unsupported database DSN") {
		t.Fatalf("Query error = %q, want unsupported database DSN error", err)
	}
}
