package editor

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/example/gotextbuffer/textbuffer"
)

// BufferAdapter handles the integration between GTK's text buffer and our custom textbuffer
type BufferAdapter struct {
	gtkBuffer  *gtk.TextBuffer
	textBuffer *textbuffer.TextBuffer
	updating   bool
}

// NewBufferAdapter creates a new buffer adapter
func NewBufferAdapter(gtkBuffer *gtk.TextBuffer, textBuffer *textbuffer.TextBuffer) *BufferAdapter {
	adapter := &BufferAdapter{
		gtkBuffer:  gtkBuffer,
		textBuffer: textBuffer,
		updating:   false,
	}

	// Connect the GTK buffer's changed signal
	gtkBuffer.ConnectChanged(func() {
		if !adapter.updating {
			adapter.onGtkBufferChanged()
		}
	})

	// Initialize the GTK buffer with the text buffer's content
	adapter.updateGtkBuffer()

	return adapter
}

// onGtkBufferChanged is called when the GTK buffer changes
func (a *BufferAdapter) onGtkBufferChanged() {
	// Get the text from the GTK buffer
	text := a.gtkBuffer.Text(a.gtkBuffer.StartIter(), a.gtkBuffer.EndIter(), false)

	// Update the text buffer
	a.textBuffer.SetText(text)
}

// updateGtkBuffer updates the GTK buffer from the text buffer
func (a *BufferAdapter) updateGtkBuffer() {
	a.updating = true
	text := a.textBuffer.GetText()
	a.gtkBuffer.SetText(text)
	a.updating = false
}

// GetTextBuffer returns the text buffer
func (a *BufferAdapter) GetTextBuffer() *textbuffer.TextBuffer {
	return a.textBuffer
}

// GetGtkBuffer returns the GTK buffer
func (a *BufferAdapter) GetGtkBuffer() *gtk.TextBuffer {
	return a.gtkBuffer
}

// SetCursorPosition sets the cursor position in the GTK buffer
func (a *BufferAdapter) SetCursorPosition(position textbuffer.Position) {
	iter := a.gtkBuffer.GetIterAtLineOffset(position.Line, position.Column)
	a.gtkBuffer.PlaceCursor(iter)
}

// GetCursorPosition gets the cursor position from the GTK buffer
func (a *BufferAdapter) GetCursorPosition() textbuffer.Position {
	mark := a.gtkBuffer.GetMark("insert")
	iter := a.gtkBuffer.GetIterAtMark(mark)
	return textbuffer.Position{
		Line:   iter.Line(),
		Column: iter.LineOffset(),
	}
}

// SelectRange selects a range in the GTK buffer
func (a *BufferAdapter) SelectRange(r textbuffer.Range) {
	startIter := a.gtkBuffer.GetIterAtLineOffset(r.Start.Line, r.Start.Column)
	endIter := a.gtkBuffer.GetIterAtLineOffset(r.End.Line, r.End.Column)
	a.gtkBuffer.SelectRange(startIter, endIter)
}

// GetSelection gets the current selection from the GTK buffer
func (a *BufferAdapter) GetSelection() (textbuffer.Range, bool) {
	if a.gtkBuffer.HasSelection() {
		start, end := a.gtkBuffer.SelectionBounds()
		return textbuffer.Range{
			Start: textbuffer.Position{
				Line:   start.Line(),
				Column: start.LineOffset(),
			},
			End: textbuffer.Position{
				Line:   end.Line(),
				Column: end.LineOffset(),
			},
		}, true
	}
	return textbuffer.Range{}, false
}
