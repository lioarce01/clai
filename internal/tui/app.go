package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/lioarce01/clai/internal/config"
	"github.com/lioarce01/clai/internal/llm"
	"github.com/lioarce01/clai/internal/markdown"
	"github.com/lioarce01/clai/internal/storage"
	"github.com/rivo/tview"
)

const (
	pageMain     = "main"
	pageSessions = "sessions"
	pageSettings = "settings"
)

// App is the root tview application.
type App struct {
	tviewApp *tview.Application

	// Layout primitives
	pages      *tview.Pages
	header     *tview.TextView
	footer     *tview.TextView
	chatView   *tview.TextView
	inputField *tview.InputField

	// Dependencies
	cfg       *config.Config
	llmClient llm.Client
	store     storage.Store
	renderer  *markdown.Renderer

	// Chat state — all access on the tview goroutine (via QueueUpdateDraw)
	currentSession   *storage.Session
	completedMsgs    []string // rendered completed messages
	streamContent    string   // accumulating streaming content
	streamReasoning  string   // accumulating streaming reasoning
	streaming        bool
	cancelStream     context.CancelFunc
	promptTokens     int
	completionTokens int
}

// New creates a new App.
func New(cfg *config.Config, llmClient llm.Client, store storage.Store) (*App, error) {
	renderer, err := markdown.New(76)
	if err != nil {
		return nil, fmt.Errorf("create markdown renderer: %w", err)
	}

	a := &App{
		cfg:       cfg,
		llmClient: llmClient,
		store:     store,
		renderer:  renderer,
	}
	a.buildUI()
	return a, nil
}

// Run starts the tview event loop. It blocks until the user quits.
func (a *App) Run() error {
	// Load or create default session in a goroutine, then update UI safely.
	go func() {
		sess, err := a.loadOrCreateDefault()
		a.tviewApp.QueueUpdateDraw(func() {
			if err != nil {
				a.setFooterError(err.Error())
				return
			}
			a.activateSession(sess)
		})
	}()

	return a.tviewApp.Run()
}

// ── UI construction ─────────────────────────────────────────────────────────

