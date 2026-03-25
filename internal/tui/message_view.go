package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lioarce01/clai/internal/llm"
	"github.com/lioarce01/clai/internal/markdown"
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

		var parts []string

		// Render reasoning block if present
		if msg.Reasoning != "" {
			parts = append(parts, mv.renderReasoning(msg.Reasoning, isStreaming, innerWidth))
		}

		// Render main content
		rendered, err := mv.renderer.Render(msg.Content)
		if err != nil {
			rendered = msg.Content
		}
		content = strings.TrimRight(rendered, "\n")
		if isStreaming && content == "" && msg.Reasoning == "" {
			content = lipgloss.NewStyle().
				Foreground(mv.styles.theme.Warning).
				Render("▋")
		}
		if content != "" {
			parts = append(parts, content)
		}

		bubble := mv.styles.AssistantBubble.Width(innerWidth)
		ts := mv.styles.TextSubtle.Render(msg.CreatedAt.Format("15:04"))
		header := fmt.Sprintf("%s  %s", badge, ts)
		body := strings.Join(parts, "\n")
		return bubble.Render(header + "\n" + body)

	case llm.RoleSystem:
		badge = mv.styles.SystemBadge.Render("  System")
		content = mv.styles.TextMuted.Render(msg.Content)
		bubble := mv.styles.SystemBubble.Width(innerWidth)
		return bubble.Render(badge + "\n" + content)

	default:
		return msg.Content
	}
}

// renderReasoning renders the model's internal reasoning in a distinct "thinking" block.
func (mv MessageView) renderReasoning(reasoning string, isStreaming bool, innerWidth int) string {
	t := mv.styles.theme

	label := lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Italic(true).
		Render("💭 Thinking")

	if isStreaming {
		label += lipgloss.NewStyle().
			Foreground(t.Warning).
			Render(" ●")
	}

	// Wrap reasoning text, dim and italic
	reasoningStyle := lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Italic(true)

	// Word-wrap manually to innerWidth - 4 (account for block padding)
	wrapWidth := innerWidth - 4
	if wrapWidth < 20 {
		wrapWidth = 20
	}
	wrapped := wordWrap(reasoning, wrapWidth)
	body := reasoningStyle.Render(wrapped)

	block := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(t.TextSubtle).
		PaddingLeft(1).
		Width(innerWidth - 2).
		Render(label + "\n" + body)

	return block
}

// wordWrap wraps text at the given column width, preserving existing newlines.
func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	var result strings.Builder
	for _, line := range strings.Split(text, "\n") {
		if len(line) <= width {
			result.WriteString(line)
			result.WriteByte('\n')
			continue
		}
		words := strings.Fields(line)
		col := 0
		for i, w := range words {
			wl := len(w)
			if col > 0 && col+1+wl > width {
				result.WriteByte('\n')
				col = 0
			}
			if col > 0 {
				result.WriteByte(' ')
				col++
			}
			result.WriteString(w)
			col += wl
			_ = i
		}
		result.WriteByte('\n')
	}
	return strings.TrimRight(result.String(), "\n")
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
