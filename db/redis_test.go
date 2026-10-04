//go:build redis

package db

import "testing"

func TestRedisTextValues(t *testing.T) {
	got := redisTextValues([]any{"hello", []byte("world"), int64(3)})
	want := []string{"hello", "world", "3"}
	if len(got) != len(want) {
		t.Fatalf("redisTextValues() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("redisTextValues() = %v, want %v", got, want)
		}
	}
}

func TestQueryRedisRejectsEmptyCommand(t *testing.T) {
	if _, err := Query("redis://localhost:6379/0", "   "); err == nil {
		t.Fatal("Query returned nil error for an empty Redis command")
	}
}

func TestIsRedisDsn(t *testing.T) {
	tests := []struct {
		dsn  string
		want bool
	}{
		{dsn: "redis://localhost:6379/0", want: true},
		{dsn: "rediss://localhost:6379/0", want: true},
		{dsn: "redis-cluster://localhost:6379", want: false},
		{dsn: "Redis://localhost:6379", want: false},
		{dsn: "", want: false},
	}

	for _, test := range tests {
		if got := IsRedisDSN(test.dsn); got != test.want {
			t.Errorf("isRedisDsn(%q) = %t, want %t", test.dsn, got, test.want)
		}
	}
}
