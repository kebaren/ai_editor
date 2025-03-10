package editor

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/example/gotextbuffer/textbuffer"
)

// EditorWindow represents the main editor window
type EditorWindow struct {
	*gtk.ApplicationWindow
	textView        *gtk.TextView
	bufferAdapter   *BufferAdapter
	textBuffer      *textbuffer.TextBuffer
	filePath        string
	searchBar       *gtk.SearchBar
	searchEntry     *gtk.Entry
	statusBar       *gtk.Statusbar
	contextID       uint
	caseSensitive   bool
	wholeWord       bool
	useRegex        bool
	searchDirection bool // true for forward, false for backward
}

// NewEditorWindow creates a new editor window
func NewEditorWindow(app *gtk.Application) *EditorWindow {
	// Create a new window
	win := gtk.NewApplicationWindow(app)
	win.SetTitle("GoTextBuffer Editor")
	win.SetDefaultSize(800, 600)

	// Create the editor window
	editor := &EditorWindow{
		ApplicationWindow: win,
		textBuffer:        textbuffer.NewTextBuffer(),
		caseSensitive:     false,
		wholeWord:         false,
		useRegex:          false,
		searchDirection:   true,
	}

	// Set up the UI
	editor.setupUI()

	// Set up keyboard shortcuts
	editor.SetupShortcuts()

	// Connect window events
	win.ConnectDestroy(func() {
		editor.textBuffer.Close()
	})

	return editor
}

// setupUI sets up the UI components
func (e *EditorWindow) setupUI() {
	// Create a vertical box to hold the UI components
	vbox := gtk.NewBox(gtk.OrientationVertical, 0)
	e.SetChild(vbox)

	// Create the menu bar
	menuBar := e.createMenuBar()
	vbox.Append(menuBar)

	// Create the toolbar
	toolbar := e.createToolbar()
	vbox.Append(toolbar)

	// Create the search bar
	e.searchBar = gtk.NewSearchBar()
	searchBox := gtk.NewBox(gtk.OrientationHorizontal, 5)
	e.searchEntry = gtk.NewEntry()
	e.searchEntry.SetPlaceholderText("Search text...")
	e.searchEntry.SetHExpand(true)

	// Connect search entry events
	e.searchEntry.ConnectActivate(func() {
		e.search(e.searchDirection)
	})

	// Add search options
	caseSensitiveCheck := gtk.NewCheckButton()
	caseSensitiveCheck.SetLabel("Case sensitive")
	caseSensitiveCheck.ConnectToggled(func() {
		e.caseSensitive = caseSensitiveCheck.Active()
	})

	wholeWordCheck := gtk.NewCheckButton()
	wholeWordCheck.SetLabel("Whole word")
	wholeWordCheck.ConnectToggled(func() {
		e.wholeWord = wholeWordCheck.Active()
	})

	regexCheck := gtk.NewCheckButton()
	regexCheck.SetLabel("Regex")
	regexCheck.ConnectToggled(func() {
		e.useRegex = regexCheck.Active()
	})

	prevButton := gtk.NewButton()
	prevButton.SetIconName("go-up-symbolic")
	prevButton.ConnectClicked(func() {
		e.search(false)
	})

	nextButton := gtk.NewButton()
	nextButton.SetIconName("go-down-symbolic")
	nextButton.ConnectClicked(func() {
		e.search(true)
	})

	closeButton := gtk.NewButton()
	closeButton.SetIconName("window-close-symbolic")
	closeButton.ConnectClicked(func() {
		e.searchBar.SetSearchMode(false)
	})

	searchBox.Append(e.searchEntry)
	searchBox.Append(caseSensitiveCheck)
	searchBox.Append(wholeWordCheck)
	searchBox.Append(regexCheck)
	searchBox.Append(prevButton)
	searchBox.Append(nextButton)
	searchBox.Append(closeButton)

	e.searchBar.SetChild(searchBox)
	vbox.Append(e.searchBar)

	// Create a scrolled window to hold the text view
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetVExpand(true)
	scrolledWindow.SetHExpand(true)
	vbox.Append(scrolledWindow)

	// Create the text view
	e.textView = gtk.NewTextView()
	gtkBuffer := e.textView.Buffer()
	e.textView.SetWrapMode(gtk.WrapNone)
	e.textView.SetMonospace(true)
	scrolledWindow.SetChild(e.textView)

	// Create the buffer adapter
	e.bufferAdapter = NewBufferAdapter(gtkBuffer, e.textBuffer)

	// Create the status bar
	e.statusBar = gtk.NewStatusbar()
	e.contextID = e.statusBar.GetContextID("editor")
	vbox.Append(e.statusBar)

	// Update the status bar
	e.updateStatusBar()

	// Connect cursor position changed signal
	gtkBuffer.ConnectNotify("cursor-position", func() {
		e.updateStatusBar()
	})
}

