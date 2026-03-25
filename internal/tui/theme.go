package tui

import "github.com/charmbracelet/lipgloss"

// Theme holds the full color palette for the application.
type Theme struct {
	// Backgrounds
	Background lipgloss.AdaptiveColor
	Surface    lipgloss.AdaptiveColor
	SurfaceAlt lipgloss.AdaptiveColor

	// Text
	Text       lipgloss.AdaptiveColor
	TextMuted  lipgloss.AdaptiveColor
	TextSubtle lipgloss.AdaptiveColor

	// Accents
	Primary    lipgloss.AdaptiveColor
	PrimaryDim lipgloss.AdaptiveColor

	// Roles
	UserBadge      lipgloss.AdaptiveColor
	AssistantBadge lipgloss.AdaptiveColor
	SystemBadge    lipgloss.AdaptiveColor

	// State
	Error   lipgloss.AdaptiveColor
	Success lipgloss.AdaptiveColor
	Warning lipgloss.AdaptiveColor

	// Borders
	BorderFocused   lipgloss.AdaptiveColor
	BorderUnfocused lipgloss.AdaptiveColor
}

// DefaultTheme returns the default dark/light adaptive theme.
func DefaultTheme() Theme {
	return Theme{
		Background: lipgloss.AdaptiveColor{Dark: "#0D1117", Light: "#FFFFFF"},
		Surface:    lipgloss.AdaptiveColor{Dark: "#161B22", Light: "#F6F8FA"},
		SurfaceAlt: lipgloss.AdaptiveColor{Dark: "#1F2937", Light: "#E5E7EB"},

		Text:       lipgloss.AdaptiveColor{Dark: "#E6EDF3", Light: "#1F2937"},
		TextMuted:  lipgloss.AdaptiveColor{Dark: "#8B949E", Light: "#6B7280"},
		TextSubtle: lipgloss.AdaptiveColor{Dark: "#6B7280", Light: "#9CA3AF"},

		Primary:    lipgloss.AdaptiveColor{Dark: "#7C3AED", Light: "#6D28D9"},
		PrimaryDim: lipgloss.AdaptiveColor{Dark: "#5B21B6", Light: "#7C3AED"},

		UserBadge:      lipgloss.AdaptiveColor{Dark: "#06B6D4", Light: "#0891B2"},
		AssistantBadge: lipgloss.AdaptiveColor{Dark: "#A855F7", Light: "#9333EA"},
		SystemBadge:    lipgloss.AdaptiveColor{Dark: "#6B7280", Light: "#9CA3AF"},

		Error:   lipgloss.AdaptiveColor{Dark: "#EF4444", Light: "#DC2626"},
		Success: lipgloss.AdaptiveColor{Dark: "#10B981", Light: "#059669"},
		Warning: lipgloss.AdaptiveColor{Dark: "#F59E0B", Light: "#D97706"},

		BorderFocused:   lipgloss.AdaptiveColor{Dark: "#7C3AED", Light: "#6D28D9"},
		BorderUnfocused: lipgloss.AdaptiveColor{Dark: "#374151", Light: "#D1D5DB"},
	}
}
