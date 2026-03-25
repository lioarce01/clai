package markdown

import (
	"sync"

	"github.com/charmbracelet/glamour"
)

// Renderer wraps glamour to convert Markdown into styled terminal output.
type Renderer struct {
	mu       sync.Mutex
	renderer *glamour.TermRenderer
	width    int
}

// New creates a Renderer configured for the given terminal width.
func New(width int) (*Renderer, error) {
	r := &Renderer{width: width}
	if err := r.rebuild(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Renderer) rebuild() error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(r.width),
		glamour.WithEmoji(),
	)
	if err != nil {
		return err
	}
	r.renderer = renderer
	return nil
}

// Render converts markdown text to ANSI-styled terminal output.
func (r *Renderer) Render(markdown string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.renderer.Render(markdown)
}

// SetWidth updates the word-wrap width and rebuilds the renderer.
func (r *Renderer) SetWidth(width int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.width == width {
		return
	}
	r.width = width
	_ = r.rebuild() // best effort; old renderer remains on error
}
