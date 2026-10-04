//go:build postgres

package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/BimaAdi/dbmarkdown/format"
	_ "github.com/lib/pq"
)

type postgresDB struct{}

var _ DBInterface = postgresDB{}

func (postgresDB) IsMyDSN(dsn string) bool {
	return strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://")
}

func (postgresDB) Query(dsn, query string) (string, error) {
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

	return markdownTablePostgres(rows, columns)
}

func isPostgresDsn(dsn string) bool {
	return (postgresDB{}).IsMyDSN(dsn)
}

func markdownTablePostgres(rows *sql.Rows, columns []string) (string, error) {
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

func init() {
	RegisterDatabase(postgresDB{})
}
