package tui

import (
	"github.com/charmbracelet/glamour"
)

var mdRenderer *glamour.TermRenderer

func initMarkdown(width int) {
	if width < 20 {
		width = 80
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return
	}
	mdRenderer = r
}

// renderMarkdown renders markdown text using glamour.
// Falls back to raw text on error.
func renderMarkdown(text string) string {
	if mdRenderer == nil {
		initMarkdown(80)
	}
	if mdRenderer == nil {
		return text
	}
	out, err := mdRenderer.Render(text)
	if err != nil {
		return text
	}
	return out
}