// createMenuBar creates the menu bar
func (e *EditorWindow) createMenuBar() *gtk.Box {
	box := gtk.NewBox(gtk.OrientationHorizontal, 0)

	// File menu
	fileMenu := gtk.NewPopoverMenu()
	fileMenu.SetHasArrow(false)

	fileMenuButton := gtk.NewMenuButton()
	fileMenuButton.SetLabel("File")
	fileMenuButton.SetPopover(fileMenu)

	fileMenuModel := gio.NewMenu()

	newAction := gio.NewSimpleAction("new", nil)
	newAction.ConnectActivate(func(parameter *gio.Variant) {
		e.newFile()
	})
	e.ApplicationWindow.AddAction(newAction)
	fileMenuModel.Append("New", "win.new")

	openAction := gio.NewSimpleAction("open", nil)
	openAction.ConnectActivate(func(parameter *gio.Variant) {
		e.openFile()
	})
	e.ApplicationWindow.AddAction(openAction)
	fileMenuModel.Append("Open", "win.open")

	saveAction := gio.NewSimpleAction("save", nil)
	saveAction.ConnectActivate(func(parameter *gio.Variant) {
		e.saveFile()
	})
	e.ApplicationWindow.AddAction(saveAction)
	fileMenuModel.Append("Save", "win.save")

	saveAsAction := gio.NewSimpleAction("save-as", nil)
	saveAsAction.ConnectActivate(func(parameter *gio.Variant) {
		e.saveFileAs()
	})
	e.ApplicationWindow.AddAction(saveAsAction)
	fileMenuModel.Append("Save As", "win.save-as")

	fileMenuModel.Append("Quit", "app.quit")

	fileMenu.SetMenuModel(fileMenuModel)

	// Edit menu
	editMenu := gtk.NewPopoverMenu()
	editMenu.SetHasArrow(false)

	editMenuButton := gtk.NewMenuButton()
	editMenuButton.SetLabel("Edit")
	editMenuButton.SetPopover(editMenu)

	editMenuModel := gio.NewMenu()

	undoAction := gio.NewSimpleAction("undo", nil)
	undoAction.ConnectActivate(func(parameter *gio.Variant) {
		e.undo()
	})
	e.ApplicationWindow.AddAction(undoAction)
	editMenuModel.Append("Undo", "win.undo")

	redoAction := gio.NewSimpleAction("redo", nil)
	redoAction.ConnectActivate(func(parameter *gio.Variant) {
		e.redo()
	})
	e.ApplicationWindow.AddAction(redoAction)
	editMenuModel.Append("Redo", "win.redo")

	editMenuModel.Append("Cut", "win.cut")
	editMenuModel.Append("Copy", "win.copy")
	editMenuModel.Append("Paste", "win.paste")

	findAction := gio.NewSimpleAction("find", nil)
	findAction.ConnectActivate(func(parameter *gio.Variant) {
		e.showSearchBar()
	})
	e.ApplicationWindow.AddAction(findAction)
	editMenuModel.Append("Find", "win.find")

	replaceAction := gio.NewSimpleAction("replace", nil)
	replaceAction.ConnectActivate(func(parameter *gio.Variant) {
		e.showReplaceDialog()
	})
	e.ApplicationWindow.AddAction(replaceAction)
	editMenuModel.Append("Replace", "win.replace")

	editMenu.SetMenuModel(editMenuModel)

	// Add standard clipboard actions
	cutAction := gio.NewSimpleAction("cut", nil)
	cutAction.ConnectActivate(func(parameter *gio.Variant) {
		e.textView.EmitSignal("cut-clipboard", nil)
	})
	e.ApplicationWindow.AddAction(cutAction)

	copyAction := gio.NewSimpleAction("copy", nil)
	copyAction.ConnectActivate(func(parameter *gio.Variant) {
		e.textView.EmitSignal("copy-clipboard", nil)
	})
	e.ApplicationWindow.AddAction(copyAction)

	pasteAction := gio.NewSimpleAction("paste", nil)
	pasteAction.ConnectActivate(func(parameter *gio.Variant) {
		e.textView.EmitSignal("paste-clipboard", nil)
	})
	e.ApplicationWindow.AddAction(pasteAction)

	// Add menu buttons to the box
	box.Append(fileMenuButton)
	box.Append(editMenuButton)

	return box
}

