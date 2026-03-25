package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lioarce01/clai/internal/llm"
	"github.com/lioarce01/clai/internal/markdown"
)

// MessageView renders a single chat message in a flat, minimal style.
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
	mv.renderer.SetWidth(w - 4)
}

// Render converts a message to a styled terminal string.
func (mv MessageView) Render(msg llm.Message, isStreaming bool) string {
	t := mv.styles.theme
	contentWidth := mv.width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	switch msg.Role {
	case llm.RoleUser:
		label := mv.styles.UserLabel.Render("You")
		ts := lipgloss.NewStyle().Foreground(t.TextSubtle).Render(msg.CreatedAt.Format("15:04"))
		header := fmt.Sprintf("%s  %s", label, ts)

		rendered, err := mv.renderer.Render(msg.Content)
		if err != nil {
			rendered = msg.Content
		}
		body := strings.TrimSpace(rendered)

		return mv.styles.UserBlock.Width(contentWidth).Render(header + "\n" + body)

	case llm.RoleAssistant:
		labelText := "Assistant"
		if isStreaming {
			dot := lipgloss.NewStyle().Foreground(t.Warning).Render(" ●")
			labelText = mv.styles.AILabel.Render("Assistant") + dot
		} else {
			labelText = mv.styles.AILabel.Render("Assistant")
		}
		ts := lipgloss.NewStyle().Foreground(t.TextSubtle).Render(msg.CreatedAt.Format("15:04"))
		header := fmt.Sprintf("%s  %s", labelText, ts)

		var sections []string
		sections = append(sections, header)

		// Reasoning block
		if msg.Reasoning != "" {
			sections = append(sections, mv.renderThinking(msg.Reasoning, isStreaming))
		}

		// Main content
		if msg.Content != "" {
			rendered, err := mv.renderer.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}
			sections = append(sections, strings.TrimSpace(rendered))
		} else if isStreaming && msg.Reasoning == "" {
			// Show cursor only when nothing at all has arrived yet
			sections = append(sections, lipgloss.NewStyle().Foreground(t.Warning).Render("▋"))
		}

		body := strings.Join(sections, "\n")
		return mv.styles.AIBlock.Width(contentWidth).Render(body)

	case llm.RoleSystem:
		label := lipgloss.NewStyle().Foreground(t.TextSubtle).Italic(true).Render("system")
		body := lipgloss.NewStyle().Foreground(t.TextSubtle).Italic(true).Render(msg.Content)
		return mv.styles.AIBlock.Width(contentWidth).Render(label + "\n" + body)

	default:
		return msg.Content
	}
}

// renderThinking renders the model's reasoning in a subtle, left-bordered block.
func (mv MessageView) renderThinking(reasoning string, isStreaming bool) string {
	t := mv.styles.theme

	label := mv.styles.ThinkingLabel.Render("thinking")
	if isStreaming {
		label += lipgloss.NewStyle().Foreground(t.Warning).Render(" ●")
	}

	body := lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Italic(true).
		Render(reasoning)

	return mv.styles.ThinkingBlock.Render(label + "\n" + body)
}

// RenderWelcome returns the welcome screen shown before any conversation.
func RenderWelcome(styles Styles, width int) string {
	t := styles.theme

	title := lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Render("✦ clai")

	sub := lipgloss.NewStyle().
		Foreground(t.TextMuted).
		Render("terminal chat for LLMs")

	tip := lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Render("ctrl+o  configure api key & model")

	_ = width
	return "\n  " + title + "  " + sub + "\n  " + tip + "\n"
}
