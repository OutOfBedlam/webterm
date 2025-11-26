package webtail

import (
	"regexp"
	"strings"

	"github.com/OutOfBedlam/webterm"
)

type Plugin interface {
	// Apply processes a line and returns the modified line
	// and a boolean indicating processing ahead
	// if further plugins should continue processing
	// or drop the line.
	Apply(line string) (string, bool)
}

func NewWithSyntaxHighlighting(syntax ...string) Plugin {
	return syntaxColoring(syntax)
}

type syntaxColoring []string

var slogKeyValuePattern = regexp.MustCompile(`(\w+)=("(?:[^"\\]|\\.)*"|[^\s]+)`)

func (c syntaxColoring) Apply(line string) (string, bool) {
	// Default Keywords
	for _, syntax := range c {
		switch strings.ToLower(syntax) {
		case "level", "levels":
			line = strings.ReplaceAll(line, "TRACE", webterm.ColorDarkGray+"TRACE"+webterm.ColorReset)
			line = strings.ReplaceAll(line, "DEBUG", webterm.ColorLightGray+"DEBUG"+webterm.ColorReset)
			line = strings.ReplaceAll(line, "INFO", webterm.ColorGreen+"INFO"+webterm.ColorReset)
			line = strings.ReplaceAll(line, "WARN", webterm.ColorYellow+"WARN"+webterm.ColorReset)
			line = strings.ReplaceAll(line, "ERROR", webterm.ColorRed+"ERROR"+webterm.ColorReset)
		case "slog-text":
			// Color name=value patterns in slog format
			line = slogKeyValuePattern.ReplaceAllStringFunc(line, func(match string) string {
				parts := strings.SplitN(match, "=", 2)
				if len(parts) == 2 {
					key := parts[0]
					value := parts[1]
					return webterm.ColorCyan + key + webterm.ColorReset + "=" + webterm.ColorBlue + value + webterm.ColorReset
				}
				return match
			})
		case "slog-json":
			// Color JSON key:value patterns in slog format
			line = regexp.MustCompile(`"(\w+)":\s*("(?:[^"\\]|\\.)*"|[^\s,}]+)`).ReplaceAllStringFunc(line, func(match string) string {
				parts := strings.SplitN(match, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					return webterm.ColorCyan + key + webterm.ColorReset + ":" + webterm.ColorBlue + value + webterm.ColorReset
				}
				return match
			})
		case "syslog":
			// /var/log/syslog specific coloring
			// Pattern: timestamp hostname process[pid]: message
			syslogPattern := regexp.MustCompile(`^(\S+)\s+(\S+)\s+([^\s:]+(?:\[\d+\])?):(.*)$`)
			line = syslogPattern.ReplaceAllStringFunc(line, func(match string) string {
				matches := syslogPattern.FindStringSubmatch(match)
				if len(matches) == 5 {
					timestamp := webterm.ColorBlue + matches[1] + webterm.ColorReset
					hostname := webterm.ColorCyan + matches[2] + webterm.ColorReset
					process := webterm.ColorYellow + matches[3] + webterm.ColorReset
					message := matches[4]
					return timestamp + " " + hostname + " " + process + ":" + message
				}
				return match
			})
			// syslog file encodes ESC as #033[
			line = strings.ReplaceAll(line, "#033[", "\033[")
		}
	}
	return line, true
}
