package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lioarce01/clai/internal/storage"
	"github.com/rivo/tview"
)

// buildSessionModal constructs and returns a tview modal containing a session
// list. The caller is responsible for adding/removing it from the pages.
//
// onSelect is called with the selected session ID when the user presses Enter.
// onDelete is called with the session ID when the user presses 'd'.
// onNew is called when the user presses 'n'.
// onClose is called when the user presses Esc.
func buildSessionModal(
	sessions []storage.Session,
	activeID string,
	onSelect func(id string),
	onDelete func(id string),
	onNew func(),
	onClose func(),
) *tview.Frame {
	list := tview.NewList()
	list.SetBorder(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.SetSelectedTextColor(tcell.ColorDefault)
	list.SetSelectedBackgroundColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)

	for _, s := range sessions {
		sess := s // capture loop variable
		count := len(sess.Messages)
		subtitle := fmt.Sprintf("  %d messages · %s", count, sess.UpdatedAt.Format("Jan 2, 15:04"))
		prefix := "  "
		if sess.ID == activeID {
			prefix = "> "
		}
		title := prefix + sess.Name
		list.AddItem(title, subtitle, 0, func() {
			onSelect(sess.ID)
		})
	}

	list.SetSelectedFunc(func(_ int, _ string, _ string, _ rune) {
		// handled per-item above
	})

	// Build hint line
	hint := tview.NewTextView()
	hint.SetDynamicColors(true)
	hint.SetBackgroundColor(tcell.ColorDefault)
	hint.SetTextColor(tcell.ColorDefault)
	hint.SetText(Dim + "[n]ew  [d]elete  [enter]select  [esc]close" + Reset)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true).
		AddItem(hint, 1, 0, false)

	frame := tview.NewFrame(flex).
		SetBorders(1, 1, 1, 1, 2, 2)
	frame.SetTitle(" Sessions ")
	frame.SetBorder(true)
	frame.SetBackgroundColor(tcell.ColorDefault)

	// Key handling on the list
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			onClose()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'n':
				onNew()
				return nil
			case 'd':
				idx := list.GetCurrentItem()
				if idx >= 0 && idx < len(sessions) {
					onDelete(sessions[idx].ID)
				}
				return nil
			}
		}
		return event
	})

	return frame
}
