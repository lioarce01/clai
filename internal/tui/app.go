package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/lioarce01/clai/internal/config"
	"github.com/lioarce01/clai/internal/llm"
	"github.com/lioarce01/clai/internal/markdown"
	"github.com/lioarce01/clai/internal/storage"
)

// viewState represents the current UI mode.
type viewState int

const (
	viewChat viewState = iota
	viewSessions
	viewSettings
)

// streamTokenMsg carries a streaming delta from the LLM.
type streamTokenMsg struct {
	delta llm.StreamDelta
}

// sessionsLoadedMsg carries sessions loaded from storage.
type sessionsLoadedMsg struct {
	sessions []storage.Session
	err      error
}

// sessionActivatedMsg carries a fully loaded session to make active.
type sessionActivatedMsg struct {
	session *storage.Session
}

// sessionPickerReadyMsg signals that the session picker data is loaded and state should switch.
type sessionPickerReadyMsg struct {
	sessions []storage.Session
	activeID string
}

// App is the root Bubble Tea model.
type App struct {
	cfg       *config.Config
	llmClient llm.Client
	store     storage.Store
	renderer  *markdown.Renderer
	styles    Styles
	keymap    KeyMap

	// UI state
	state  viewState
	width  int
	height int
	err    string

	// Components
	chat      Chat
	sessions  SessionPicker
	settings  Settings
	header    Header
	statusBar StatusBar

	// Session state
	currentSession *storage.Session
	promptTokens   int
	completeTokens int
	streaming      bool
	cancelStream   context.CancelFunc

	// Active stream channel (for the canonical Bubble Tea channel pattern)
	streamCh <-chan llm.StreamDelta
}

// New creates a new App model.
func New(cfg *config.Config, llmClient llm.Client, store storage.Store) (*App, error) {
	renderer, err := markdown.New(80)
	if err != nil {
		return nil, fmt.Errorf("create markdown renderer: %w", err)
	}

	theme := DefaultTheme()
	styles := NewStyles(theme)
	keymap := DefaultKeyMap()

	// Initialize components with placeholder dimensions (resized on first WindowSizeMsg)
	chat := NewChat(styles, renderer, keymap, 80, 22)
	header := NewHeader(styles, "v0.1.0")
	header.SetModel(cfg.Model.Name)
	statusBar := NewStatusBar(styles)

	app := &App{
		cfg:       cfg,
		llmClient: llmClient,
		store:     store,
		renderer:  renderer,
		styles:    styles,
		keymap:    keymap,
		state:     viewChat,
		width:     80,
		height:    24,
		chat:      chat,
		header:    header,
		statusBar: statusBar,
	}

	return app, nil
}

// Init is called once when the program starts.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("CLAI"),
		a.loadOrCreateDefaultSession(),
	)
}

func (a *App) loadOrCreateDefaultSession() tea.Cmd {
	return func() tea.Msg {
		sessions, err := a.store.ListSessions()
		if err != nil {
			return sessionsLoadedMsg{err: err}
		}

		if len(sessions) == 0 {
			sess, err := a.store.CreateSession("New Session")
			if err != nil {
				return sessionsLoadedMsg{err: err}
			}
			return sessionsLoadedMsg{sessions: []storage.Session{*sess}}
		}

		return sessionsLoadedMsg{sessions: sessions}
	}
}

// waitForStreamDelta returns a Cmd that reads one delta from the stream channel.
func waitForStreamDelta(ch <-chan llm.StreamDelta) tea.Cmd {
	return func() tea.Msg {
		delta, ok := <-ch
		if !ok {
			return streamTokenMsg{delta: llm.StreamDelta{Done: true}}
		}
		return streamTokenMsg{delta: delta}
	}
}

