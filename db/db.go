package db

import (
	"fmt"
	"strings"
)

// Query executes a database query and formats its result as Markdown.
func Query(dsn, query string) (string, error) {
	if IsRedisDSN(dsn) {
		return queryRedis(dsn, query)
	}
	if isSQLiteDSN(dsn) {
		return querySQLite(strings.TrimPrefix(dsn, "sqlite://"), query)
	}
	if isPostgresDsn(dsn) {
		return queryPostgres(dsn, query)
	}
	if isMySQLDSN(dsn) {
		return queryMySQL(dsn, query)
	}
	return "", fmt.Errorf("unsupported database DSN: %q", dsn)
}

func valueToString(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}
