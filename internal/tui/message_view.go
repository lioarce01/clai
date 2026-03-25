package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/clai/internal/llm"
	"github.com/yourusername/clai/internal/markdown"
)

// MessageView renders a single chat message.
type MessageView struct {
	styles   Styles
	renderer *markdown.Renderer
	width    int
}

func NewMessageView(styles Styles, renderer *markdown.Renderer, width int) MessageView {
	return MessageView{styles: styles, renderer: renderer, width: width}
}

func (mv *MessageView) SetWidth(w int) {
	mv.width = w
	mv.renderer.SetWidth(w - 6) // account for border + padding
}

// Render converts a message to a fully-styled terminal string.
func (mv MessageView) Render(msg llm.Message, isStreaming bool) string {
	innerWidth := mv.width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}

	var badge, content string

	switch msg.Role {
	case llm.RoleUser:
		badge = mv.styles.UserBadge.Render("  You")
		rendered, err := mv.renderer.Render(msg.Content)
		if err != nil {
			rendered = msg.Content
		}
		content = strings.TrimRight(rendered, "\n")
		bubble := mv.styles.UserBubble.Width(innerWidth)
		ts := mv.styles.TextSubtle.Render(msg.CreatedAt.Format("15:04"))
		header := fmt.Sprintf("%s  %s", badge, ts)
		return bubble.Render(header + "\n" + content)

	case llm.RoleAssistant:
		badge = mv.styles.AssistantBadge.Render("  Assistant")
		if isStreaming {
			badge += lipgloss.NewStyle().
				Foreground(mv.styles.theme.Warning).
				Render(" ●")
		}
		rendered, err := mv.renderer.Render(msg.Content)
		if err != nil {
			rendered = msg.Content
		}
		content = strings.TrimRight(rendered, "\n")
		if isStreaming && content == "" {
			content = lipgloss.NewStyle().
				Foreground(mv.styles.theme.Warning).
				Render("▋")
		}
		bubble := mv.styles.AssistantBubble.Width(innerWidth)
		ts := mv.styles.TextSubtle.Render(msg.CreatedAt.Format("15:04"))
		header := fmt.Sprintf("%s  %s", badge, ts)
		return bubble.Render(header + "\n" + content)

	case llm.RoleSystem:
		badge = mv.styles.SystemBadge.Render("  System")
		content = mv.styles.TextMuted.Render(msg.Content)
		bubble := mv.styles.SystemBubble.Width(innerWidth)
		return bubble.Render(badge + "\n" + content)

	default:
		return msg.Content
	}
}

// RenderWelcome returns the initial welcome message shown before any conversation.
func RenderWelcome(styles Styles, width int) string {
	t := styles.theme
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1, 3).
		Width(width - 4)

	title := lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Render("✦ Welcome to CLAI")

	subtitle := lipgloss.NewStyle().
		Foreground(t.TextMuted).
		Render("A high-performance terminal LLM chat client")

	tips := lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Render("Start typing below  •  Ctrl+O to configure  •  Ctrl+H for help")

	return "\n" + box.Render(title+"\n"+subtitle+"\n\n"+tips) + "\n"
}
