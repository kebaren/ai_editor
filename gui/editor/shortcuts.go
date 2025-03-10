package editor

import (
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

		return false
	})
}
