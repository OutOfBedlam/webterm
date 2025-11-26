package webterm

import (
	"fmt"
	"regexp"
)

// Remove any ANSI color codes from label, with regexp
var stripAnsiCodesRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func StripAnsiCodes(s string) string {
	return stripAnsiCodesRegexp.ReplaceAllString(s, "")
}

func Colorize(s string, color string) string {
	return fmt.Sprintf("%s%s%s", color, s, ColorReset)
}

// ANSI color codes
const (
	ColorReset         = "\033[0m"
	ColorBlack         = "\033[30m"       // Black
	ColorRed           = "\033[31m"       // Red
	ColorGreen         = "\033[32m"       // Green
	ColorYellow        = "\033[33m"       // Yellow
	ColorBlue          = "\033[34m"       // Blue
	ColorMagenta       = "\033[35m"       // Magenta
	ColorCyan          = "\033[36m"       // Cyan for keys
	ColorLightGray     = "\033[37m"       // Light gray
	ColorNavy          = "\033[38;5;17m"  // Navy
	ColorTeal          = "\033[38;5;51m"  // Teal
	ColorMaroon        = "\033[38;5;52m"  // Maroon
	ColorIndigo        = "\033[38;5;57m"  // Indigo
	ColorLightBlue     = "\033[38;5;81m"  // Light Blue
	ColorBrown         = "\033[38;5;94m"  // Brown
	ColorOlive         = "\033[38;5;100m" // Olive
	ColorLightGreen    = "\033[38;5;120m" // Light Green
	ColorPurple        = "\033[38;5;135m" // Purple
	ColorLime          = "\033[38;5;154m" // Lime
	ColorPink          = "\033[38;5;205m" // Pink
	ColorOrange        = "\033[38;5;208m" // Orange
	ColorGray          = "\033[38;5;245m" // Gray
	ColorDarkGray      = "\033[90m"       // Dark gray
	ColorBrightRed     = "\033[91m"       // Bright Red
	ColorBrightGreen   = "\033[92m"       // Bright Green
	ColorBrightYellow  = "\033[93m"       // Bright Yellow
	ColorBrightBlue    = "\033[94m"       // Bright Blue
	ColorBrightMagenta = "\033[95m"       // Bright Magenta
	ColorBrightCyan    = "\033[96m"       // Bright Cyan
	ColorWhite         = "\033[97m"       // White
)