// createToolbar creates the toolbar
func (e *EditorWindow) createToolbar() *gtk.Box {
	toolbar := gtk.NewBox(gtk.OrientationHorizontal, 5)
	toolbar.SetMarginStart(5)
	toolbar.SetMarginEnd(5)
	toolbar.SetMarginTop(5)
	toolbar.SetMarginBottom(5)

	// New button
	newButton := gtk.NewButton()
	newButton.SetIconName("document-new")
	newButton.SetTooltipText("New File")
	newButton.ConnectClicked(func() {
		e.newFile()
	})
	toolbar.Append(newButton)

	// Open button
	openButton := gtk.NewButton()
	openButton.SetIconName("document-open")
	openButton.SetTooltipText("Open File")
	openButton.ConnectClicked(func() {
		e.openFile()
	})
	toolbar.Append(openButton)

	// Save button
	saveButton := gtk.NewButton()
	saveButton.SetIconName("document-save")
	saveButton.SetTooltipText("Save File")
	saveButton.ConnectClicked(func() {
		e.saveFile()
	})
	toolbar.Append(saveButton)

	// Separator
	separator1 := gtk.NewSeparator(gtk.OrientationVertical)
	toolbar.Append(separator1)

	// Undo button
	undoButton := gtk.NewButton()
	undoButton.SetIconName("edit-undo")
	undoButton.SetTooltipText("Undo")
	undoButton.ConnectClicked(func() {
		e.undo()
	})
	toolbar.Append(undoButton)

	// Redo button
	redoButton := gtk.NewButton()
	redoButton.SetIconName("edit-redo")
	redoButton.SetTooltipText("Redo")
	redoButton.ConnectClicked(func() {
		e.redo()
	})
	toolbar.Append(redoButton)

	// Separator
	separator2 := gtk.NewSeparator(gtk.OrientationVertical)
	toolbar.Append(separator2)

	// Find button
	findButton := gtk.NewButton()
	findButton.SetIconName("edit-find")
	findButton.SetTooltipText("Find")
	findButton.ConnectClicked(func() {
		e.showSearchBar()
	})
	toolbar.Append(findButton)

	// Replace button
	replaceButton := gtk.NewButton()
	replaceButton.SetIconName("edit-find-replace")
	replaceButton.SetTooltipText("Replace")
	replaceButton.ConnectClicked(func() {
		e.showReplaceDialog()
	})
	toolbar.Append(replaceButton)

	return toolbar
}

// newFile creates a new empty file
func (e *EditorWindow) newFile() {
	// Check if the current file has been modified
	if e.textBuffer.IsModified() {
		dialog := gtk.NewMessageDialog(
			e.ApplicationWindow,
			gtk.DialogFlagsModal,
			gtk.MessageQuestion,
			gtk.ButtonsYesNo,
			"The current file has been modified. Do you want to save it?",
		)
		dialog.SetTitle("Save Changes")

		response := dialog.Run()
		dialog.Destroy()

		if response == gtk.ResponseYes {
			e.saveFile()
		}
	}

	// Clear the buffer
	e.textBuffer.Clear()
	e.bufferAdapter.updateGtkBuffer()

	// Reset the file path
	e.filePath = ""
	e.SetTitle("GoTextBuffer Editor - [New File]")

	// Update the status bar
	e.updateStatusBar()
}

