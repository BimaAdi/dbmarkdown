package db

import "fmt"

type DBInterface interface {
	IsMyDSN(dsn string) bool
	Query(dsn, query string) (string, error)
}

var databases = []DBInterface{}

func RegisterDatabase(dbi DBInterface) {
	databases = append(databases, dbi)
}

// Query executes a database query and formats its result as Markdown.
func Query(dsn, query string) (string, error) {
	for _, database := range databases {
		if database.IsMyDSN(dsn) {
			return database.Query(dsn, query)
		}
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
