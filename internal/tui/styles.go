package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all pre-built Lip Gloss styles.
type Styles struct {
	theme Theme

	// Chrome
	Header    lipgloss.Style
	StatusBar lipgloss.Style
	Divider   lipgloss.Style

	// Messages
	UserBlock lipgloss.Style // grey background, no border
	AIBlock   lipgloss.Style // no background, no border

	// Input area — top + bottom border only
	InputFocused   lipgloss.Style
	InputUnfocused lipgloss.Style

	// Overlay
	OverlayBox lipgloss.Style

	// Text helpers
	TextMuted   lipgloss.Style
	TextSubtle  lipgloss.Style
	TextError   lipgloss.Style
	TextSuccess lipgloss.Style
	TextWarning lipgloss.Style
	TextPrimary lipgloss.Style
}

func NewStyles(t Theme) Styles {
	s := Styles{theme: t}

	s.Header = lipgloss.NewStyle().
		Foreground(t.TextMuted).
		Padding(0, 2)

	s.StatusBar = lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Padding(0, 2)

	s.Divider = lipgloss.NewStyle().
		Foreground(t.BorderFaint)

	// User message: subtle grey background highlight, padded, no border
	s.UserBlock = lipgloss.NewStyle().
		Background(t.Surface).
		Padding(1, 2).
		MarginTop(1)

	// Assistant message: plain, just spacing
	s.AIBlock = lipgloss.NewStyle().
		Padding(0, 2).
		MarginTop(1)

	// Input: top + bottom border, grey, no side borders
	inputBorder := lipgloss.Border{
		Top:    "─",
		Bottom: "─",
	}
	s.InputFocused = lipgloss.NewStyle().
		BorderTop(true).
		BorderBottom(true).
		BorderStyle(inputBorder).
		BorderForeground(t.Border).
		Padding(1, 2)

	s.InputUnfocused = lipgloss.NewStyle().
		BorderTop(true).
		BorderBottom(true).
		BorderStyle(inputBorder).
		BorderForeground(t.BorderFaint).
		Padding(1, 2)

	s.OverlayBox = lipgloss.NewStyle().
		Background(t.Surface).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1, 2)

	s.TextMuted = lipgloss.NewStyle().Foreground(t.TextMuted)
	s.TextSubtle = lipgloss.NewStyle().Foreground(t.TextSubtle)
	s.TextError = lipgloss.NewStyle().Foreground(t.Error).Bold(true)
	s.TextSuccess = lipgloss.NewStyle().Foreground(t.Success)
	s.TextWarning = lipgloss.NewStyle().Foreground(t.Warning)
	s.TextPrimary = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)

	return s
}