// Update handles all incoming messages.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch m := msg.(type) {
	// ── Window resize ──────────────────────────────────────────────────────
	case tea.WindowSizeMsg:
		a.width = m.Width
		a.height = m.Height
		a.resizeAll()
		return a, nil

	// ── Quit ───────────────────────────────────────────────────────────────
	case tea.KeyMsg:
		if m.String() == "ctrl+c" {
			if a.cancelStream != nil {
				a.cancelStream()
			}
			return a, tea.Quit
		}

		// Overlay-specific keys
		if a.state == viewSessions || a.state == viewSettings {
			if m.String() == "esc" {
				a.state = viewChat
				return a, nil
			}
		}

		// Global keys (only in chat view)
		if a.state == viewChat {
			switch m.String() {
			case "ctrl+n":
				return a, a.createNewSession()
			case "ctrl+s":
				return a, a.loadSessionPickerCmd()
			case "ctrl+o":
				a.openSettings()
				return a, nil
			case "ctrl+l":
				a.chat.ClearView()
				return a, nil
			}
		}

	// ── Sessions loaded (initial load or after delete) ────────────────────
	case sessionsLoadedMsg:
		if m.err != nil {
			a.err = m.err.Error()
			return a, nil
		}
		if len(m.sessions) > 0 {
			return a, a.activateSession(m.sessions[0].ID)
		}
		return a, nil

	// ── Session picker data ready ─────────────────────────────────────────
	case sessionPickerReadyMsg:
		a.sessions = NewSessionPicker(a.styles, m.sessions, m.activeID, a.width, a.height)
		a.state = viewSessions
		return a, nil

	// ── Session activated ─────────────────────────────────────────────────
	case sessionActivatedMsg:
		a.currentSession = m.session
		a.statusBar.SetSessionName(m.session.Name)
		a.header.SetConnected(true)
		a.chat.LoadMessages(m.session.Messages)
		return a, nil

	// ── Send message ──────────────────────────────────────────────────────
	case SendMessageMsg:
		if a.currentSession == nil || a.streaming {
			return a, nil
		}
		return a, a.sendMessage(m.Content)

	// ── Stream token ──────────────────────────────────────────────────────
	case streamTokenMsg:
		if m.delta.Error != nil {
			a.streaming = false
			a.header.SetStreaming(false)
			a.err = m.delta.Error.Error()
			return a, nil
		}
		if m.delta.Done {
			a.streaming = false
			a.header.SetStreaming(false)
			a.streamCh = nil
			finalMsg := a.chat.FinishStream()
			if finalMsg != nil && a.currentSession != nil {
				_ = a.store.AddMessage(a.currentSession.ID, *finalMsg)
				// Refresh session to get the potentially auto-renamed name
				if sess, err := a.store.GetSession(a.currentSession.ID); err == nil {
					a.currentSession = sess
					a.statusBar.SetSessionName(sess.Name)
				}
			}
			if m.delta.Usage != nil {
				a.promptTokens = m.delta.Usage.PromptTokens
				a.completeTokens = m.delta.Usage.CompletionTokens
				a.statusBar.SetTokens(a.promptTokens, a.completeTokens)
			}
			return a, nil
		}
		// Append the token and schedule reading the next one
		a.chat.AppendStream(m.delta.Content, m.delta.Reasoning)
		if a.streamCh != nil {
			return a, waitForStreamDelta(a.streamCh)
		}
		return a, nil

	// ── Stream channel ready ─────────────────────────────────────────────
	case streamChanReadyMsg:
		a.streamCh = m.ch
		a.cancelStream = m.cancel
		return a, waitForStreamDelta(a.streamCh)

	// ── Session selected from picker ──────────────────────────────────────
	case SessionSelectedMsg:
		a.state = viewChat
		return a, a.activateSession(m.ID)

	// ── Session delete ────────────────────────────────────────────────────
	case SessionDeleteMsg:
		return a, a.deleteSession(m.ID)

	// ── New session ───────────────────────────────────────────────────────
	case NewSessionMsg:
		a.state = viewChat
		return a, a.createNewSession()

	// ── Settings saved ────────────────────────────────────────────────────
	case SettingsSavedMsg:
		a.cfg = m.Config
		_ = config.Save(a.cfg)
		a.llmClient = llm.NewClient(a.cfg.API.APIKey, a.cfg.API.BaseURL)
		a.header.SetModel(a.cfg.Model.Name)
		a.state = viewChat
		return a, nil
	}

	// Delegate to sub-components
	switch a.state {
	case viewChat:
		var chatCmd tea.Cmd
		a.chat, chatCmd = a.chat.Update(msg)
		cmds = append(cmds, chatCmd)

	case viewSessions:
		var spCmd tea.Cmd
		a.sessions, spCmd = a.sessions.Update(msg)
		cmds = append(cmds, spCmd)

	case viewSettings:
		var setCmd tea.Cmd
		a.settings, setCmd = a.settings.Update(msg)
		cmds = append(cmds, setCmd)
	}

	return a, tea.Batch(cmds...)
}

