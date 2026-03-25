package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all pre-built Lip Gloss styles.
type Styles struct {
	theme Theme

	// Chrome
	Header    lipgloss.Style
	StatusBar lipgloss.Style
	Divider   lipgloss.Style

	// Role labels
	UserLabel lipgloss.Style
	AILabel   lipgloss.Style

	// Message layout — left-border accent, no box
	UserBlock lipgloss.Style
	AIBlock   lipgloss.Style

	// Thinking block
	ThinkingLabel lipgloss.Style
	ThinkingBlock lipgloss.Style

	// Input
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

	// Role labels — bold, colored
	s.UserLabel = lipgloss.NewStyle().
		Foreground(t.UserAccent).
		Bold(true)

	s.AILabel = lipgloss.NewStyle().
		Foreground(t.AIAccent).
		Bold(true)

	// Message blocks — left border accent, no top/bottom/right border
	s.UserBlock = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(t.UserAccent).
		PaddingLeft(2).
		MarginTop(1)

	s.AIBlock = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(t.AIAccent).
		PaddingLeft(2).
		MarginTop(1)

	// Thinking block — faint left border, dimmed italic text
	s.ThinkingLabel = lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Italic(true)

	s.ThinkingBlock = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.BorderFaint).
		PaddingLeft(2).
		MarginTop(1)

	// Input area
	s.InputFocused = lipgloss.NewStyle().
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.Primary).
		Padding(0, 2)

	s.InputUnfocused = lipgloss.NewStyle().
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.BorderFaint).
		Padding(0, 2)

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
