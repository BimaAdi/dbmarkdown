package db

import (
	"fmt"
	"strings"
)

// Query executes a database query and formats its result as Markdown.
func Query(dsn, query string) (string, error) {
	if isSQLiteDSN(dsn) {
		return querySQLite(strings.TrimPrefix(dsn, "sqlite://"), query)
	}
	if isPostgresDsn(dsn) {
	  return queryPostgres(dsn, query)
	}
	return "", fmt.Errorf("unsupported database DSN: %q", dsn)
}