// openFile opens a file
func (e *EditorWindow) openFile() {
	// Check if the current file has been modified
	if e.textBuffer.IsModified() {
		dialog := gtk.NewMessageDialog(
			e.ApplicationWindow,
			gtk.DialogFlagsModal,
			gtk.MessageQuestion,
			gtk.ButtonsYesNo,
			"The current file has been modified. Do you want to save it?",
		)
		dialog.SetTitle("Save Changes")

		response := dialog.Run()
		dialog.Destroy()

		if response == gtk.ResponseYes {
			e.saveFile()
		}
	}

	// Create a file chooser dialog
	dialog := gtk.NewFileChooserDialog(
		"Open File",
		e.ApplicationWindow,
		gtk.FileChooserActionOpen,
		"_Cancel", gtk.ResponseCancel,
		"_Open", gtk.ResponseAccept,
	)

	// Add filters
	filter := gtk.NewFileFilter()
	filter.SetName("Text Files")
	filter.AddPattern("*.txt")
	dialog.AddFilter(filter)

	filter = gtk.NewFileFilter()
	filter.SetName("All Files")
	filter.AddPattern("*")
	dialog.AddFilter(filter)

	// Run the dialog
	response := dialog.Run()

	if response == gtk.ResponseAccept {
		// Get the selected file
		filename := dialog.Filename()

		// Load the file
		err := e.textBuffer.LoadFromFile(filename)
		if err != nil {
			errorDialog := gtk.NewMessageDialog(
				e.ApplicationWindow,
				gtk.DialogFlagsModal,
				gtk.MessageError,
				gtk.ButtonsOk,
				"Error opening file: %s", err.Error(),
			)
			errorDialog.SetTitle("Error")
			errorDialog.Run()
			errorDialog.Destroy()
		} else {
			// Update the GTK buffer
			e.bufferAdapter.updateGtkBuffer()

			// Update the file path
			e.filePath = filename
			e.SetTitle(fmt.Sprintf("GoTextBuffer Editor - [%s]", filepath.Base(filename)))

			// Update the status bar
			e.updateStatusBar()
		}
	}

	dialog.Destroy()
}

// saveFile saves the current file
func (e *EditorWindow) saveFile() {
	// If the file path is empty, use saveFileAs
	if e.filePath == "" {
		e.saveFileAs()
		return
	}

	// Save the file
	err := e.textBuffer.SaveToFile(e.filePath)
	if err != nil {
		dialog := gtk.NewMessageDialog(
			e.ApplicationWindow,
			gtk.DialogFlagsModal,
			gtk.MessageError,
			gtk.ButtonsOk,
			"Error saving file: %s", err.Error(),
		)
		dialog.SetTitle("Error")
		dialog.Run()
		dialog.Destroy()
	} else {
		// Update the status bar
		e.updateStatusBar()
	}
}

// saveFileAs saves the current file with a new name
func (e *EditorWindow) saveFileAs() {
	// Create a file chooser dialog
	dialog := gtk.NewFileChooserDialog(
		"Save File As",
		e.ApplicationWindow,
		gtk.FileChooserActionSave,
		"_Cancel", gtk.ResponseCancel,
		"_Save", gtk.ResponseAccept,
	)

	// Add filters
	filter := gtk.NewFileFilter()
	filter.SetName("Text Files")
	filter.AddPattern("*.txt")
	dialog.AddFilter(filter)

	filter = gtk.NewFileFilter()
	filter.SetName("All Files")
	filter.AddPattern("*")
	dialog.AddFilter(filter)

	// Set the current file name if available
	if e.filePath != "" {
		dialog.SetCurrentName(filepath.Base(e.filePath))
	} else {
		dialog.SetCurrentName("untitled.txt")
	}

	// Run the dialog
	response := dialog.Run()

	if response == gtk.ResponseAccept {
		// Get the selected file
		filename := dialog.Filename()

		// Save the file
		err := e.textBuffer.SaveToFile(filename)
		if err != nil {
			errorDialog := gtk.NewMessageDialog(
				e.ApplicationWindow,
				gtk.DialogFlagsModal,
				gtk.MessageError,
				gtk.ButtonsOk,
				"Error saving file: %s", err.Error(),
			)
			errorDialog.SetTitle("Error")
			errorDialog.Run()
			errorDialog.Destroy()
		} else {
			// Update the file path
			e.filePath = filename
			e.SetTitle(fmt.Sprintf("GoTextBuffer Editor - [%s]", filepath.Base(filename)))

			// Update the status bar
			e.updateStatusBar()
		}
	}

	dialog.Destroy()
}

