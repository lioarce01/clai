package tui

import (
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
	mv.renderer.SetWidth(w - 6)
}

func (mv MessageView) Render(msg llm.Message, isStreaming bool) string {
	t := mv.styles.theme
	contentWidth := mv.width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	switch msg.Role {
	case llm.RoleUser:
		rendered, err := mv.renderer.Render(msg.Content)
		if err != nil {
			rendered = msg.Content
		}
		body := strings.TrimSpace(rendered)
		return mv.styles.UserBlock.Width(contentWidth).Render(body)

	case llm.RoleAssistant:
		var parts []string

		// Thinking block (reasoning models)
		if msg.Reasoning != "" {
			parts = append(parts, mv.renderThinking(msg.Reasoning, isStreaming))
		}

		// Main content
		if msg.Content != "" {
			rendered, err := mv.renderer.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}
			dot := lipgloss.NewStyle().Foreground(t.AIAccent).Render("●  ")
			parts = append(parts, dot+strings.TrimSpace(rendered))
		} else if isStreaming {
			dot := lipgloss.NewStyle().Foreground(t.Warning).Render("●")
			if msg.Reasoning == "" {
				// Nothing yet — pulsing dot
				parts = append(parts, dot)
			}
			// If reasoning is streaming, dot is already shown in thinking block
		}

		return mv.styles.AIBlock.Width(contentWidth).Render(strings.Join(parts, "\n"))

	case llm.RoleSystem:
		body := lipgloss.NewStyle().Foreground(t.TextSubtle).Italic(true).Render(msg.Content)
		return mv.styles.AIBlock.Width(contentWidth).Render(body)

	default:
		return msg.Content
	}
}

// renderThinking renders the model's reasoning as a subtle dim block.
func (mv MessageView) renderThinking(reasoning string, isStreaming bool) string {
	t := mv.styles.theme

	label := lipgloss.NewStyle().Foreground(t.TextSubtle).Italic(true).Render("thinking")
	if isStreaming {
		label += " " + lipgloss.NewStyle().Foreground(t.Warning).Render("●")
	}

	body := lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Italic(true).
		Render(reasoning)

	block := lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.BorderFaint).
		PaddingLeft(2).
		Render(label + "\n" + body)

	return block
}

// RenderWelcome returns the welcome screen shown before any conversation.
func RenderWelcome(styles Styles, width int) string {
	t := styles.theme
	_ = width

	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("✦ clai")
	sub := lipgloss.NewStyle().Foreground(t.TextMuted).Render("terminal chat for LLMs")
	tip := lipgloss.NewStyle().Foreground(t.TextSubtle).Render("ctrl+o  configure api key & model")

	return "\n  " + title + "  " + sub + "\n  " + tip + "\n"
}