func (a *App) buildUI() {
	a.tviewApp = tview.NewApplication()
	a.tviewApp.EnableMouse(true)

	// Header
	a.header = tview.NewTextView()
	a.header.SetDynamicColors(true)
	a.header.SetBackgroundColor(tcell.ColorDefault)
	a.header.SetTextColor(tcell.ColorDefault)

	// Footer
	a.footer = tview.NewTextView()
	a.footer.SetDynamicColors(true)
	a.footer.SetBackgroundColor(tcell.ColorDefault)
	a.footer.SetTextColor(tcell.ColorDefault)

	// Chat view
	a.chatView = tview.NewTextView()
	a.chatView.SetDynamicColors(true)
	a.chatView.SetScrollable(true)
	a.chatView.SetWrap(true)
	a.chatView.SetWordWrap(true)
	a.chatView.SetBackgroundColor(tcell.ColorDefault)
	a.chatView.SetTextColor(tcell.ColorDefault)
	a.chatView.SetChangedFunc(func() {
		a.tviewApp.Draw()
	})

	// Write welcome message through ANSI writer
	fmt.Fprint(tview.ANSIWriter(a.chatView), renderWelcome())

	// Input field
	a.inputField = tview.NewInputField()
	a.inputField.SetLabel("")
	a.inputField.SetPlaceholder("Type a message…")
	a.inputField.SetBackgroundColor(tcell.ColorDefault)
	a.inputField.SetFieldBackgroundColor(tcell.ColorDefault)
	a.inputField.SetFieldTextColor(tcell.ColorDefault)
	a.inputField.SetPlaceholderTextColor(tcell.ColorDefault)
	a.inputField.SetLabelColor(tcell.ColorDefault)

	a.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			a.sendMessage()
		}
	})

	// Separator lines (simple TextViews with "─" characters)
	sep1 := tview.NewTextView()
	sep1.SetDynamicColors(false)
	sep1.SetBackgroundColor(tcell.ColorDefault)
	sep1.SetTextColor(tcell.ColorDefault)
	sep1.SetText(strings.Repeat("─", 200)) // over-wide; tview clips to terminal width

	sep2 := tview.NewTextView()
	sep2.SetDynamicColors(false)
	sep2.SetBackgroundColor(tcell.ColorDefault)
	sep2.SetTextColor(tcell.ColorDefault)
	sep2.SetText(strings.Repeat("─", 200))

	sep3 := tview.NewTextView()
	sep3.SetDynamicColors(false)
	sep3.SetBackgroundColor(tcell.ColorDefault)
	sep3.SetTextColor(tcell.ColorDefault)
	sep3.SetText(strings.Repeat("─", 200))

	// Input wrapper (separator + field)
	inputFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(sep2, 1, 0, false).
		AddItem(a.inputField, 1, 0, true).
		AddItem(sep3, 1, 0, false)

	// Main flex layout
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.header, 1, 0, false).
		AddItem(sep1, 1, 0, false).
		AddItem(a.chatView, 0, 1, false).
		AddItem(inputFlex, 3, 0, true).
		AddItem(a.footer, 1, 0, false)

	a.pages = tview.NewPages()
	a.pages.AddPage(pageMain, mainFlex, true, true)

	a.tviewApp.SetRoot(a.pages, true)
	a.tviewApp.SetFocus(a.inputField)

	// Global key bindings
	a.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Only act on the main page (no overlay active)
		name, _ := a.pages.GetFrontPage()

		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyCtrlQ:
			if a.cancelStream != nil {
				a.cancelStream()
			}
			a.tviewApp.Stop()
			return nil

		case tcell.KeyEscape:
			if name != pageMain {
				a.closeOverlay()
				return nil
			}

		case tcell.KeyCtrlN:
			if name == pageMain {
				go a.newSessionAsync()
				return nil
			}

		case tcell.KeyCtrlS:
			if name == pageMain {
				go a.openSessionPickerAsync()
				return nil
			}

		case tcell.KeyCtrlO:
			if name == pageMain {
				a.openSettings()
				return nil
			}

		case tcell.KeyCtrlL:
			if name == pageMain {
				a.clearChatView()
				return nil
			}

		// Keyboard scroll — works in any terminal including MinGW
		case tcell.KeyPgUp:
			if name == pageMain {
				row, col := a.chatView.GetScrollOffset()
				a.chatView.ScrollTo(row-15, col)
				return nil
			}
		case tcell.KeyPgDn:
			if name == pageMain {
				row, col := a.chatView.GetScrollOffset()
				a.chatView.ScrollTo(row+15, col)
				return nil
			}
		case tcell.KeyUp:
			if name == pageMain {
				row, col := a.chatView.GetScrollOffset()
				a.chatView.ScrollTo(row-3, col)
				return nil
			}
		case tcell.KeyDown:
			if name == pageMain {
				row, col := a.chatView.GetScrollOffset()
				a.chatView.ScrollTo(row+3, col)
				return nil
			}
		case tcell.KeyHome:
			if name == pageMain {
				a.chatView.ScrollToBeginning()
				return nil
			}
		case tcell.KeyEnd:
			if name == pageMain {
				a.chatView.ScrollToEnd()
				return nil
			}
		}
		return event
	})

	// Initial header/footer render
	a.renderHeader()
	a.renderFooter("", 0, 0)
}

// ── Header / Footer ─────────────────────────────────────────────────────────

func (a *App) renderHeader() {
	modelName := a.cfg.Model.Name
	left := Bold + "clai" + Reset
	right := Dim + modelName + Reset
	// Use a spacer that tview will naturally handle via text alignment trick:
	// write left-aligned text then right-aligned via padding-between
	text := left + "  " + Dim + "·" + Reset + "  " + right
	a.header.Clear()
	fmt.Fprint(tview.ANSIWriter(a.header), text)
}

func (a *App) renderFooter(sessionName string, promptToks, completionToks int) {
	left := Dim + sessionName + Reset
	if promptToks > 0 || completionToks > 0 {
		left += Dim + fmt.Sprintf("  %d↑ %d↓", promptToks, completionToks) + Reset
	}
	right := Dim + "^N ^S ^O ^C" + Reset
	text := left + "  " + right
	a.footer.Clear()
	fmt.Fprint(tview.ANSIWriter(a.footer), text)
}

func (a *App) setFooterError(msg string) {
	a.footer.Clear()
	fmt.Fprint(tview.ANSIWriter(a.footer), Bold+"error: "+Reset+Dim+msg+Reset)
}

// ── Chat view helpers ────────────────────────────────────────────────────────

// rebuildChatView rewrites the entire chatView content from completedMsgs +
// optional streaming state.  Must be called from within QueueUpdateDraw.
func (a *App) rebuildChatView() {
	_, _, w, _ := a.chatView.GetInnerRect()
	if w <= 0 {
		w = 80
	}

	var sb strings.Builder

	if len(a.completedMsgs) == 0 && !a.streaming {
		sb.WriteString(renderWelcome())
	} else {
		for _, rendered := range a.completedMsgs {
			sb.WriteString(rendered)
			sb.WriteString("\n\n")
		}
		if a.streaming {
			msg := renderAssistantMessage(a.streamContent, a.streamReasoning, true, w)
			sb.WriteString(msg)
			sb.WriteString("\n")
		}
	}

	a.chatView.Clear()
	fmt.Fprint(tview.ANSIWriter(a.chatView), sb.String())
	a.chatView.ScrollToEnd()
}

