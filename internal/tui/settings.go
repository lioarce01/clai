package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/clai/internal/config"
)

// SettingsSavedMsg is sent when the user saves settings.
type SettingsSavedMsg struct {
	Config *config.Config
}

// Settings is the configuration panel overlay.
type Settings struct {
	styles  Styles
	cfg     *config.Config
	inputs  []textinput.Model
	labels  []string
	focused int
	width   int
	height  int
	err     string
}

const (
	fieldAPIKey = iota
	fieldBaseURL
	fieldModel
	fieldTemperature
	fieldMaxTokens
	fieldTopP
	fieldSystemPrompt
	numFields
)

func NewSettings(styles Styles, cfg *config.Config, width, height int) Settings {
	inputs := make([]textinput.Model, numFields)
	labels := []string{
		"API Key",
		"Base URL",
		"Model",
		"Temperature",
		"Max Tokens",
		"Top P",
		"System Prompt",
	}

	for i := range inputs {
		ti := textinput.New()
		ti.CharLimit = 512
		ti.Prompt = "  "
		ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.theme.Primary)
		ti.TextStyle = lipgloss.NewStyle().Foreground(styles.theme.Text)
		ti.Cursor.Style = lipgloss.NewStyle().Foreground(styles.theme.Primary)
		inputs[i] = ti
	}

	// Mask API key
	inputs[fieldAPIKey].EchoMode = textinput.EchoPassword
	inputs[fieldAPIKey].EchoCharacter = '•'

	// Populate from config
	inputs[fieldAPIKey].SetValue(cfg.API.APIKey)
	inputs[fieldBaseURL].SetValue(cfg.API.BaseURL)
	inputs[fieldModel].SetValue(cfg.Model.Name)
	inputs[fieldTemperature].SetValue(fmt.Sprintf("%.2f", cfg.Model.Temperature))
	inputs[fieldMaxTokens].SetValue(fmt.Sprintf("%d", cfg.Model.MaxTokens))
	inputs[fieldTopP].SetValue(fmt.Sprintf("%.2f", cfg.Model.TopP))
	inputs[fieldSystemPrompt].SetValue(cfg.Model.SystemPrompt)

	inputs[0].Focus()

	return Settings{
		styles:  styles,
		cfg:     cfg,
		inputs:  inputs,
		labels:  labels,
		focused: 0,
		width:   width,
		height:  height,
	}
}

func (s Settings) Init() tea.Cmd { return textinput.Blink }

func (s Settings) Update(msg tea.Msg) (Settings, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "tab", "down":
			s.inputs[s.focused].Blur()
			s.focused = (s.focused + 1) % numFields
			s.inputs[s.focused].Focus()
			return s, textinput.Blink

		case "shift+tab", "up":
			s.inputs[s.focused].Blur()
			s.focused = (s.focused - 1 + numFields) % numFields
			s.inputs[s.focused].Focus()
			return s, textinput.Blink

		case "enter":
			cfg, err := s.buildConfig()
			if err != nil {
				s.err = err.Error()
				return s, nil
			}
			s.err = ""
			return s, func() tea.Msg {
				return SettingsSavedMsg{Config: cfg}
			}
		}
	}

	var cmd tea.Cmd
	s.inputs[s.focused], cmd = s.inputs[s.focused].Update(msg)
	return s, cmd
}

func (s *Settings) buildConfig() (*config.Config, error) {
	temp, err := strconv.ParseFloat(s.inputs[fieldTemperature].Value(), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid temperature: %w", err)
	}
	maxTok, err := strconv.Atoi(s.inputs[fieldMaxTokens].Value())
	if err != nil {
		return nil, fmt.Errorf("invalid max tokens: %w", err)
	}
	topP, err := strconv.ParseFloat(s.inputs[fieldTopP].Value(), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid top P: %w", err)
	}

	cfg := *s.cfg
	cfg.API.APIKey = s.inputs[fieldAPIKey].Value()
	cfg.API.BaseURL = s.inputs[fieldBaseURL].Value()
	cfg.Model.Name = s.inputs[fieldModel].Value()
	cfg.Model.Temperature = temp
	cfg.Model.MaxTokens = maxTok
	cfg.Model.TopP = topP
	cfg.Model.SystemPrompt = s.inputs[fieldSystemPrompt].Value()

	return &cfg, nil
}

func (s Settings) View() string {
	t := s.styles.theme

	title := lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Render("⚙  Settings")

	var rows []string
	rows = append(rows, title, "")

	for i, input := range s.inputs {
		label := s.labels[i]
		var labelStyle lipgloss.Style
		if i == s.focused {
			labelStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
		} else {
			labelStyle = lipgloss.NewStyle().Foreground(t.TextMuted)
		}

		var inputStyle lipgloss.Style
		if i == s.focused {
			inputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(t.BorderFocused)
		} else {
			inputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(t.BorderUnfocused)
		}

		inputWidth := 40
		rows = append(rows,
			labelStyle.Render(label),
			inputStyle.Width(inputWidth).Render(input.View()),
			"",
		)
	}

	if s.err != "" {
		rows = append(rows, s.styles.TextError.Render("Error: "+s.err))
	}

	hint := lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Render("[Tab] next  [Shift+Tab] prev  [Enter] save  [Esc] close")
	rows = append(rows, "", hint)

	content := strings.Join(rows, "\n")

	overlayWidth := 55
	if s.width < overlayWidth+10 {
		overlayWidth = s.width - 10
	}

	box := s.styles.OverlayBox.Width(overlayWidth).Render(content)

	leftPad := (s.width - lipgloss.Width(box)) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	topPad := (s.height - lipgloss.Height(box)) / 3
	if topPad < 0 {
		topPad = 0
	}

	var lines []string
	for i := 0; i < topPad; i++ {
		lines = append(lines, "")
	}
	for _, line := range strings.Split(box, "\n") {
		lines = append(lines, strings.Repeat(" ", leftPad)+line)
	}

	return strings.Join(lines, "\n")
}
