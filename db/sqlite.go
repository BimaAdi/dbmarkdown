package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

func isSQLiteDSN(dsn string) bool {
	return strings.HasPrefix(dsn, "sqlite://")
}

func querySQLite(path, query string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("open database: SQLite path is empty")
	}

	database, err := sql.Open("sqlite", path)
	if err != nil {
		return "", fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	rows, err := database.Query(query)
	if err != nil {
		return "", fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("read query result: %w", err)
	}

	return markdownTable(rows, columns)
}
