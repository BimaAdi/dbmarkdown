package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/BimaAdi/dbmarkdown/format"
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

	return markdownTableSQLite(rows, columns)
}

func markdownTableSQLite(rows *sql.Rows, columns []string) (string, error) {
	data := make([][]string, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return "", fmt.Errorf("read query result: %w", err)
		}

		cells := make([]string, len(values))
		for i, value := range values {
			cells[i] = valueToString(value)
		}
		data = append(data, cells)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("read query result: %w", err)
	}
	return format.MarkdownTable(columns, data), nil
}
