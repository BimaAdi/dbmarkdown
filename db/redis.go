package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/BimaAdi/dbmarkdown/format"
	"github.com/redis/go-redis/v9"
)

// IsRedisDSN reports whether dsn uses a Redis connection scheme.
func IsRedisDSN(dsn string) bool {
	return strings.HasPrefix(dsn, "redis://") || strings.HasPrefix(dsn, "rediss://")
}

func queryRedis(dsn, query string) (string, error) {
	options, err := redis.ParseURL(dsn)
	if err != nil {
		return "", fmt.Errorf("open database: %w", err)
	}

	client := redis.NewClient(options)
	defer client.Close()

	parts := strings.Fields(query)
	if len(parts) == 0 {
		return "", fmt.Errorf("execute query: Redis command is empty")
	}
	args := make([]any, len(parts))
	for i, part := range parts {
		args[i] = part
	}

	result, err := client.Do(context.Background(), args...).Result()
	if err != nil {
		return "", fmt.Errorf("execute query: %w", err)
	}
	return format.Text(redisTextValues(result)), nil
}

func redisTextValues(value any) []string {
	switch value := value.(type) {
	case nil:
		return nil
	case []byte:
		return []string{string(value)}
	case string:
		return []string{value}
	case []string:
		return value
	case []any:
		values := make([]string, 0, len(value))
		for _, item := range value {
			values = append(values, redisTextValues(item)...)
		}
		return values
	default:
		return []string{fmt.Sprint(value)}
	}
}
