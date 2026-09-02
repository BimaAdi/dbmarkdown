package db

import "testing"

func TestQueryRejectsInvalidDSN(t *testing.T) {
	if _, err := Query("invalid=dsn", "SELECT 1"); err == nil {
		t.Fatal("Query returned nil error for an invalid DSN")
	}
}
