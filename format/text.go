package format

import "strings"

// Text renders values as one plain-text value per line.
func Text(values []string) string {
	return "```\n" + strings.Join(values, "\n") + "\n```\n"
}