// View renders the full terminal UI.
func (a *App) View() string {
	if a.width == 0 {
		return "Loading…"
	}

	header := a.header.View()
	statusBar := a.statusBar.View()

	headerH := lipgloss.Height(header)
	statusH := lipgloss.Height(statusBar)
	bodyH := a.height - headerH - statusH
	if bodyH < 1 {
		bodyH = 1
	}

	var body string
	switch a.state {
	case viewChat:
		body = a.chat.View()
	case viewSessions:
		body = a.sessions.View()
	case viewSettings:
		body = a.settings.View()
	}

	// Show error banner if set
	if a.err != "" {
		errLine := a.styles.TextError.Width(a.width).Render("⚠ " + a.err)
		body = errLine + "\n" + body
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
		statusBar,
	)
}

// ── Private helpers ────────────────────────────────────────────────────────

func (a *App) resizeAll() {
	a.header.SetWidth(a.width)
	a.statusBar.SetWidth(a.width)
	headerH := 1
	statusH := 1
	bodyH := a.height - headerH - statusH
	if bodyH < 1 {
		bodyH = 1
	}
	a.chat.Resize(a.width, bodyH)
	a.renderer.SetWidth(a.width - 6)
}

func (a *App) activateSession(id string) tea.Cmd {
	return func() tea.Msg {
		sess, err := a.store.GetSession(id)
		if err != nil {
			return sessionsLoadedMsg{err: err}
		}
		return sessionActivatedMsg{session: sess}
	}
}

func (a *App) sendMessage(content string) tea.Cmd {
	// Build user message
	userMsg := llm.Message{
		ID:        uuid.New().String(),
		Role:      llm.RoleUser,
		Content:   content,
		CreatedAt: time.Now(),
	}

	// Persist the user message
	if err := a.store.AddMessage(a.currentSession.ID, userMsg); err != nil {
		a.err = err.Error()
		return nil
	}

	// Reload session to pick up potential name change
	if sess, err := a.store.GetSession(a.currentSession.ID); err == nil {
		a.currentSession = sess
		a.statusBar.SetSessionName(sess.Name)
	}

	a.chat.AddMessage(userMsg)
	a.chat.StartStream()
	a.streaming = true
	a.header.SetStreaming(true)
	a.err = ""

	// Capture the messages for the LLM call (snapshot before async)
	msgs := make([]llm.Message, len(a.currentSession.Messages))
	copy(msgs, a.currentSession.Messages)

	cfg := a.cfg

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	a.cancelStream = cancel

	return func() tea.Msg {
		ch, err := a.llmClient.ChatCompletionStream(ctx, llm.CompletionParams{
			Model:        cfg.Model.Name,
			Messages:     msgs,
			Temperature:  cfg.Model.Temperature,
			MaxTokens:    cfg.Model.MaxTokens,
			TopP:         cfg.Model.TopP,
			SystemPrompt: cfg.Model.SystemPrompt,
		})
		if err != nil {
			cancel()
			return streamTokenMsg{delta: llm.StreamDelta{Error: err, Done: true}}
		}

		// Store the channel on the app — but we're inside a Cmd here.
		// We return a special msg to hand the channel back to the model.
		return streamChanReadyMsg{ch: ch, cancel: cancel}
	}
}

// streamChanReadyMsg is sent when the LLM stream channel is ready.
type streamChanReadyMsg struct {
	ch     <-chan llm.StreamDelta
	cancel context.CancelFunc
}

// createNewSession creates a new session and activates it.
func (a *App) createNewSession() tea.Cmd {
	return func() tea.Msg {
		sess, err := a.store.CreateSession("New Session")
		if err != nil {
			return sessionsLoadedMsg{err: err}
		}
		return sessionActivatedMsg{session: sess}
	}
}

// loadSessionPickerCmd loads sessions and signals the picker to open.
func (a *App) loadSessionPickerCmd() tea.Cmd {
	activeID := ""
	if a.currentSession != nil {
		activeID = a.currentSession.ID
	}
	return func() tea.Msg {
		sessions, err := a.store.ListSessions()
		if err != nil {
			return sessionsLoadedMsg{err: err}
		}
		return sessionPickerReadyMsg{sessions: sessions, activeID: activeID}
	}
}

func (a *App) openSettings() {
	a.settings = NewSettings(a.styles, a.cfg, a.width, a.height)
	a.state = viewSettings
}

func (a *App) deleteSession(id string) tea.Cmd {
	return func() tea.Msg {
		_ = a.store.DeleteSession(id)
		sessions, err := a.store.ListSessions()
		if err != nil {
			return sessionsLoadedMsg{err: err}
		}
		if len(sessions) == 0 {
			sess, _ := a.store.CreateSession("New Session")
			if sess != nil {
				sessions = []storage.Session{*sess}
			}
		}
		return sessionsLoadedMsg{sessions: sessions}
	}
}
