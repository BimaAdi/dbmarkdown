//go:build mysql

package db

import "testing"

func TestIsMySQLDSN(t *testing.T) {
	tests := []struct {
		dsn  string
		want bool
	}{
		{dsn: "mysql://user:password@localhost:3306/database", want: true},
		{dsn: "mysql://localhost/database", want: true},
		{dsn: "mysqlx://localhost/database", want: false},
		{dsn: "postgresql://localhost/database", want: false},
		{dsn: "", want: false},
	}

	for _, test := range tests {
		if got := isMySQLDSN(test.dsn); got != test.want {
			t.Errorf("isMySQLDSN(%q) = %t, want %t", test.dsn, got, test.want)
		}
	}
}

func TestMySQLNativeDSN(t *testing.T) {
	got, err := mysqlNativeDSN("mysql://user:p%40ss@db.example:3306/todos?parseTime=true")
	if err != nil {
		t.Fatal(err)
	}

	want := "user:p@ss@tcp(db.example:3306)/todos?allowNativePasswords=false&checkConnLiveness=false&maxAllowedPacket=0&parseTime=true"
	if got != want {
		t.Fatalf("mysqlNativeDSN() = %q, want %q", got, want)
	}
}

func TestMySQLNativeDSNRejectsInvalidURL(t *testing.T) {
	if _, err := mysqlNativeDSN("mysql://"); err == nil {
		t.Fatal("mysqlNativeDSN returned nil error for an invalid URL")
	}
}
