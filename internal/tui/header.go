package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Header renders the top application bar.
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

	title := h.styles.TextPrimary.Render(fmt.Sprintf("✦ CLAI %s", h.version))

	var connStatus string
	if h.streaming {
		connStatus = lipgloss.NewStyle().Foreground(t.Warning).Render("● streaming")
	} else if h.connected {
		connStatus = lipgloss.NewStyle().Foreground(t.Success).Render("● connected")
	} else {
		connStatus = lipgloss.NewStyle().Foreground(t.Error).Render("○ disconnected")
	}

	modelStr := h.styles.TextMuted.Render(h.model)

	right := fmt.Sprintf("%s  %s", modelStr, connStatus)
	leftWidth := lipgloss.Width(title)
	rightWidth := lipgloss.Width(right)
	gap := h.width - leftWidth - rightWidth - 2 // account for padding
	if gap < 1 {
		gap = 1
	}
	spaces := lipgloss.NewStyle().Render(fmt.Sprintf("%*s", gap, ""))

	return h.styles.Header.Width(h.width).Render(title + spaces + right)
}
