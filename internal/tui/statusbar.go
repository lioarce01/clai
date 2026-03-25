package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders the bottom application bar.
type StatusBar struct {
	styles       Styles
	width        int
	sessionName  string
	promptToks   int
	completeToks int
	keyHints     string
}

func NewStatusBar(styles Styles) StatusBar {
	return StatusBar{
		styles:   styles,
		keyHints: "^N new  ^S sessions  ^O settings  ^C quit",
	}
}

func (s *StatusBar) SetWidth(w int)          { s.width = w }
func (s *StatusBar) SetSessionName(n string) { s.sessionName = n }
func (s *StatusBar) SetTokens(prompt, complete int) {
	s.promptToks = prompt
	s.completeToks = complete
}

func (s StatusBar) View() string {
	t := s.styles.theme

	sessionPart := lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Render(fmt.Sprintf("  %s", s.sessionName))

	var tokenPart string
	if s.promptToks > 0 || s.completeToks > 0 {
		tokenPart = lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Render(fmt.Sprintf("  ↑%d ↓%d tok", s.promptToks, s.completeToks))
	}

	hints := lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Render(s.keyHints + "  ")

	// Left section
	left := sessionPart + tokenPart
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(hints)
	gap := s.width - leftWidth - rightWidth
	if gap < 0 {
		gap = 0
	}

	spaces := fmt.Sprintf("%*s", gap, "")
	return s.styles.StatusBar.Width(s.width).Render(left + spaces + hints)
}
