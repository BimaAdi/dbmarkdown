package db

import (
	"strings"

	"github.com/BimaAdi/dbmarkdown/format"
)

type mockDB struct{}

var _ DBInterface = mockDB{}

func (mockDB) IsMyDSN(dsn string) bool {
	return strings.HasPrefix(dsn, "mock://")
}

func (mockDB) Query(_, _ string) (string, error) {
	header := []string{"id", "name", "is_done"}
	data := [][]string{{"1", "first todo", "1"}}
	return format.MarkdownTable(header, data), nil
}

// init function is a special, built-in function in go
// that automatically executes before the main() function
// or when a package is imported
//
// in dbmarkdown it used for build tags
// so you can create minimalis cli with only db driver
// that you actually need
//
// let's say you just need postgres and redis
// you just need build with tags
// `go build -tags="postgres redis"`
func init() {
	RegisterDatabase(mockDB{})
}
