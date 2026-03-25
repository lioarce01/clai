package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lioarce01/clai/internal/storage"
)

// SessionSelectedMsg is sent when the user selects a session.
type SessionSelectedMsg struct{ ID string }

// SessionDeleteMsg is sent when the user deletes a session.
type SessionDeleteMsg struct{ ID string }

// NewSessionMsg is sent when the user requests a new session.
type NewSessionMsg struct{}

// sessionItem wraps storage.Session for the bubbles/list component.
type sessionItem struct {
	session storage.Session
}

func (si sessionItem) FilterValue() string { return si.session.Name }
func (si sessionItem) Title() string       { return si.session.Name }
func (si sessionItem) Description() string {
	count := len(si.session.Messages)
	return fmt.Sprintf("%d messages • %s", count, si.session.UpdatedAt.Format("Jan 2, 15:04"))
}

// sessionDelegate renders each list item.
type sessionDelegate struct {
	styles Styles
}

func (d sessionDelegate) Height() int                              { return 2 }
func (d sessionDelegate) Spacing() int                            { return 1 }
func (d sessionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d sessionDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	si, ok := item.(sessionItem)
	if !ok {
		return
	}

	t := d.styles.theme
	isSelected := index == m.Index()

	title := si.Title()
	desc := si.Description()

	if isSelected {
		title = lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true).
			Render("> " + title)
		desc = lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Render("  " + desc)
	} else {
		title = lipgloss.NewStyle().
			Foreground(t.Text).
			Render("  " + title)
		desc = lipgloss.NewStyle().
			Foreground(t.TextSubtle).
			Render("  " + desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// SessionPicker is the session list overlay.
type SessionPicker struct {
	styles   Styles
	list     list.Model
	sessions []storage.Session
	width    int
	height   int
	active   string // currently active session ID
}

func NewSessionPicker(styles Styles, sessions []storage.Session, activeID string, width, height int) SessionPicker {
	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		items[i] = sessionItem{session: s}
	}

	delegate := sessionDelegate{styles: styles}
	l := list.New(items, delegate, width-6, height-8)
	l.Title = "Sessions"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(styles.theme.Primary).
		Bold(true)
	l.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(styles.theme.TextMuted)
	l.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(styles.theme.Primary)
	l.SetShowHelp(false)

	return SessionPicker{
		styles:   styles,
		list:     l,
		sessions: sessions,
		width:    width,
		height:   height,
		active:   activeID,
	}
}

func (sp SessionPicker) Init() tea.Cmd { return nil }

func (sp SessionPicker) Update(msg tea.Msg) (SessionPicker, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "enter":
			if item, ok := sp.list.SelectedItem().(sessionItem); ok {
				return sp, func() tea.Msg {
					return SessionSelectedMsg{ID: item.session.ID}
				}
			}
		case "n":
			return sp, func() tea.Msg { return NewSessionMsg{} }
		case "d":
			if item, ok := sp.list.SelectedItem().(sessionItem); ok {
				return sp, func() tea.Msg {
					return SessionDeleteMsg{ID: item.session.ID}
				}
			}
		}
	}

	var cmd tea.Cmd
	sp.list, cmd = sp.list.Update(msg)
	return sp, cmd
}

func (sp SessionPicker) View() string {
	t := sp.styles.theme

	hint := lipgloss.NewStyle().
		Foreground(t.TextSubtle).
		Render("[n]ew  [d]elete  [enter]select  [/]filter  [esc]close")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		sp.list.View(),
		"",
		hint,
	)

	overlayWidth := sp.width - 10
	if overlayWidth > 60 {
		overlayWidth = 60
	}

	box := sp.styles.OverlayBox.Width(overlayWidth).Render(content)

	// Center the overlay
	leftPad := (sp.width - lipgloss.Width(box)) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	topPad := (sp.height - lipgloss.Height(box)) / 2
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
