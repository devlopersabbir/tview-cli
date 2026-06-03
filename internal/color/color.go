// Package color provides ANSI terminal color constants for tview-cli rendering.
package color

// ANSI escape sequences for terminal styling.
const (
	Reset    = "\033[0m"
	Bold     = "\033[1m"
	Green    = "\033[38;5;86m"
	Red      = "\033[38;5;203m"
	Gray     = "\033[38;5;244m"
	DarkGray = "\033[38;5;238m"
	White    = "\033[97m"
	Yellow   = "\033[38;5;220m"
)

// For returns the appropriate color for a bullish or bearish candle.
func For(bull bool) string {
	if bull {
		return Green
	}
	return Red
}
