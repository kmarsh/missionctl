package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// output chooses between tables for people and JSON for scripts.
type output struct {
	json  bool
	color bool
}

func (a *app) output() output {
	json := a.json || !isTerminal(os.Stdout)
	return output{json: json, color: !json && os.Getenv("NO_COLOR") == ""}
}

// isTerminal is true for an interactive terminal; /dev/null and pipes aren't.
func isTerminal(file *os.File) bool {
	return term.IsTerminal(int(file.Fd()))
}

func (o output) printJSON(value any) {
	writeJSON(os.Stdout, value)
}

func writeJSON(w io.Writer, value any) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
}

var colorEscape = regexp.MustCompile("\x1b\\[[0-9;]*m")

// visibleWidth counts characters as they appear on screen, ignoring color escapes.
func visibleWidth(text string) int {
	return utf8.RuneCountInString(colorEscape.ReplaceAllString(text, ""))
}

// table prints left-aligned columns separated by two spaces, leaving out columns
// that are empty on every row. The last column isn't padded.
func (o output) table(rows [][]string) {
	columns := 0
	for _, row := range rows {
		columns = max(columns, len(row))
	}
	var kept []int
	for column := 0; column < columns; column++ {
		for _, row := range rows {
			if column < len(row) && row[column] != "" {
				kept = append(kept, column)
				break
			}
		}
	}
	widths := make([]int, len(kept))
	for _, row := range rows {
		for i, column := range kept {
			if column < len(row) {
				widths[i] = max(widths[i], visibleWidth(row[column]))
			}
		}
	}
	for _, row := range rows {
		var cells []string
		for i, column := range kept {
			cell := ""
			if column < len(row) {
				cell = row[column]
			}
			if i < len(kept)-1 {
				cell += strings.Repeat(" ", widths[i]-visibleWidth(cell))
			}
			cells = append(cells, cell)
		}
		fmt.Println(strings.TrimRight(strings.Join(cells, "  "), " "))
	}
}

// dot is a bullet in a project's #RRGGBB color, when the terminal can show it.
func (o output) dot(hex string) string {
	value, err := strconv.ParseUint(strings.TrimPrefix(hex, "#"), 16, 32)
	if !o.color || len(hex) != 7 || !strings.HasPrefix(hex, "#") || err != nil {
		return "●"
	}
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm●\x1b[0m", value>>16&0xFF, value>>8&0xFF, value&0xFF)
}

// summary is the first line of text, cut to limit characters.
func summary(text string, limit int) string {
	line, _, _ := strings.Cut(text, "\n")
	if utf8.RuneCountInString(line) <= limit {
		return line
	}
	return string([]rune(line)[:limit-1]) + "…"
}

// Accessors for the API's JSON, which is kept as maps so --json can print it as-is.

func object(value any) map[string]any {
	m, _ := value.(map[string]any)
	return m
}

func objects(value any) []map[string]any {
	items, _ := value.([]any)
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, object(item))
	}
	return result
}

func str(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

func integer(m map[string]any, key string) int {
	n, _ := m[key].(json.Number)
	i, _ := n.Int64()
	return int(i)
}

// flag reads a boolean field, treating a missing one as fallback.
func flag(m map[string]any, key string, fallback bool) bool {
	if b, ok := m[key].(bool); ok {
		return b
	}
	return fallback
}
