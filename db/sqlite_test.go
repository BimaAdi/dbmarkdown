//go:build sqlite

package db

import "testing"

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
func TestQuerySQLite(t *testing.T) {
	got, err := Query("sqlite://:memory:", "SELECT 1 AS id, 'todo' AS name, true AS is_done")
	if err != nil {
		t.Fatal(err)
	}

	want := "| id | name | is_done |\n|----|------|---------|\n| 1  | todo | 1       |"
	if got != want {
		t.Fatalf("Query() = %q, want %q", got, want)
	}
}

func TestQuerySQLiteRejectsEmptyPath(t *testing.T) {
	if _, err := Query("sqlite://", "SELECT 1"); err == nil {
		t.Fatal("Query returned nil error for an empty SQLite path")
	}
}
