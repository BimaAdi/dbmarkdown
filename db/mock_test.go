package db

import "testing"

func TestMockDBIsMyDSN(t *testing.T) {
	tests := []struct {
		dsn  string
		want bool
	}{
		{dsn: "mock://", want: true},
		{dsn: "mock://test", want: true},
		{dsn: "mocked://test", want: false},
		{dsn: "sqlite://test.db", want: false},
		{dsn: "", want: false},
	}

	for _, test := range tests {
		if got := (mockDB{}).IsMyDSN(test.dsn); got != test.want {
			t.Errorf("IsMyDSN(%q) = %t, want %t", test.dsn, got, test.want)
		}
	}
}

func TestMockDBQueryReturnsMarkdownTable(t *testing.T) {
	got, err := Query("mock://", "SELECT * FROM todo")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	want := "| id | name       | is_done |\n" +
		"|----|------------|---------|\n" +
		"| 1  | first todo | 1       |"
	if got != want {
		t.Errorf("Query() =\n%s\nwant\n%s", got, want)
	}
}
