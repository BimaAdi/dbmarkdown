package db

import (
	"database/sql"
	"fmt"

	"github.com/BimaAdi/dbmarkdown/format"
	_ "github.com/lib/pq"
)

// Query executes a PostgreSQL query and formats its result as Markdown.
func Query(dsn, query string) (string, error) {
	database, err := sql.Open("postgres", dsn)
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
