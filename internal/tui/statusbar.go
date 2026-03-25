package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders the minimal bottom bar.
type StatusBar struct {
	styles       Styles
	width        int
	sessionName  string
	promptToks   int
	completeToks int
}

func NewStatusBar(styles Styles) StatusBar {
	return StatusBar{styles: styles}
}

func (s *StatusBar) SetWidth(w int)          { s.width = w }
func (s *StatusBar) SetSessionName(n string) { s.sessionName = n }
func (s *StatusBar) SetTokens(prompt, complete int) {
	s.promptToks = prompt
	s.completeToks = complete
}

func (s StatusBar) View() string {
	t := s.styles.theme

	div := s.styles.Divider.Render(strings.Repeat("─", s.width))

	session := lipgloss.NewStyle().Foreground(t.TextMuted).Render(s.sessionName)

	var toks string
	if s.promptToks > 0 || s.completeToks > 0 {
		toks = lipgloss.NewStyle().Foreground(t.TextSubtle).
			Render(fmt.Sprintf("  %d↑ %d↓", s.promptToks, s.completeToks))
	}

	hints := lipgloss.NewStyle().Foreground(t.TextSubtle).
		Render("^N  ^S  ^O  ^C")

	left := session + toks
	gap := s.width - lipgloss.Width(left) - lipgloss.Width(hints) - 4
	if gap < 1 {
		gap = 1
	}

	bar := s.styles.StatusBar.Width(s.width).Render(
		left + strings.Repeat(" ", gap) + hints,
	)

	return div + "\n" + bar
}
