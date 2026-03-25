package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Header renders the minimal top bar.
type Header struct {
	styles    Styles
	width     int
	model     string
	connected bool
	streaming bool
	version   string
}

func NewHeader(styles Styles, version string) Header {
	return Header{styles: styles, version: version}
}

func (h *Header) SetWidth(w int)      { h.width = w }
func (h *Header) SetModel(m string)   { h.model = m }
func (h *Header) SetConnected(c bool) { h.connected = c }
func (h *Header) SetStreaming(s bool) { h.streaming = s }

func (h Header) View() string {
	t := h.styles.theme

	left := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("✦ clai")

	var status string
	if h.streaming {
		status = lipgloss.NewStyle().Foreground(t.Warning).Render("● generating")
	} else if h.connected {
		status = lipgloss.NewStyle().Foreground(t.Success).Render("●")
	} else {
		status = lipgloss.NewStyle().Foreground(t.Error).Render("○")
	}

	model := lipgloss.NewStyle().Foreground(t.TextMuted).Render(h.model)
	right := fmt.Sprintf("%s  %s", model, status)

	gap := h.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if gap < 1 {
		gap = 1
	}

	line := h.styles.Header.Width(h.width).Render(
		left + strings.Repeat(" ", gap) + right,
	)

	// Faint divider below header
	div := h.styles.Divider.Render(strings.Repeat("─", h.width))
	return line + "\n" + div
}