// undo undoes the last operation
func (e *EditorWindow) undo() {
	err := e.textBuffer.Undo()
	if err != nil {
		log.Printf("Undo error: %s", err.Error())
	} else {
		e.bufferAdapter.updateGtkBuffer()
	}
}

// redo redoes the last undone operation
func (e *EditorWindow) redo() {
	err := e.textBuffer.Redo()
	if err != nil {
		log.Printf("Redo error: %s", err.Error())
	} else {
		e.bufferAdapter.updateGtkBuffer()
	}
}

// showSearchBar shows the search bar
func (e *EditorWindow) showSearchBar() {
	e.searchBar.SetSearchMode(true)
	e.searchEntry.GrabFocus()

	// If there's selected text, use it as the search text
	gtkBuffer := e.bufferAdapter.GetGtkBuffer()
	if gtkBuffer.HasSelection() {
		start, end := gtkBuffer.SelectionBounds()
		text := gtkBuffer.Text(start, end, false)
		e.searchEntry.SetText(text)
	}
}

// search performs a search operation
func (e *EditorWindow) search(forward bool) {
	searchText := e.searchEntry.Text()
	if searchText == "" {
		return
	}

	e.searchDirection = forward

	// Get the current cursor position
	cursorPos := e.bufferAdapter.GetCursorPosition()

	// Perform the search
	var searchRange textbuffer.Range
	var err error

	if forward {
		searchRange, err = e.textBuffer.FindNext(searchText, cursorPos, e.caseSensitive, e.wholeWord, e.useRegex)
	} else {
		searchRange, err = e.textBuffer.FindPrevious(searchText, cursorPos, e.caseSensitive, e.wholeWord, e.useRegex)
	}

	if err != nil {
		e.statusBar.Push(e.contextID, fmt.Sprintf("Search failed: %s", err.Error()))
		return
	}

	// Select the found text
	e.bufferAdapter.SelectRange(searchRange)

	// Scroll to the selection
	gtkBuffer := e.bufferAdapter.GetGtkBuffer()
	startIter := gtkBuffer.GetIterAtLineOffset(searchRange.Start.Line, searchRange.Start.Column)
	e.textView.ScrollToIter(startIter, 0.1, false, 0, 0)

	e.statusBar.Push(e.contextID, "Text found")
}

