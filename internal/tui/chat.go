package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lioarce01/clai/internal/llm"
	"github.com/lioarce01/clai/internal/markdown"
)


// SendMessageMsg is sent when the user submits a message.
type SendMessageMsg struct {
	Content string
}

// Chat is the main chat view: message viewport + textarea input.
type Chat struct {
	styles       Styles
	renderer     *markdown.Renderer
	msgView      MessageView
	viewport     viewport.Model
	textarea     textarea.Model
	keymap       KeyMap
	messages     []llm.Message
	streamingMsg *llm.Message // current assistant message being streamed
	atBottom     bool
	width        int
	height       int
}

func NewChat(styles Styles, renderer *markdown.Renderer, keymap KeyMap, width, height int) Chat {
	ta := textarea.New()
	ta.Placeholder = "Type your message... (Enter to send, Shift+Enter for new line)"
	ta.Focus()
	ta.CharLimit = 0
	ta.SetWidth(width - 4)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetKeys("shift+enter")

	// Style the textarea
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(styles.theme.TextSubtle)
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(styles.theme.TextSubtle)

	inputHeight := 5 // textarea + border
	vpHeight := height - inputHeight - 1

	vp := viewport.New(width-2, vpHeight)
	vp.SetContent(RenderWelcome(styles, width-2))
	vp.GotoBottom()

	mv := NewMessageView(styles, renderer, width-2)

	return Chat{
		styles:   styles,
		renderer: renderer,
		msgView:  mv,
		viewport: vp,
		textarea: ta,
		keymap:   keymap,
		messages: []llm.Message{},
		atBottom: true,
		width:    width,
		height:   height,
	}
}

func (c Chat) Init() tea.Cmd {
	return textarea.Blink
}

func (c Chat) Update(msg tea.Msg) (Chat, tea.Cmd) {
	var cmds []tea.Cmd

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "enter":
			content := strings.TrimSpace(c.textarea.Value())
			if content != "" {
				c.textarea.Reset()
				return c, func() tea.Msg {
					return SendMessageMsg{Content: content}
				}
			}
			return c, nil
		case "pgup":
			c.viewport.HalfViewUp()
			c.atBottom = c.viewport.AtBottom()
			return c, nil
		case "pgdown":
			c.viewport.HalfViewDown()
			c.atBottom = c.viewport.AtBottom()
			return c, nil
		case "alt+up":
			c.viewport.LineUp(3)
			c.atBottom = c.viewport.AtBottom()
			return c, nil
		case "alt+down":
			c.viewport.LineDown(3)
			c.atBottom = c.viewport.AtBottom()
			return c, nil
		}

	case tea.MouseMsg:
		switch m.Button {
		case tea.MouseButtonWheelUp:
			c.viewport.LineUp(3)
			c.atBottom = c.viewport.AtBottom()
			return c, nil
		case tea.MouseButtonWheelDown:
			c.viewport.LineDown(3)
			c.atBottom = c.viewport.AtBottom()
			return c, nil
		}

	case tea.WindowSizeMsg:
		c.Resize(m.Width, m.Height)
		return c, nil
	}

	var taCmd tea.Cmd
	c.textarea, taCmd = c.textarea.Update(msg)
	cmds = append(cmds, taCmd)

	lineCount := c.textarea.LineCount()
	if lineCount < 1 {
		lineCount = 1
	}
	if lineCount > 8 {
		lineCount = 8
	}
	c.textarea.SetHeight(lineCount)

	return c, tea.Batch(cmds...)
}

// LoadMessages replaces the current message list and re-renders.
func (c *Chat) LoadMessages(msgs []llm.Message) {
	c.messages = msgs
	c.streamingMsg = nil
	c.rebuildViewport()
}

// AddMessage appends a message and re-renders.
func (c *Chat) AddMessage(msg llm.Message) {
	c.messages = append(c.messages, msg)
	c.rebuildViewport()
	if c.atBottom {
		c.viewport.GotoBottom()
	}
}

// StartStream begins a new streaming assistant message.
func (c *Chat) StartStream() {
	now := time.Now()
	msg := llm.Message{
		Role:      llm.RoleAssistant,
		Content:   "",
		CreatedAt: now,
	}
	c.streamingMsg = &msg
	c.rebuildViewport()
	c.viewport.GotoBottom()
}

// AppendStream adds content and/or reasoning to the ongoing streaming message.
func (c *Chat) AppendStream(content, reasoning string) {
	if c.streamingMsg == nil {
		c.StartStream()
	}
	c.streamingMsg.Content += content
	c.streamingMsg.Reasoning += reasoning
	c.rebuildViewport()
	if c.atBottom {
		c.viewport.GotoBottom()
	}
}

// FinishStream finalizes the streaming message and moves it into messages.
func (c *Chat) FinishStream() *llm.Message {
	if c.streamingMsg == nil {
		return nil
	}
	msg := *c.streamingMsg
	c.messages = append(c.messages, msg)
	c.streamingMsg = nil
	c.rebuildViewport()
	if c.atBottom {
		c.viewport.GotoBottom()
	}
	return &msg
}

// ClearView resets the viewport to the welcome screen without deleting messages.
func (c *Chat) ClearView() {
	c.viewport.SetContent(RenderWelcome(c.styles, c.width-2))
	c.viewport.GotoBottom()
}

// Resize updates the component dimensions.
func (c *Chat) Resize(width, height int) {
	c.width = width
	c.height = height
	c.textarea.SetWidth(width - 4)
	inputHeight := c.textarea.Height() + 4
	vpHeight := height - inputHeight
	if vpHeight < 1 {
		vpHeight = 1
	}
	c.viewport.Width = width - 2
	c.viewport.Height = vpHeight
	c.msgView.SetWidth(width - 2)
	c.rebuildViewport()
	if c.atBottom {
		c.viewport.GotoBottom()
	}
}

func (c *Chat) rebuildViewport() {
	var sb strings.Builder

	if len(c.messages) == 0 && c.streamingMsg == nil {
		sb.WriteString(RenderWelcome(c.styles, c.viewport.Width))
	} else {
		for _, msg := range c.messages {
			sb.WriteString(c.msgView.Render(msg, false))
			sb.WriteString("\n")
		}
		if c.streamingMsg != nil {
			sb.WriteString(c.msgView.Render(*c.streamingMsg, true))
			sb.WriteString("\n")
		}
	}

	c.viewport.SetContent(sb.String())
}

func (c Chat) View() string {
	var inputStyle lipgloss.Style
	if c.textarea.Focused() {
		inputStyle = c.styles.InputFocused
	} else {
		inputStyle = c.styles.InputUnfocused
	}

	inputBox := inputStyle.Width(c.width - 4).Render(c.textarea.View())

	return lipgloss.JoinVertical(
		lipgloss.Left,
		c.viewport.View(),
		inputBox,
	)
}
