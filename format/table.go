package format

import (
	"fmt"
	"strings"
)

// MarkdownTable renders a header and rows as a Markdown table.
func MarkdownTable(header []string, data [][]string) string {
	widths := make([]int, len(header))
	for i, column := range header {
		widths[i] = len(formatCell(column))
	}
	for _, row := range data {
		for i, cell := range row {
			if i >= len(widths) {
				break
			}
			cell = formatCell(cell)
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var builder strings.Builder
	writeRow := func(cells []string) {
		padded := make([]string, len(widths))
		for i := range widths {
			cell := ""
			if i < len(cells) {
				cell = formatCell(cells[i])
			}
			padded[i] = cell + strings.Repeat(" ", widths[i]-len(cell))
		}
		fmt.Fprintf(&builder, "| %s |\n", strings.Join(padded, " | "))
	}
	writeRow(header)
	separators := make([]string, len(header))
	for i, width := range widths {
		separators[i] = strings.Repeat("-", max(width+2, 3))
	}
	fmt.Fprintf(&builder, "|%s|\n", strings.Join(separators, "|"))
	for _, row := range data {
		writeRow(row)
	}
	return strings.TrimSuffix(builder.String(), "\n")
}

func formatCell(value any) string {
	if value == nil {
		return ""
	}
	var text string
	switch v := value.(type) {
	case []byte:
		text = string(v)
	default:
		text = fmt.Sprint(v)
	}
	return strings.ReplaceAll(strings.ReplaceAll(text, "|", "\\|"), "\n", "<br>")
}
