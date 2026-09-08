package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/BimaAdi/dbmarkdown/format"
	"github.com/go-sql-driver/mysql"
)

func isMySQLDSN(dsn string) bool {
	return strings.HasPrefix(dsn, "mysql://")
}

func queryMySQL(dsn, query string) (string, error) {
	nativeDSN, err := mysqlNativeDSN(dsn)
	if err != nil {
		return "", fmt.Errorf("open database: %w", err)
	}

	database, err := sql.Open("mysql", nativeDSN)
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

	return markdownTableMySQL(rows, columns)
}

func markdownTableMySQL(rows *sql.Rows, columns []string) (string, error) {
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

func mysqlNativeDSN(dsn string) (string, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("invalid MySQL DSN: %w", err)
	}
	if parsed.Scheme != "mysql" || parsed.Host == "" {
		return "", fmt.Errorf("invalid MySQL DSN: expected mysql://user:password@host/database")
	}

	config := mysql.Config{
		Net:    "tcp",
		Addr:   parsed.Host,
		DBName: strings.TrimPrefix(parsed.Path, "/"),
	}
	if parsed.User != nil {
		config.User = parsed.User.Username()
		config.Passwd, _ = parsed.User.Password()
	}
	if config.Params, err = queryParams(parsed.Query()); err != nil {
		return "", fmt.Errorf("invalid MySQL DSN parameters: %w", err)
	}

	return config.FormatDSN(), nil
}

func queryParams(values url.Values) (map[string]string, error) {
	params := make(map[string]string, len(values))
	for key, value := range values {
		if len(value) != 1 {
			return nil, fmt.Errorf("parameter %q must be specified once", key)
		}
		params[key] = value[0]
	}
	return params, nil
}
