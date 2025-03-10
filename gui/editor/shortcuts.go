package editor

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// SetupShortcuts sets up keyboard shortcuts for the editor window
func (e *EditorWindow) SetupShortcuts() {
	// Create a new event controller for key events
	keyController := gtk.NewEventControllerKey()
	e.AddController(keyController)

	// Connect the key-pressed signal
	keyController.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		// Check for Ctrl+S (Save)
		if keyval == gdk.KEY_s && state&gdk.ControlMask != 0 {
			e.saveFile()
			return true
		}

		// Check for Ctrl+Shift+S (Save As)
		if keyval == gdk.KEY_S && state&gdk.ControlMask != 0 && state&gdk.ShiftMask != 0 {
			e.saveFileAs()
			return true
		}

		// Check for Ctrl+O (Open)
		if keyval == gdk.KEY_o && state&gdk.ControlMask != 0 {
			e.openFile()
			return true
		}

		// Check for Ctrl+N (New)
		if keyval == gdk.KEY_n && state&gdk.ControlMask != 0 {
			e.newFile()
			return true
		}

		// Check for Ctrl+Z (Undo)
		if keyval == gdk.KEY_z && state&gdk.ControlMask != 0 {
			e.undo()
			return true
		}

		// Check for Ctrl+Y (Redo)
		if keyval == gdk.KEY_y && state&gdk.ControlMask != 0 {
			e.redo()
			return true
		}

		// Check for Ctrl+F (Find)
		if keyval == gdk.KEY_f && state&gdk.ControlMask != 0 {
			e.showSearchBar()
			return true
		}

		// Check for Ctrl+H (Replace)
		if keyval == gdk.KEY_h && state&gdk.ControlMask != 0 {
			e.showReplaceDialog()
			return true
		}

		// Check for F3 (Find Next)
		if keyval == gdk.KEY_F3 {
			e.search(true)
			return true
		}

		// Check for Shift+F3 (Find Previous)
		if keyval == gdk.KEY_F3 && state&gdk.ShiftMask != 0 {
			e.search(false)
			return true
		}

		// Check for Escape (Close search bar)
		if keyval == gdk.KEY_Escape && e.searchBar.SearchMode() {
			e.searchBar.SetSearchMode(false)
			return true
		}

		// Check for Ctrl+D (Duplicate Line)
		if keyval == gdk.KEY_d && state&gdk.ControlMask != 0 {
			e.duplicateLine()
			return true
		}

		// Check for Ctrl+Shift+K (Delete Line)
		if keyval == gdk.KEY_k && state&gdk.ControlMask != 0 && state&gdk.ShiftMask != 0 {
			e.deleteLine()
			return true
		}

		// Check for Alt+Up (Move Line Up)
		if keyval == gdk.KEY_Up && state&gdk.AltMask != 0 {
			e.moveLineUp()
			return true
		}

		// Check for Alt+Down (Move Line Down)
		if keyval == gdk.KEY_Down && state&gdk.AltMask != 0 {
			e.moveLineDown()
			return true
		}

		// Check for Tab (Increase Indent)
		buffer := e.bufferAdapter.GetGtkBuffer()
		if keyval == gdk.KEY_Tab && buffer.HasSelection() {
			e.increaseIndent()
			return true
		}

		// Check for Shift+Tab (Decrease Indent)
		if keyval == gdk.KEY_Tab && state&gdk.ShiftMask != 0 && buffer.HasSelection() {
			e.decreaseIndent()
			return true
		}

		// Check for Ctrl+/ (Toggle Line Comment)
		if keyval == gdk.KEY_slash && state&gdk.ControlMask != 0 {
			// 如果当前行已有注释，则移除注释，否则添加注释
			cursor := e.bufferAdapter.GetGtkBuffer().GetInsert()
			iter := e.bufferAdapter.GetGtkBuffer().IterAtMark(cursor)
			line := iter.Line()

			// 获取行首位置
			lineIter := e.bufferAdapter.GetGtkBuffer().StartIter()
			for i := 0; i < line; i++ {
				lineIter.ForwardLine()
			}

			// 检查行首是否有注释
			lineEndIter := lineIter.Copy()
			lineEndIter.ForwardToLineEnd()
			lineText := e.bufferAdapter.GetGtkBuffer().Text(lineIter, lineEndIter, false)

			if strings.HasPrefix(strings.TrimSpace(lineText), "//") {
				e.removeLineComment()
			} else {
				e.addLineComment()
			}
			return true
		}

		// Check for Ctrl+Shift+/ (Add Block Comment)
		if keyval == gdk.KEY_slash && state&gdk.ControlMask != 0 && state&gdk.ShiftMask != 0 {
			e.addBlockComment()
			return true
		}

		// Check for Ctrl+U (Transform to Uppercase)
		if keyval == gdk.KEY_u && state&gdk.ControlMask != 0 {
			e.transformToUpper()
			return true
		}

		// Check for Ctrl+L (Transform to Lowercase)
		if keyval == gdk.KEY_l && state&gdk.ControlMask != 0 {
			e.transformToLower()
			return true
		}

		// Check for Ctrl+T (Transform to Title Case)
		if keyval == gdk.KEY_t && state&gdk.ControlMask != 0 {
			e.transformToTitle()
			return true
		}

		return false
	})
}
