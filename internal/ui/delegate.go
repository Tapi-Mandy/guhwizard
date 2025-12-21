// FILE: internal/ui/delegate.go
package ui

import (
    "fmt"
    "io"
    "guhwizard/internal/styles"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbletea"
)

// CustomDelegate handles the rendering of list items
type CustomDelegate struct{}

func (d CustomDelegate) Height() int { return 2 }
func (d CustomDelegate) Spacing() int { return 1 }
func (d CustomDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d CustomDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    i, ok := item.(listItem)
    if !ok { return }

    title := i.Title() // Already contains [x] or [ ]
    desc := i.Description()

    if index == m.Index() {
        // --- SELECTED RENDER (The one with the colored bar) ---
        fmt.Fprintln(w, styles.ItemSelectedTitle.Render(title))
        fmt.Fprint(w, styles.ItemSelectedDesc.Render(desc))
    } else {
        // --- NORMAL RENDER ---
        fmt.Fprintln(w, styles.ItemNormalTitle.Render(title))
        fmt.Fprint(w, styles.ItemNormalDesc.Render(desc))
    }
}
