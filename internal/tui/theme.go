package tui

// Monochrome ANSI escape constants for styling.
// No color — only weight/style modifiers and grey 256-color shades.
const (
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Italic = "\033[3m"
	Reset  = "\033[0m"

	// 256-color grey shades (foreground)
	Grey300 = "\033[38;5;245m" // medium grey
	Grey500 = "\033[38;5;240m" // dark grey
	Grey700 = "\033[38;5;235m" // very dark grey
)