// showReplaceDialog shows the replace dialog
func (e *EditorWindow) showReplaceDialog() {
	// Create a dialog
	dialog := gtk.NewDialog()
	dialog.SetTitle("Replace")
	dialog.SetTransientFor(e.ApplicationWindow)
	dialog.SetModal(true)
	dialog.SetDefaultSize(400, 200)

	// Get the content area
	contentArea := dialog.ContentArea()

	// Create a grid for the dialog content
	grid := gtk.NewGrid()
	grid.SetRowSpacing(10)
	grid.SetColumnSpacing(10)
	grid.SetMarginStart(10)
	grid.SetMarginEnd(10)
	grid.SetMarginTop(10)
	grid.SetMarginBottom(10)
	contentArea.Append(grid)

	// Find label and entry
	findLabel := gtk.NewLabel("Find:")
	findLabel.SetHAlign(gtk.AlignEnd)
	grid.Attach(findLabel, 0, 0, 1, 1)

	findEntry := gtk.NewEntry()
	findEntry.SetHExpand(true)
	grid.Attach(findEntry, 1, 0, 1, 1)

	// Replace label and entry
	replaceLabel := gtk.NewLabel("Replace with:")
	replaceLabel.SetHAlign(gtk.AlignEnd)
	grid.Attach(replaceLabel, 0, 1, 1, 1)

	replaceEntry := gtk.NewEntry()
	replaceEntry.SetHExpand(true)
	grid.Attach(replaceEntry, 1, 1, 1, 1)

	// Options
	optionsFrame := gtk.NewFrame("Options")
	optionsBox := gtk.NewBox(gtk.OrientationVertical, 5)
	optionsFrame.SetChild(optionsBox)
	grid.Attach(optionsFrame, 0, 2, 2, 1)

	caseSensitiveCheck := gtk.NewCheckButton()
	caseSensitiveCheck.SetLabel("Case sensitive")
	caseSensitiveCheck.SetActive(e.caseSensitive)
	optionsBox.Append(caseSensitiveCheck)

	wholeWordCheck := gtk.NewCheckButton()
	wholeWordCheck.SetLabel("Whole word")
	wholeWordCheck.SetActive(e.wholeWord)
	optionsBox.Append(wholeWordCheck)

	regexCheck := gtk.NewCheckButton()
	regexCheck.SetLabel("Regular expression")
	regexCheck.SetActive(e.useRegex)
	optionsBox.Append(regexCheck)

	// Add buttons
	dialog.AddButton("_Cancel", gtk.ResponseCancel)
	dialog.AddButton("Replace _All", gtk.ResponseYes)
	dialog.AddButton("_Replace", gtk.ResponseAccept)

	// If there's selected text, use it as the find text
	if e.bufferAdapter.GetGtkBuffer().HasSelection() {
		start, end := e.bufferAdapter.GetGtkBuffer().SelectionBounds()
		text := e.bufferAdapter.GetGtkBuffer().Text(start, end, false)
		findEntry.SetText(text)
	}

	// Show the dialog
	dialog.Show()

	// Connect the response signal
	dialog.ConnectResponse(func(responseId int) {
		if responseId == gtk.ResponseAccept || responseId == gtk.ResponseYes {
			findText := findEntry.Text()
			replaceText := replaceEntry.Text()

			if findText == "" {
				return
			}

			// Update search options
			e.caseSensitive = caseSensitiveCheck.Active()
			e.wholeWord = wholeWordCheck.Active()
			e.useRegex = regexCheck.Active()

			if responseId == gtk.ResponseYes {
				// Replace all
				count, err := e.textBuffer.ReplaceAll(findText, replaceText, e.caseSensitive, e.wholeWord, e.useRegex)
				if err != nil {
					e.statusBar.Push(e.contextID, fmt.Sprintf("Replace failed: %s", err.Error()))
				} else {
					e.bufferAdapter.updateGtkBuffer()
					e.statusBar.Push(e.contextID, fmt.Sprintf("Replaced %d occurrences", count))
				}
			} else {
				// Replace current selection
				if e.bufferAdapter.GetGtkBuffer().HasSelection() {
					start, end := e.bufferAdapter.GetGtkBuffer().SelectionBounds()
					selectedText := e.bufferAdapter.GetGtkBuffer().Text(start, end, false)

					// Check if the selection matches the find text
					matches := false
					if e.caseSensitive {
						matches = selectedText == findText
					} else {
						matches = strings.ToLower(selectedText) == strings.ToLower(findText)
					}

					if matches {
						// Replace the selection
						e.bufferAdapter.GetGtkBuffer().Delete(start, end)
						e.bufferAdapter.GetGtkBuffer().Insert(start, replaceText)

						// Update the text buffer
						text := e.bufferAdapter.GetGtkBuffer().Text(e.bufferAdapter.GetGtkBuffer().StartIter(), e.bufferAdapter.GetGtkBuffer().EndIter(), false)
						e.textBuffer.SetText(text)

						// Search for the next occurrence
						e.searchEntry.SetText(findText)
						e.search(true)
					} else {
						// Search for the first occurrence
						e.searchEntry.SetText(findText)
						e.search(true)
					}
				} else {
					// Search for the first occurrence
					e.searchEntry.SetText(findText)
					e.search(true)
				}
			}
		}

		dialog.Destroy()
	})
}

// updateGtkBuffer updates the GTK buffer from the text buffer
func (e *EditorWindow) updateGtkBuffer() {
	e.bufferAdapter.updateGtkBuffer()
}

// updateStatusBar updates the status bar
func (e *EditorWindow) updateStatusBar() {
	// Get the cursor position
	cursorPos := e.bufferAdapter.GetCursorPosition()
	line := cursorPos.Line + 1
	column := cursorPos.Column + 1

	// Get the file status
	status := "Modified"
	if !e.textBuffer.IsModified() {
		status = "Saved"
	}

	// Get the file path
	path := e.filePath
	if path == "" {
		path = "New File"
	}

	// Update the status bar
	e.statusBar.Push(e.contextID, fmt.Sprintf("Line: %d, Column: %d | %s | %s", line, column, status, path))
}
