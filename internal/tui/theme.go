package tui

import "github.com/charmbracelet/lipgloss"

// Theme holds the full color palette for the application.
type Theme struct {
	Text        lipgloss.AdaptiveColor
	TextMuted   lipgloss.AdaptiveColor
	TextSubtle  lipgloss.AdaptiveColor
	Primary     lipgloss.AdaptiveColor
	UserAccent  lipgloss.AdaptiveColor
	AIAccent    lipgloss.AdaptiveColor
	Error       lipgloss.AdaptiveColor
	Success     lipgloss.AdaptiveColor
	Warning     lipgloss.AdaptiveColor
	Border      lipgloss.AdaptiveColor
	BorderFaint lipgloss.AdaptiveColor
	Surface     lipgloss.AdaptiveColor
}

func DefaultTheme() Theme {
	return Theme{
		Text:        lipgloss.AdaptiveColor{Dark: "#E2E8F0", Light: "#1E293B"},
		TextMuted:   lipgloss.AdaptiveColor{Dark: "#94A3B8", Light: "#64748B"},
		TextSubtle:  lipgloss.AdaptiveColor{Dark: "#475569", Light: "#94A3B8"},
		Primary:     lipgloss.AdaptiveColor{Dark: "#818CF8", Light: "#4F46E5"},
		UserAccent:  lipgloss.AdaptiveColor{Dark: "#38BDF8", Light: "#0284C7"},
		AIAccent:    lipgloss.AdaptiveColor{Dark: "#C084FC", Light: "#9333EA"},
		Error:       lipgloss.AdaptiveColor{Dark: "#F87171", Light: "#DC2626"},
		Success:     lipgloss.AdaptiveColor{Dark: "#34D399", Light: "#059669"},
		Warning:     lipgloss.AdaptiveColor{Dark: "#FBBF24", Light: "#D97706"},
		Border:      lipgloss.AdaptiveColor{Dark: "#334155", Light: "#CBD5E1"},
		BorderFaint: lipgloss.AdaptiveColor{Dark: "#1E293B", Light: "#E2E8F0"},
		Surface:     lipgloss.AdaptiveColor{Dark: "#0F172A", Light: "#F8FAFC"},
	}
}