func (a *App) clearChatView() {
	a.chatView.Clear()
	fmt.Fprint(tview.ANSIWriter(a.chatView), renderWelcome())
}

// ── Session management ───────────────────────────────────────────────────────

func (a *App) loadOrCreateDefault() (*storage.Session, error) {
	sessions, err := a.store.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	if len(sessions) == 0 {
		sess, err := a.store.CreateSession("New Session")
		if err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}
		return sess, nil
	}
	sess, err := a.store.GetSession(sessions[0].ID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return sess, nil
}

// activateSession switches to the given session.  Must be called on the tview
// goroutine (or from within QueueUpdateDraw).
func (a *App) activateSession(sess *storage.Session) {
	a.currentSession = sess
	a.completedMsgs = nil
	a.streamContent = ""
	a.streamReasoning = ""
	a.streaming = false

	_, _, w, _ := a.chatView.GetInnerRect()
	if w <= 0 {
		w = 80
	}

	for _, msg := range sess.Messages {
		rendered := a.renderMessage(msg, false, w)
		if rendered != "" {
			a.completedMsgs = append(a.completedMsgs, rendered)
		}
	}

	a.rebuildChatView()
	a.renderHeader()
	a.renderFooter(sess.Name, a.promptTokens, a.completionTokens)
}

func (a *App) newSessionAsync() {
	sess, err := a.store.CreateSession("New Session")
	a.tviewApp.QueueUpdateDraw(func() {
		if err != nil {
			a.setFooterError(err.Error())
			return
		}
		a.closeOverlay()
		a.activateSession(sess)
	})
}

func (a *App) openSessionPickerAsync() {
	sessions, err := a.store.ListSessions()
	a.tviewApp.QueueUpdateDraw(func() {
		if err != nil {
			a.setFooterError(err.Error())
			return
		}
		activeID := ""
		if a.currentSession != nil {
			activeID = a.currentSession.ID
		}
		modal := buildSessionModal(
			sessions,
			activeID,
			func(id string) {
				a.closeOverlay()
				go func() {
					sess, err := a.store.GetSession(id)
					a.tviewApp.QueueUpdateDraw(func() {
						if err != nil {
							a.setFooterError(err.Error())
							return
						}
						a.activateSession(sess)
					})
				}()
			},
			func(id string) {
				go func() {
					_ = a.store.DeleteSession(id)
					newSessions, _ := a.store.ListSessions()
					a.tviewApp.QueueUpdateDraw(func() {
						a.closeOverlay()
						if len(newSessions) == 0 {
							go a.newSessionAsync()
							return
						}
						go func() {
							sess, err := a.store.GetSession(newSessions[0].ID)
							a.tviewApp.QueueUpdateDraw(func() {
								if err == nil {
									a.activateSession(sess)
								}
							})
						}()
					})
				}()
			},
			func() {
				a.closeOverlay()
				go a.newSessionAsync()
			},
			func() { a.closeOverlay() },
		)
		a.showOverlay(pageSessions, modal)
	})
}

func (a *App) openSettings() {
	modal := buildSettingsModal(
		a.cfg,
		func(newCfg *config.Config) {
			a.cfg = newCfg
			_ = config.Save(newCfg)
			a.llmClient = llm.NewClient(newCfg.API.APIKey, newCfg.API.BaseURL)
			a.closeOverlay()
			a.renderHeader()
		},
		func() { a.closeOverlay() },
	)
	a.showOverlay(pageSettings, modal)
}

// showOverlay adds a centered overlay page and brings it to the front.
func (a *App) showOverlay(name string, content tview.Primitive) {
	// Wrap in a Flex to center the modal
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(content, 0, 3, true).
				AddItem(nil, 0, 1, false),
			0, 2, true,
		).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage(name, centered, true, true)
	a.tviewApp.SetFocus(content)
}

func (a *App) closeOverlay() {
	name, _ := a.pages.GetFrontPage()
	if name != pageMain {
		a.pages.RemovePage(name)
	}
	a.tviewApp.SetFocus(a.inputField)
}

// ── Message sending & streaming ──────────────────────────────────────────────

