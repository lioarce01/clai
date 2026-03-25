package tui

import (
	"strings"
)

// renderUserMessage renders a user message with a bold `> ` prefix, right-indented.
// width is the available terminal width (used for wrapping hints, not enforced here
// since tview handles line wrapping in the TextView).
func renderUserMessage(content string, _ int) string {
	if content == "" {
		return ""
	}

	var sb strings.Builder
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if i == 0 {
			sb.WriteString(Bold + "  > " + line + Reset)
		} else {
			sb.WriteString("\n" + Bold + "    " + line + Reset)
		}
	}
	return sb.String()
}

// renderAssistantMessage renders an assistant message with optional thinking block.
// isStreaming adds a blinking cursor when content is empty.
func renderAssistantMessage(content string, reasoning string, isStreaming bool, _ int) string {
	var sb strings.Builder

	// Thinking block
	if reasoning != "" {
		sb.WriteString(renderThinkingBlock(reasoning, isStreaming && content == ""))
		sb.WriteString("\n")
	}

	// Main content
	if content != "" {
		lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
		for i, line := range lines {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString("  " + line)
		}
	} else if isStreaming {
		if reasoning == "" {
			// Nothing received yet — show cursor
			sb.WriteString("  " + Dim + "▋" + Reset)
		}
		// If reasoning is streaming, the thinking block already shows the cursor
	}

	return sb.String()
}

// renderThinkingBlock renders the reasoning/thinking block in dim+italic.
func renderThinkingBlock(reasoning string, showCursor bool) string {
	var sb strings.Builder

	// Header line
	label := Dim + Italic + "  thinking" + Reset
	if showCursor {
		label += " " + Dim + "▋" + Reset
	}
	sb.WriteString(label)

	// Body lines — deeper indent, dim+italic
	lines := strings.Split(strings.TrimRight(reasoning, "\n"), "\n")
	for _, line := range lines {
		sb.WriteString("\n")
		sb.WriteString(Dim + Italic + "    " + line + Reset)
	}

	return sb.String()
}

// renderWelcome returns the welcome message shown before any conversation.
func renderWelcome() string {
	return "\n" +
		"  " + Bold + "clai" + Reset + "  " + Dim + "terminal chat for LLMs" + Reset + "\n" +
		"  " + Dim + "ctrl+o  configure api key & model" + Reset + "\n"
}

// renderSeparator returns a visual separator line.
func renderSeparator(width int) string {
	if width <= 0 {
		width = 80
	}
	return strings.Repeat("─", width)
}
