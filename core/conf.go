package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// Config maps a configuration name to a database DSN.
type Config map[string]string

// defaultConfigPath is the configuration file used when neither the --conf
// flag nor a "conf:" section in the Markdown file provides a configuration.
const defaultConfigPath = "db.json"

// LoadConfig resolves the database configuration using the following priority,
// highest first:
//  1. path, when it is not empty (the user-provided --conf value). The config
//     is read from that file, and a missing file is reported as
//     "config <path> not found".
//  2. a "conf:" section in markdown, which must be followed by a ```json code
//     fence holding the configuration.
//  3. the default db.json file.
func LoadConfig(path, markdown string) (Config, error) {
	if path != "" {
		return readConfigFile(path)
	}
	cfg, found, err := configFromMarkdown(markdown)
	if err != nil {
		return nil, err
	}
	if found {
		return cfg, nil
	}
	return readConfigFile(defaultConfigPath)
}

// readConfigFile reads and parses a JSON configuration file.
func readConfigFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("config %s not found", path)
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// configFromMarkdown extracts the configuration from a "conf:" section of the
// Markdown file. The section is a "conf:" line followed by a ```json code
// fence. Lines inside any code fence are ignored, so a "conf:" example shown
// in a fenced block is not mistaken for a real one. found reports whether such
// a section exists.
func configFromMarkdown(markdown string) (Config, bool, error) {
	lines := strings.SplitAfter(markdown, "\n")
	inFence := false
	fenceLen := 0
	for i := range lines {
		trimmed := trimmedLine(lines[i])
		run := 0
		for run < len(trimmed) && trimmed[run] == '`' {
			run++
		}
		if inFence {
			// A fence only closes on a backtick run at least as long as
			// the one that opened it, with no info string.
			if run >= fenceLen && strings.TrimSpace(trimmed[run:]) == "" {
				inFence = false
			}
			continue
		}
		if run >= 3 {
			inFence = true
			fenceLen = run
			continue
		}
		if trimmed != "conf:" {
			continue
		}
		if i+1 >= len(lines) || trimmedLine(lines[i+1]) != "```json" {
			return nil, true, errors.New(`config section "conf:" must be followed by a ` + "```json" + ` code fence`)
		}
		var content strings.Builder
		for j := i + 2; ; j++ {
			if j >= len(lines) {
				return nil, true, errors.New(`config section "conf:" has an unclosed ` + "```json" + ` code fence`)
			}
			if trimmedLine(lines[j]) == "```" {
				break
			}
			content.WriteString(lines[j])
		}
		var cfg Config
		if err := json.Unmarshal([]byte(content.String()), &cfg); err != nil {
			return nil, true, fmt.Errorf("parse config: %w", err)
		}
		return cfg, true, nil
	}
	return nil, false, nil
}

func trimmedLine(line string) string {
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
}
