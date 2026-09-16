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
