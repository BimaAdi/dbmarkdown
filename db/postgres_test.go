//go:build postgres

package db

import "testing"

func TestIsPostgresDsn(t *testing.T) {
	tests := []struct {
		dsn  string
		want bool
	}{
		{dsn: "postgres://user:password@localhost/db", want: true},
		{dsn: "postgresql://user:password@localhost/db", want: true},
		{dsn: "sqlite://database.db", want: false},
		{dsn: "postgresqlx://database", want: false},
		{dsn: "", want: false},
	}

	for _, test := range tests {
		if got := isPostgresDsn(test.dsn); got != test.want {
			t.Errorf("isPostgresDsn(%q) = %t, want %t", test.dsn, got, test.want)
		}
	}
}

func TestQueryRejectsInvalidDSN(t *testing.T) {
	if _, err := Query("invalid=dsn", "SELECT 1"); err == nil {
		t.Fatal("Query returned nil error for an invalid DSN")
	}
}