func (a *App) sendMessage() {
	if a.currentSession == nil || a.streaming {
		return
	}
	content := strings.TrimSpace(a.inputField.GetText())
	if content == "" {
		return
	}
	a.inputField.SetText("")

	userMsg := llm.Message{
		ID:        uuid.New().String(),
		Role:      llm.RoleUser,
		Content:   content,
		CreatedAt: time.Now(),
	}

	// Persist user message
	if err := a.store.AddMessage(a.currentSession.ID, userMsg); err != nil {
		a.setFooterError(err.Error())
		return
	}

	// Refresh session
	if sess, err := a.store.GetSession(a.currentSession.ID); err == nil {
		a.currentSession = sess
	}

	// Render and add to history
	_, _, w, _ := a.chatView.GetInnerRect()
	if w <= 0 {
		w = 80
	}
	a.completedMsgs = append(a.completedMsgs, a.renderMessage(userMsg, false, w))

	// Begin streaming
	a.streaming = true
	a.streamContent = ""
	a.streamReasoning = ""
	a.rebuildChatView()
	a.renderFooter(a.currentSession.Name, a.promptTokens, a.completionTokens)

	// Snapshot for goroutine
	msgs := make([]llm.Message, len(a.currentSession.Messages))
	copy(msgs, a.currentSession.Messages)
	cfg := a.cfg

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	a.cancelStream = cancel

	go func() {
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
			a.tviewApp.QueueUpdateDraw(func() {
				a.streaming = false
				a.setFooterError(err.Error())
			})
			return
		}

		for delta := range ch {
			if delta.Error != nil {
				cancel()
				a.tviewApp.QueueUpdateDraw(func() {
					a.streaming = false
					a.setFooterError(delta.Error.Error())
				})
				return
			}
			if delta.Done {
				usage := delta.Usage
				a.tviewApp.QueueUpdateDraw(func() {
					a.finishStream(usage)
				})
				return
			}
			// Accumulate and redraw
			contentChunk := delta.Content
			reasoningChunk := delta.Reasoning
			a.tviewApp.QueueUpdateDraw(func() {
				a.streamContent += contentChunk
				a.streamReasoning += reasoningChunk
				a.rebuildChatView()
			})
		}
		// Channel closed without Done — treat as finished
		a.tviewApp.QueueUpdateDraw(func() {
			a.finishStream(nil)
		})
	}()
}

// finishStream finalizes the streaming assistant message.
// Must be called from within QueueUpdateDraw.
func (a *App) finishStream(usage *llm.Usage) {
	if !a.streaming {
		return
	}
	a.streaming = false

	if a.cancelStream != nil {
		a.cancelStream()
		a.cancelStream = nil
	}

	// Build final llm.Message and persist
	finalMsg := llm.Message{
		ID:        uuid.New().String(),
		Role:      llm.RoleAssistant,
		Content:   a.streamContent,
		Reasoning: a.streamReasoning,
		CreatedAt: time.Now(),
	}

	if a.currentSession != nil {
		_ = a.store.AddMessage(a.currentSession.ID, finalMsg)
		// Refresh session (for potential auto-rename)
		if sess, err := a.store.GetSession(a.currentSession.ID); err == nil {
			a.currentSession = sess
		}
	}

	// Move streamed content into completed messages
	_, _, w, _ := a.chatView.GetInnerRect()
	if w <= 0 {
		w = 80
	}
	rendered := a.renderMessage(finalMsg, false, w)
	if rendered != "" {
		a.completedMsgs = append(a.completedMsgs, rendered)
	}
	a.streamContent = ""
	a.streamReasoning = ""

	if usage != nil {
		a.promptTokens = usage.PromptTokens
		a.completionTokens = usage.CompletionTokens
	}

	a.rebuildChatView()
	sessionName := ""
	if a.currentSession != nil {
		sessionName = a.currentSession.Name
	}
	a.renderFooter(sessionName, a.promptTokens, a.completionTokens)
}

// ── Message rendering ────────────────────────────────────────────────────────

// renderMessage converts an llm.Message to a styled string for display.
func (a *App) renderMessage(msg llm.Message, isStreaming bool, width int) string {
	switch msg.Role {
	case llm.RoleUser:
		// Use markdown renderer for content, then apply user formatting
		rendered, err := a.renderer.Render(msg.Content)
		if err != nil {
			rendered = msg.Content
		}
		rendered = strings.TrimSpace(rendered)
		return renderUserMessage(rendered, width)

	case llm.RoleAssistant:
		// Use markdown renderer for content
		displayContent := msg.Content
		if displayContent != "" {
			rendered, err := a.renderer.Render(msg.Content)
			if err == nil {
				displayContent = strings.TrimSpace(rendered)
			}
		}
		return renderAssistantMessage(displayContent, msg.Reasoning, isStreaming, width)

	case llm.RoleSystem:
		return Dim + Italic + "  [system] " + msg.Content + Reset

	default:
		return msg.Content
	}
}
