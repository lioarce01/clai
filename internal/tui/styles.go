package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all pre-built Lip Gloss styles derived from the theme.
type Styles struct {
	theme Theme

	// App chrome
	Header    lipgloss.Style
	StatusBar lipgloss.Style

	// Chat area
	ChatViewport lipgloss.Style

	// Message bubbles
	UserBubble      lipgloss.Style
	AssistantBubble lipgloss.Style
	SystemBubble    lipgloss.Style

	// Role badges
	UserBadge      lipgloss.Style
	AssistantBadge lipgloss.Style
	SystemBadge    lipgloss.Style

	// Input
	InputBorderFocused   lipgloss.Style
	InputBorderUnfocused lipgloss.Style

	// Overlay
	OverlayBox lipgloss.Style

	// Text variants
	TextMuted   lipgloss.Style
	TextSubtle  lipgloss.Style
	TextError   lipgloss.Style
	TextSuccess lipgloss.Style
	TextWarning lipgloss.Style
	TextPrimary lipgloss.Style
}

// NewStyles builds all styles from the given theme.
func NewStyles(t Theme) Styles {
	s := Styles{theme: t}

	s.Header = lipgloss.NewStyle().
		Background(t.Surface).
		Foreground(t.Text).
		Padding(0, 1).
		Bold(true)

	s.StatusBar = lipgloss.NewStyle().
		Background(t.SurfaceAlt).
		Foreground(t.TextMuted).
		Padding(0, 1)

	s.ChatViewport = lipgloss.NewStyle().
		Padding(0, 1)

	s.UserBubble = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.UserBadge).
		Padding(0, 1).
		MarginTop(1)

	s.AssistantBubble = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.AssistantBadge).
		Padding(0, 1).
		MarginTop(1)

	s.SystemBubble = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.SystemBadge).
		Padding(0, 1).
		MarginTop(1)

	s.UserBadge = lipgloss.NewStyle().
		Foreground(t.UserBadge).
		Bold(true)

	s.AssistantBadge = lipgloss.NewStyle().
		Foreground(t.AssistantBadge).
		Bold(true)

	s.SystemBadge = lipgloss.NewStyle().
		Foreground(t.SystemBadge).
		Bold(true).
		Italic(true)

	s.InputBorderFocused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocused).
		Padding(0, 1)

	s.InputBorderUnfocused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderUnfocused).
		Padding(0, 1)

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
