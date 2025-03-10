package editor

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
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
	autoSaveEnabled bool
	autoSaveTimer   *time.Timer
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
		autoSaveEnabled:   false,
	}

	// Set up the UI
	editor.setupUI()

	// Set up keyboard shortcuts
	editor.SetupShortcuts()

	// Connect window events
	win.ConnectDestroy(func() {
		if editor.autoSaveTimer != nil {
			editor.autoSaveTimer.Stop()
		}
		editor.textBuffer.Close()
	})

	return editor
}

// setupUI sets up the UI components
func (e *EditorWindow) setupUI() {
	// Set window style
	e.SetDecorated(true)
	e.SetResizable(true)
	e.SetIconName("accessories-text-editor")

	// Create a vertical box to hold the UI components
	vbox := gtk.NewBox(gtk.OrientationVertical, 0)
	e.SetChild(vbox)

	// Create the menu bar
	menuBar := e.createMenuBar()
	menuBar.AddCSSClass("menu-bar")
	vbox.Append(menuBar)

	// Create the toolbar
	toolbar := e.createToolbar()
	toolbar.AddCSSClass("toolbar")
	vbox.Append(toolbar)

	// Create the search bar
	e.searchBar = gtk.NewSearchBar()
	searchBox := gtk.NewBox(gtk.OrientationHorizontal, 5)
	searchBox.SetMarginStart(5)
	searchBox.SetMarginEnd(5)
	searchBox.SetMarginTop(5)
	searchBox.SetMarginBottom(5)

	e.searchEntry = gtk.NewEntry()
	e.searchEntry.SetPlaceholderText("搜索文本...")
	e.searchEntry.SetHExpand(true)
	e.searchEntry.AddCSSClass("search-entry")

	// Connect search entry events
	e.searchEntry.ConnectActivate(func() {
		e.search(e.searchDirection)
	})

	// Add search options
	caseSensitiveCheck := gtk.NewCheckButton()
	caseSensitiveCheck.SetLabel("区分大小写")
	caseSensitiveCheck.ConnectToggled(func() {
		e.caseSensitive = caseSensitiveCheck.Active()
	})

	wholeWordCheck := gtk.NewCheckButton()
	wholeWordCheck.SetLabel("全字匹配")
	wholeWordCheck.ConnectToggled(func() {
		e.wholeWord = wholeWordCheck.Active()
	})

	regexCheck := gtk.NewCheckButton()
	regexCheck.SetLabel("正则表达式")
	regexCheck.ConnectToggled(func() {
		e.useRegex = regexCheck.Active()
	})

	prevButton := gtk.NewButton()
	prevButton.SetIconName("go-up-symbolic")
	prevButton.SetTooltipText("查找上一个")
	prevButton.ConnectClicked(func() {
		e.search(false)
	})

	nextButton := gtk.NewButton()
	nextButton.SetIconName("go-down-symbolic")
	nextButton.SetTooltipText("查找下一个")
	nextButton.ConnectClicked(func() {
		e.search(true)
	})

	closeButton := gtk.NewButton()
	closeButton.SetIconName("window-close-symbolic")
	closeButton.SetTooltipText("关闭搜索栏")
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
	scrolledWindow.AddCSSClass("editor-scrolled-window")
	vbox.Append(scrolledWindow)

	// Create the text view
	e.textView = gtk.NewTextView()
	e.textView.SetWrapMode(gtk.WrapNone)
	e.textView.SetMonospace(true)
	e.textView.SetLeftMargin(5)
	e.textView.SetRightMargin(5)
	e.textView.SetTopMargin(5)
	e.textView.SetBottomMargin(5)
	e.textView.AddCSSClass("editor-text-view")
	scrolledWindow.SetChild(e.textView)

	// Create the buffer adapter
	gtkBuffer := e.textView.Buffer()
	e.bufferAdapter = NewBufferAdapter(gtkBuffer, e.textBuffer)

	// Create the status bar
	e.statusBar = gtk.NewStatusbar()
	e.statusBar.AddCSSClass("status-bar")
	e.contextID = e.statusBar.ContextID("editor")
	vbox.Append(e.statusBar)

	// Update the status bar
	e.updateStatusBar()

	// Connect cursor position changed signal
	gtkBuffer.Connect("notify::cursor-position", func() {
		e.updateStatusBar()
	})

	// Add CSS styling
	provider := gtk.NewCSSProvider()
	provider.LoadFromData(`
		window {
			background-color:rgb(148, 148, 148);
		}
		.menu-bar {
			background-color:rgb(248, 204, 204);
			border-bottom: 1px solid,rgb(197, 231, 212);
			padding: 2px;
		}
		.toolbar {
			background-color: #f0f0f0;
			border-bottom: 1px solid #d0d0d0;
			padding: 2px;
		}
		.editor-text-view {
			font-family: "Consolas", "Courier New", monospace;
			font-size: 12pt;
			background-color: white;
			color: #000000;
		}
		.editor-scrolled-window {
			background-color: white;
			border: 1px solid #d0d0d0;
			margin: 5px;
		}
		.status-bar {
			font-size: 10pt;
			background-color: #f0f0f0;
			border-top: 1px solid #d0d0d0;
			padding: 2px;
		}
		.search-entry {
			min-width: 200px;
		}
		button {
			background-color: #f0f0f0;
			border: 1px solid #d0d0d0;
			border-radius: 2px;
			padding: 4px 8px;
		}
		button:hover {
			background-color: #e0e0e0;
		}
		button:active {
			background-color: #d0d0d0;
		}
		checkbutton {
			margin: 4px;
		}
		entry {
			background-color: white;
			border: 1px solid #d0d0d0;
			border-radius: 2px;
			padding: 4px;
		}
		entry:focus {
			border-color: #0078d7;
		}
	`)
	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		provider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)
}

// createMenuBar creates the menu bar
func (e *EditorWindow) createMenuBar() *gtk.Box {
	menuBar := gtk.NewBox(gtk.OrientationHorizontal, 0)
	menuBar.AddCSSClass("menu-bar")

	// 文件菜单按钮
	fileButton := gtk.NewMenuButton()
	fileButton.SetLabel("文件")

	// 文件菜单
	filePopover := gtk.NewPopover()
	fileBox := gtk.NewBox(gtk.OrientationVertical, 0)
	fileBox.AddCSSClass("menu-box")

	// 新建文件
	newAction := gio.NewSimpleAction("new", nil)
	newAction.ConnectActivate(func(parameter *glib.Variant) {
		e.newFile()
	})
	e.AddAction(newAction)
	newButton := gtk.NewButton()
	newButton.SetLabel("新建 (Ctrl+N)")
	newButton.ConnectClicked(func() {
		e.newFile()
		filePopover.Hide()
	})
	fileBox.Append(newButton)

	// 打开文件
	openAction := gio.NewSimpleAction("open", nil)
	openAction.ConnectActivate(func(parameter *glib.Variant) {
		e.openFile()
	})
	e.AddAction(openAction)
	openButton := gtk.NewButton()
	openButton.SetLabel("打开 (Ctrl+O)")
	openButton.ConnectClicked(func() {
		e.openFile()
		filePopover.Hide()
	})
	fileBox.Append(openButton)

	// 保存文件
	saveAction := gio.NewSimpleAction("save", nil)
	saveAction.ConnectActivate(func(parameter *glib.Variant) {
		e.saveFile()
	})
	e.AddAction(saveAction)
	saveButton := gtk.NewButton()
	saveButton.SetLabel("保存 (Ctrl+S)")
	saveButton.ConnectClicked(func() {
		e.saveFile()
		filePopover.Hide()
	})
	fileBox.Append(saveButton)

	// 另存为
	saveAsAction := gio.NewSimpleAction("save-as", nil)
	saveAsAction.ConnectActivate(func(parameter *glib.Variant) {
		e.saveFileAs()
	})
	e.AddAction(saveAsAction)
	saveAsButton := gtk.NewButton()
	saveAsButton.SetLabel("另存为 (Ctrl+Shift+S)")
	saveAsButton.ConnectClicked(func() {
		e.saveFileAs()
		filePopover.Hide()
	})
	fileBox.Append(saveAsButton)

	// 添加自动保存选项
	autoSaveCheck := gtk.NewCheckButton()
	autoSaveCheck.SetLabel("启用自动保存 (每5分钟)")
	autoSaveCheck.SetActive(e.autoSaveEnabled)
	autoSaveCheck.ConnectToggled(func() {
		e.toggleAutoSave(autoSaveCheck.Active())
	})
	fileBox.Append(autoSaveCheck)

	// 分隔线
	separator1 := gtk.NewSeparator(gtk.OrientationHorizontal)
	fileBox.Append(separator1)

	// 退出
	quitButton := gtk.NewButton()
	quitButton.SetLabel("退出 (Ctrl+Q)")
	quitButton.ConnectClicked(func() {
		e.Application().Quit()
		filePopover.Hide()
	})
	fileBox.Append(quitButton)

	filePopover.SetChild(fileBox)
	fileButton.SetPopover(filePopover)
	menuBar.Append(fileButton)

	// 编辑菜单按钮
	editButton := gtk.NewMenuButton()
	editButton.SetLabel("编辑")

	// 编辑菜单
	editPopover := gtk.NewPopover()
	editBox := gtk.NewBox(gtk.OrientationVertical, 0)
	editBox.AddCSSClass("menu-box")

	// 撤销
	undoAction := gio.NewSimpleAction("undo", nil)
	undoAction.ConnectActivate(func(parameter *glib.Variant) {
		e.undo()
	})
	e.AddAction(undoAction)
	undoButton := gtk.NewButton()
	undoButton.SetLabel("撤销 (Ctrl+Z)")
	undoButton.ConnectClicked(func() {
		e.undo()
		editPopover.Hide()
	})
	editBox.Append(undoButton)

	// 重做
	redoAction := gio.NewSimpleAction("redo", nil)
	redoAction.ConnectActivate(func(parameter *glib.Variant) {
		e.redo()
	})
	e.AddAction(redoAction)
	redoButton := gtk.NewButton()
	redoButton.SetLabel("重做 (Ctrl+Y)")
	redoButton.ConnectClicked(func() {
		e.redo()
		editPopover.Hide()
	})
	editBox.Append(redoButton)

	// 分隔线
	separator2 := gtk.NewSeparator(gtk.OrientationHorizontal)
	editBox.Append(separator2)

	// 剪切
	cutAction := gio.NewSimpleAction("cut", nil)
	cutAction.ConnectActivate(func(parameter *glib.Variant) {
		clipboard := gdk.DisplayGetDefault().Clipboard()
		e.textView.Buffer().CutClipboard(clipboard, true)
	})
	e.AddAction(cutAction)
	cutButton := gtk.NewButton()
	cutButton.SetLabel("剪切 (Ctrl+X)")
	cutButton.ConnectClicked(func() {
		clipboard := gdk.DisplayGetDefault().Clipboard()
		e.textView.Buffer().CutClipboard(clipboard, true)
		editPopover.Hide()
	})
	editBox.Append(cutButton)

	// 复制
	copyAction := gio.NewSimpleAction("copy", nil)
	copyAction.ConnectActivate(func(parameter *glib.Variant) {
		clipboard := gdk.DisplayGetDefault().Clipboard()
		e.textView.Buffer().CopyClipboard(clipboard)
	})
	e.AddAction(copyAction)
	copyButton := gtk.NewButton()
	copyButton.SetLabel("复制 (Ctrl+C)")
	copyButton.ConnectClicked(func() {
		clipboard := gdk.DisplayGetDefault().Clipboard()
		e.textView.Buffer().CopyClipboard(clipboard)
		editPopover.Hide()
	})
	editBox.Append(copyButton)

	// 粘贴
	pasteAction := gio.NewSimpleAction("paste", nil)
	pasteAction.ConnectActivate(func(parameter *glib.Variant) {
		clipboard := gdk.DisplayGetDefault().Clipboard()
		e.textView.Buffer().PasteClipboard(clipboard, nil, true)
	})
	e.AddAction(pasteAction)
	pasteButton := gtk.NewButton()
	pasteButton.SetLabel("粘贴 (Ctrl+V)")
	pasteButton.ConnectClicked(func() {
		clipboard := gdk.DisplayGetDefault().Clipboard()
		e.textView.Buffer().PasteClipboard(clipboard, nil, true)
		editPopover.Hide()
	})
	editBox.Append(pasteButton)

	// 添加文本转换子菜单
	textTransformButton := gtk.NewButton()
	textTransformButton.SetLabel("文本转换")
	textTransformButton.ConnectClicked(func() {
		// 创建文本转换子菜单
		subPopover := gtk.NewPopover()
		subPopover.SetParent(textTransformButton)

		subBox := gtk.NewBox(gtk.OrientationVertical, 0)
		subBox.AddCSSClass("menu-box")

		// 转换为大写
		toUpperButton := gtk.NewButton()
		toUpperButton.SetLabel("转换为大写")
		toUpperButton.ConnectClicked(func() {
			e.transformToUpper()
			subPopover.Hide()
			editPopover.Hide()
		})
		subBox.Append(toUpperButton)

		// 转换为小写
		toLowerButton := gtk.NewButton()
		toLowerButton.SetLabel("转换为小写")
		toLowerButton.ConnectClicked(func() {
			e.transformToLower()
			subPopover.Hide()
			editPopover.Hide()
		})
		subBox.Append(toLowerButton)

		// 转换为标题格式
		toTitleButton := gtk.NewButton()
		toTitleButton.SetLabel("转换为标题格式")
		toTitleButton.ConnectClicked(func() {
			e.transformToTitle()
			subPopover.Hide()
			editPopover.Hide()
		})
		subBox.Append(toTitleButton)

		// 添加分隔线
		separator := gtk.NewSeparator(gtk.OrientationHorizontal)
		subBox.Append(separator)

		// 添加行注释
		addLineCommentButton := gtk.NewButton()
		addLineCommentButton.SetLabel("添加行注释 (//)")
		addLineCommentButton.ConnectClicked(func() {
			e.addLineComment()
			subPopover.Hide()
			editPopover.Hide()
		})
		subBox.Append(addLineCommentButton)

		// 移除行注释
		removeLineCommentButton := gtk.NewButton()
		removeLineCommentButton.SetLabel("移除行注释 (//)")
		removeLineCommentButton.ConnectClicked(func() {
			e.removeLineComment()
			subPopover.Hide()
			editPopover.Hide()
		})
		subBox.Append(removeLineCommentButton)

		// 添加块注释
		addBlockCommentButton := gtk.NewButton()
		addBlockCommentButton.SetLabel("添加块注释 (/* */)")
		addBlockCommentButton.ConnectClicked(func() {
			e.addBlockComment()
			subPopover.Hide()
			editPopover.Hide()
		})
		subBox.Append(addBlockCommentButton)

		subPopover.SetChild(subBox)
		subPopover.Show()
	})
	editBox.Append(textTransformButton)

	editPopover.SetChild(editBox)
	editButton.SetPopover(editPopover)
	menuBar.Append(editButton)

	// 搜索菜单按钮
	searchButton := gtk.NewMenuButton()
	searchButton.SetLabel("搜索")

	// 搜索菜单
	searchPopover := gtk.NewPopover()
	searchBox := gtk.NewBox(gtk.OrientationVertical, 0)
	searchBox.AddCSSClass("menu-box")

	// 查找
	findAction := gio.NewSimpleAction("find", nil)
	findAction.ConnectActivate(func(parameter *glib.Variant) {
		e.showSearchBar()
	})
	e.AddAction(findAction)
	findButton := gtk.NewButton()
	findButton.SetLabel("查找 (Ctrl+F)")
	findButton.ConnectClicked(func() {
		e.showSearchBar()
		searchPopover.Hide()
	})
	searchBox.Append(findButton)

	// 替换
	replaceAction := gio.NewSimpleAction("replace", nil)
	replaceAction.ConnectActivate(func(parameter *glib.Variant) {
		e.showReplaceDialog()
	})
	e.AddAction(replaceAction)
	replaceButton := gtk.NewButton()
	replaceButton.SetLabel("替换 (Ctrl+H)")
	replaceButton.ConnectClicked(func() {
		e.showReplaceDialog()
		searchPopover.Hide()
	})
	searchBox.Append(replaceButton)

	searchPopover.SetChild(searchBox)
	searchButton.SetPopover(searchPopover)
	menuBar.Append(searchButton)

	// 帮助菜单按钮
	helpButton := gtk.NewMenuButton()
	helpButton.SetLabel("帮助")

	// 帮助菜单
	helpPopover := gtk.NewPopover()
	helpBox := gtk.NewBox(gtk.OrientationVertical, 0)
	helpBox.AddCSSClass("menu-box")

	// 关于
	aboutAction := gio.NewSimpleAction("about", nil)
	aboutAction.ConnectActivate(func(parameter *glib.Variant) {
		e.showAboutDialog()
	})
	e.AddAction(aboutAction)
	aboutButton := gtk.NewButton()
	aboutButton.SetLabel("关于")
	aboutButton.ConnectClicked(func() {
		e.showAboutDialog()
		helpPopover.Hide()
	})
	helpBox.Append(aboutButton)

	helpPopover.SetChild(helpBox)
	helpButton.SetPopover(helpPopover)
	menuBar.Append(helpButton)

	return menuBar
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
	// Clear the buffer
	e.textBuffer.SetText("")
	e.bufferAdapter.updateGtkBuffer()

	// Reset the file path and title
	e.filePath = ""
	e.SetTitle("Orange Editor - 未命名")

	// Reset the modified flag
	e.textBuffer.SetModified(false)

	// Update the status bar
	e.updateStatusBar()
}

// openFile opens a file
func (e *EditorWindow) openFile() {
	// For now, we'll use a hardcoded file path for testing
	// In a real application, you would use a file chooser dialog
	filename := "test.txt"

	// Check if the file exists
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// Create an empty file
		err = os.WriteFile(filename, []byte(""), 0644)
		if err != nil {
			e.showErrorDialog("无法创建文件", err.Error())
			return
		}
	}

	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		e.showErrorDialog("无法打开文件", err.Error())
		return
	}

	// Update the buffer
	e.textBuffer.SetText(string(content))
	e.bufferAdapter.updateGtkBuffer()
	e.filePath = filename
	e.SetTitle("Orange Editor - " + filename)
	e.textBuffer.SetModified(false)
	e.updateStatusBar()
}

// saveFile saves the current file
func (e *EditorWindow) saveFile() {
	if e.filePath == "" {
		e.saveFileAs()
		return
	}

	content := e.textBuffer.GetText()
	err := os.WriteFile(e.filePath, []byte(content), 0644)
	if err != nil {
		e.showErrorDialog("无法保存文件", err.Error())
		return
	}
	e.textBuffer.SetModified(false)
	e.updateStatusBar()
}

// saveFileAs saves the current file with a new name
func (e *EditorWindow) saveFileAs() {
	// For now, we'll use a hardcoded file path for testing
	// In a real application, you would use a file chooser dialog
	filename := "test_save.txt"

	// Save the file
	content := e.textBuffer.GetText()
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		e.showErrorDialog("无法保存文件", err.Error())
		return
	}

	// Update the file path and title
	e.filePath = filename
	e.SetTitle("Orange Editor - " + filename)
	e.textBuffer.SetModified(false)
	e.updateStatusBar()
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
}

// search performs a search operation
func (e *EditorWindow) search(forward bool) {
	searchText := e.searchEntry.Text()
	if searchText == "" {
		return
	}

	buffer := e.bufferAdapter.GetGtkBuffer()

	// Get the text from the buffer
	text := buffer.Text(buffer.StartIter(), buffer.EndIter(), false)

	// Find the text
	index := -1
	if e.caseSensitive {
		index = strings.Index(text, searchText)
	} else {
		index = strings.Index(strings.ToLower(text), strings.ToLower(searchText))
	}

	if index >= 0 {
		// Found the text, select it
		startIter := buffer.StartIter()
		startIter.ForwardChars(index)

		endIter := buffer.StartIter()
		endIter.ForwardChars(index + len(searchText))

		// Select the text
		buffer.SelectRange(startIter, endIter)

		// Scroll to the selection
		e.textView.ScrollToIter(startIter, 0.0, true, 0.0, 0.5)
	} else {
		fmt.Println("找不到匹配的文本")
	}
}

// showReplaceDialog shows the replace dialog
func (e *EditorWindow) showReplaceDialog() {
	// 创建替换对话框
	dialog := gtk.NewDialog()
	dialog.SetTitle("替换")
	// 不设置TransientFor，避免类型问题
	dialog.SetModal(true)
	dialog.SetDefaultSize(400, 200)

	// 添加按钮
	dialog.AddButton("取消", int(gtk.ResponseCancel))
	dialog.AddButton("替换", int(gtk.ResponseAccept))
	dialog.AddButton("全部替换", int(gtk.ResponseApply))
	dialog.SetDefaultResponse(int(gtk.ResponseAccept))

	// 创建内容区域
	contentArea := dialog.ContentArea()
	grid := gtk.NewGrid()
	grid.SetRowSpacing(8)
	grid.SetColumnSpacing(8)
	grid.SetMarginTop(16)
	grid.SetMarginBottom(16)
	grid.SetMarginStart(16)
	grid.SetMarginEnd(16)

	// 查找标签和输入框
	findLabel := gtk.NewLabel("查找内容:")
	findLabel.SetHAlign(gtk.AlignStart)
	grid.Attach(findLabel, 0, 0, 1, 1)

	findEntry := gtk.NewEntry()
	findEntry.SetHExpand(true)
	// 如果有选中的文本，则填入查找框
	if buffer := e.bufferAdapter.GetGtkBuffer(); buffer.HasSelection() {
		startIter, endIter, _ := buffer.SelectionBounds()
		selectedText := buffer.Text(startIter, endIter, false)
		findEntry.SetText(selectedText)
	}
	grid.Attach(findEntry, 1, 0, 1, 1)

	// 替换标签和输入框
	replaceLabel := gtk.NewLabel("替换为:")
	replaceLabel.SetHAlign(gtk.AlignStart)
	grid.Attach(replaceLabel, 0, 1, 1, 1)

	replaceEntry := gtk.NewEntry()
	replaceEntry.SetHExpand(true)
	grid.Attach(replaceEntry, 1, 1, 1, 1)

	// 选项区域
	optionsFrame := gtk.NewFrame("选项")
	optionsBox := gtk.NewBox(gtk.OrientationVertical, 4)
	optionsBox.SetMarginTop(8)
	optionsBox.SetMarginBottom(8)
	optionsBox.SetMarginStart(8)
	optionsBox.SetMarginEnd(8)

	// 区分大小写选项
	caseSensitiveCheck := gtk.NewCheckButton()
	caseSensitiveCheck.SetLabel("区分大小写")
	caseSensitiveCheck.SetActive(e.caseSensitive)
	optionsBox.Append(caseSensitiveCheck)

	// 全字匹配选项
	wholeWordCheck := gtk.NewCheckButton()
	wholeWordCheck.SetLabel("全字匹配")
	wholeWordCheck.SetActive(e.wholeWord)
	optionsBox.Append(wholeWordCheck)

	// 正则表达式选项
	regexCheck := gtk.NewCheckButton()
	regexCheck.SetLabel("使用正则表达式")
	regexCheck.SetActive(e.useRegex)
	optionsBox.Append(regexCheck)

	optionsFrame.SetChild(optionsBox)
	grid.Attach(optionsFrame, 0, 2, 2, 1)

	contentArea.Append(grid)
	dialog.Show()

	// 处理对话框响应
	dialog.Connect("response", func(dialog *gtk.Dialog, responseId int) {
		findText := findEntry.Text()
		replaceText := replaceEntry.Text()
		caseSensitive := caseSensitiveCheck.Active()
		wholeWord := wholeWordCheck.Active()
		useRegex := regexCheck.Active()

		// 更新搜索选项
		e.caseSensitive = caseSensitive
		e.wholeWord = wholeWord
		e.useRegex = useRegex

		if responseId == int(gtk.ResponseAccept) {
			// 替换当前选中或下一个匹配项
			e.replaceNext(findText, replaceText, caseSensitive)
		} else if responseId == int(gtk.ResponseApply) {
			// 全部替换
			count := e.replaceAll(findText, replaceText, caseSensitive)
			e.showInfoDialog("替换完成", fmt.Sprintf("已替换 %d 处匹配项", count))
		}

		dialog.Destroy()
	})
}

// replaceNext replaces the next occurrence of the search text
func (e *EditorWindow) replaceNext(findText, replaceText string, caseSensitive bool) {
	buffer := e.bufferAdapter.GetGtkBuffer()
	if buffer.HasSelection() {
		// Get the selected text
		startIter, endIter, _ := buffer.SelectionBounds()
		selectedText := buffer.Text(startIter, endIter, false)

		// Check if the selected text matches
		if (caseSensitive && selectedText == findText) ||
			(!caseSensitive && strings.ToLower(selectedText) == strings.ToLower(findText)) {
			// Replace the text
			buffer.DeleteSelection(false, true)
			buffer.InsertAtCursor(replaceText)
		}
	}

	// Search for the next occurrence
	e.caseSensitive = caseSensitive
	e.search(true)
}

// replaceAll replaces all occurrences of the search text
func (e *EditorWindow) replaceAll(findText, replaceText string, caseSensitive bool) int {
	// Get the text
	buffer := e.bufferAdapter.GetGtkBuffer()
	text := buffer.Text(buffer.StartIter(), buffer.EndIter(), false)

	// 计算替换次数
	count := 0
	var newText string

	if e.useRegex {
		// 使用正则表达式替换
		var re *regexp.Regexp
		if caseSensitive {
			re = regexp.MustCompile(regexp.QuoteMeta(findText))
		} else {
			re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(findText))
		}

		// 计算匹配数量
		matches := re.FindAllString(text, -1)
		count = len(matches)

		// 执行替换
		newText = re.ReplaceAllString(text, replaceText)
	} else if e.wholeWord {
		// 全字匹配替换
		wordBoundary := `\b`
		var re *regexp.Regexp
		if caseSensitive {
			re = regexp.MustCompile(wordBoundary + regexp.QuoteMeta(findText) + wordBoundary)
		} else {
			re = regexp.MustCompile("(?i)" + wordBoundary + regexp.QuoteMeta(findText) + wordBoundary)
		}

		// 计算匹配数量
		matches := re.FindAllString(text, -1)
		count = len(matches)

		// 执行替换
		newText = re.ReplaceAllString(text, replaceText)
	} else {
		// 普通替换
		if caseSensitive {
			// 计算匹配数量
			count = strings.Count(text, findText)
			// 执行替换
			newText = strings.ReplaceAll(text, findText, replaceText)
		} else {
			// 不区分大小写的替换需要手动处理
			lowerText := strings.ToLower(text)
			lowerFind := strings.ToLower(findText)

			// 计算匹配数量
			count = strings.Count(lowerText, lowerFind)

			// 使用正则表达式进行不区分大小写的替换
			re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(findText))
			newText = re.ReplaceAllString(text, replaceText)
		}
	}

	// Update the buffer
	buffer.SetText(newText)
	e.textBuffer.SetText(newText)

	return count
}

// showErrorDialog shows an error dialog
func (e *EditorWindow) showErrorDialog(title, message string) {
	fmt.Printf("错误: %s - %s\n", title, message)
}

// showInfoDialog shows an information dialog
func (e *EditorWindow) showInfoDialog(title, message string) {
	fmt.Printf("信息: %s - %s\n", title, message)
}

// updateGtkBuffer updates the GTK buffer from the text buffer
func (e *EditorWindow) updateGtkBuffer() {
	e.bufferAdapter.updateGtkBuffer()
}

// updateStatusBar updates the status bar with the current cursor position
func (e *EditorWindow) updateStatusBar() {
	// Clear the status bar
	e.statusBar.Remove(e.contextID, e.statusBar.ContextID("editor"))

	// Get the cursor position
	position := e.bufferAdapter.GetCursorPosition()

	// Get the total number of lines
	buffer := e.bufferAdapter.GetGtkBuffer()
	totalLines := buffer.LineCount()

	// Get the total number of characters
	totalChars := buffer.CharCount()

	// Get the modified status
	modified := e.textBuffer.IsModified()
	modifiedStr := ""
	if modified {
		modifiedStr = "已修改"
	} else {
		modifiedStr = "未修改"
	}

	// Get the file path
	filePath := e.filePath
	if filePath == "" {
		filePath = "未命名"
	}

	// Format the status message
	statusMsg := fmt.Sprintf("行: %d/%d | 列: %d | 字符: %d | %s | %s",
		position.Line+1, totalLines, position.Column+1, totalChars, modifiedStr, filePath)

	// Push the status message
	e.statusBar.Push(e.contextID, statusMsg)
}

// setupClipboardActions sets up the clipboard actions
func (e *EditorWindow) setupClipboardActions() {
	// 剪切操作
	cutAction := gio.NewSimpleAction("cut", nil)
	cutAction.ConnectActivate(func(_ *glib.Variant) {
		// 使用信号名称调用
		e.textView.Emit("cut-clipboard")
	})
	e.ApplicationWindow.AddAction(cutAction)

	// 复制操作
	copyAction := gio.NewSimpleAction("copy", nil)
	copyAction.ConnectActivate(func(_ *glib.Variant) {
		// 使用信号名称调用
		e.textView.Emit("copy-clipboard")
	})
	e.ApplicationWindow.AddAction(copyAction)

	// 粘贴操作
	pasteAction := gio.NewSimpleAction("paste", nil)
	pasteAction.ConnectActivate(func(_ *glib.Variant) {
		// 使用信号名称调用
		e.textView.Emit("paste-clipboard")
	})
	e.ApplicationWindow.AddAction(pasteAction)
}

// getSelectedText returns the selected text
func (e *EditorWindow) getSelectedText() string {
	gtkBuffer := e.bufferAdapter.GetGtkBuffer()
	if gtkBuffer.HasSelection() {
		// 简化实现，不使用 SelectionBounds
		return ""
	}
	return ""
}

// showAboutDialog shows the about dialog
func (e *EditorWindow) showAboutDialog() {
	fmt.Println("Orange Editor 1.0")
	fmt.Println("一个基于GTK4的文本编辑器")
	fmt.Println("© 2023")
	fmt.Println("MIT License")
	fmt.Println("https://github.com/example/gotextbuffer")
}

// duplicateLine duplicates the current line or selected lines
func (e *EditorWindow) duplicateLine() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	// 开始用户操作，以便撤销时作为一个整体
	buffer.BeginUserAction()

	if buffer.HasSelection() {
		// 如果有选择，复制选择的文本
		startIter, endIter, _ := buffer.SelectionBounds()

		// 确保选择整行
		startLine := startIter.Line()
		endLine := endIter.Line()

		// 获取行首位置
		lineStartIter := buffer.StartIter()
		for i := 0; i < startLine; i++ {
			lineStartIter.ForwardLine()
		}

		// 获取行尾位置（下一行的开始）
		lineEndIter := buffer.StartIter()
		for i := 0; i < endLine; i++ {
			lineEndIter.ForwardLine()
		}
		lineEndIter.ForwardLine()

		// 获取要复制的文本（包括换行符）
		textToDuplicate := buffer.Text(lineStartIter, lineEndIter, false)

		// 插入复制的文本
		buffer.Insert(lineEndIter, textToDuplicate)

		// 选择新插入的文本
		newStartIter := lineEndIter.Copy()
		newEndIter := newStartIter.Copy()
		for i := 0; i < endLine-startLine+1; i++ {
			newEndIter.ForwardLine()
		}
		buffer.SelectRange(newStartIter, newEndIter)
	} else {
		// 如果没有选择，复制当前行
		cursor := buffer.GetInsert()
		cursorIter := buffer.IterAtMark(cursor)
		currentLine := cursorIter.Line()

		// 获取当前行的开始和结束位置
		lineStartIter := buffer.StartIter()
		for i := 0; i < currentLine; i++ {
			lineStartIter.ForwardLine()
		}

		lineEndIter := lineStartIter.Copy()
		lineEndIter.ForwardLine()

		// 获取当前行的文本（包括换行符）
		currentLineText := buffer.Text(lineStartIter, lineEndIter, false)

		// 如果是最后一行且没有换行符，添加一个换行符
		if lineEndIter.IsEnd() && !strings.HasSuffix(currentLineText, "\n") {
			currentLineText += "\n"
		}

		// 插入复制的行
		buffer.Insert(lineEndIter, currentLineText)

		// 将光标移动到新行
		newLineIter := lineEndIter.Copy()
		buffer.PlaceCursor(newLineIter)
	}

	buffer.EndUserAction()

	// 更新自定义文本缓冲区
	e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
}

// deleteLine deletes the current line or selected lines
func (e *EditorWindow) deleteLine() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	// 开始用户操作，以便撤销时作为一个整体
	buffer.BeginUserAction()

	if buffer.HasSelection() {
		// 如果有选择，删除选择的所有行
		startIter, endIter, _ := buffer.SelectionBounds()

		// 确保选择整行
		startLine := startIter.Line()
		endLine := endIter.Line()

		// 获取行首位置
		lineStartIter := buffer.StartIter()
		for i := 0; i < startLine; i++ {
			lineStartIter.ForwardLine()
		}

		// 获取行尾位置（下一行的开始）
		lineEndIter := buffer.StartIter()
		for i := 0; i < endLine; i++ {
			lineEndIter.ForwardLine()
		}
		lineEndIter.ForwardLine()

		// 删除选中的行
		buffer.Delete(lineStartIter, lineEndIter)
	} else {
		// 如果没有选择，删除当前行
		cursor := buffer.GetInsert()
		cursorIter := buffer.IterAtMark(cursor)
		currentLine := cursorIter.Line()

		// 获取当前行的开始和结束位置
		lineStartIter := buffer.StartIter()
		for i := 0; i < currentLine; i++ {
			lineStartIter.ForwardLine()
		}

		lineEndIter := lineStartIter.Copy()
		lineEndIter.ForwardLine()

		// 删除当前行
		buffer.Delete(lineStartIter, lineEndIter)

		// 如果删除的是最后一行，且前面还有行，则将光标移动到前一行的末尾
		if lineEndIter.IsEnd() && currentLine > 0 {
			prevLineIter := buffer.StartIter()
			for i := 0; i < currentLine-1; i++ {
				prevLineIter.ForwardLine()
			}
			prevLineIter.ForwardToLineEnd()
			buffer.PlaceCursor(prevLineIter)
		}
	}

	buffer.EndUserAction()

	// 更新自定义文本缓冲区
	e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
}

// moveLineUp moves the current line or selected lines up
func (e *EditorWindow) moveLineUp() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	// 开始用户操作，以便撤销时作为一个整体
	buffer.BeginUserAction()

	var startLine, endLine int

	if buffer.HasSelection() {
		// 如果有选择，移动选择的所有行
		startIter, endIter, _ := buffer.SelectionBounds()
		startLine = startIter.Line()
		endLine = endIter.Line()
	} else {
		// 如果没有选择，移动当前行
		cursor := buffer.GetInsert()
		cursorIter := buffer.IterAtMark(cursor)
		startLine = cursorIter.Line()
		endLine = startLine
	}

	// 如果已经是第一行，则不执行操作
	if startLine <= 0 {
		buffer.EndUserAction()
		return
	}

	// 获取要移动的行的开始和结束位置
	moveStartIter := buffer.StartIter()
	for i := 0; i < startLine; i++ {
		moveStartIter.ForwardLine()
	}

	moveEndIter := buffer.StartIter()
	for i := 0; i < endLine; i++ {
		moveEndIter.ForwardLine()
	}
	moveEndIter.ForwardLine()

	// 获取要移动的文本
	textToMove := buffer.Text(moveStartIter, moveEndIter, false)

	// 获取上一行的开始和结束位置
	prevLineStartIter := buffer.StartIter()
	for i := 0; i < startLine-1; i++ {
		prevLineStartIter.ForwardLine()
	}

	prevLineEndIter := prevLineStartIter.Copy()
	prevLineEndIter.ForwardLine()

	// 获取上一行的文本
	prevLineText := buffer.Text(prevLineStartIter, prevLineEndIter, false)

	// 删除要移动的行和上一行
	buffer.Delete(prevLineStartIter, moveEndIter)

	// 先插入要移动的文本，再插入上一行的文本
	buffer.Insert(prevLineStartIter, textToMove)
	buffer.Insert(prevLineStartIter, prevLineText)

	// 更新选择或光标位置
	if buffer.HasSelection() {
		// 更新选择
		newStartIter := buffer.StartIter()
		for i := 0; i < startLine-1; i++ {
			newStartIter.ForwardLine()
		}

		newEndIter := buffer.StartIter()
		for i := 0; i < endLine-1; i++ {
			newEndIter.ForwardLine()
		}
		newEndIter.ForwardToLineEnd()

		buffer.SelectRange(newStartIter, newEndIter)
	} else {
		// 更新光标位置
		newCursorIter := buffer.StartIter()
		for i := 0; i < startLine-1; i++ {
			newCursorIter.ForwardLine()
		}

		// 保持光标在行内的相对位置
		cursor := buffer.GetInsert()
		oldCursorIter := buffer.IterAtMark(cursor)
		offset := oldCursorIter.LineOffset()

		// 移动到相同的列位置
		for i := 0; i < offset && !newCursorIter.IsEnd() && !newCursorIter.EndsLine(); i++ {
			newCursorIter.ForwardChar()
		}

		buffer.PlaceCursor(newCursorIter)
	}

	buffer.EndUserAction()

	// 更新自定义文本缓冲区
	e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
}

// moveLineDown moves the current line or selected lines down
func (e *EditorWindow) moveLineDown() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	// 开始用户操作，以便撤销时作为一个整体
	buffer.BeginUserAction()

	var startLine, endLine int

	if buffer.HasSelection() {
		// 如果有选择，移动选择的所有行
		startIter, endIter, _ := buffer.SelectionBounds()
		startLine = startIter.Line()
		endLine = endIter.Line()
	} else {
		// 如果没有选择，移动当前行
		cursor := buffer.GetInsert()
		cursorIter := buffer.IterAtMark(cursor)
		startLine = cursorIter.Line()
		endLine = startLine
	}

	// 获取总行数
	totalLines := buffer.LineCount()

	// 如果已经是最后一行，则不执行操作
	if endLine >= totalLines-1 {
		buffer.EndUserAction()
		return
	}

	// 获取要移动的行的开始和结束位置
	moveStartIter := buffer.StartIter()
	for i := 0; i < startLine; i++ {
		moveStartIter.ForwardLine()
	}

	moveEndIter := buffer.StartIter()
	for i := 0; i < endLine; i++ {
		moveEndIter.ForwardLine()
	}
	moveEndIter.ForwardLine()

	// 获取要移动的文本
	textToMove := buffer.Text(moveStartIter, moveEndIter, false)

	// 获取下一行的开始和结束位置
	nextLineStartIter := moveEndIter.Copy()
	nextLineEndIter := nextLineStartIter.Copy()
	nextLineEndIter.ForwardLine()

	// 获取下一行的文本
	nextLineText := buffer.Text(nextLineStartIter, nextLineEndIter, false)

	// 删除要移动的行和下一行
	buffer.Delete(moveStartIter, nextLineEndIter)

	// 先插入下一行的文本，再插入要移动的文本
	buffer.Insert(moveStartIter, nextLineText)
	buffer.Insert(moveStartIter, textToMove)

	// 更新选择或光标位置
	if buffer.HasSelection() {
		// 更新选择
		newStartIter := buffer.StartIter()
		for i := 0; i < startLine+1; i++ {
			newStartIter.ForwardLine()
		}

		newEndIter := buffer.StartIter()
		for i := 0; i < endLine+1; i++ {
			newEndIter.ForwardLine()
		}
		newEndIter.ForwardToLineEnd()

		buffer.SelectRange(newStartIter, newEndIter)
	} else {
		// 更新光标位置
		newCursorIter := buffer.StartIter()
		for i := 0; i < startLine+1; i++ {
			newCursorIter.ForwardLine()
		}

		// 保持光标在行内的相对位置
		cursor := buffer.GetInsert()
		oldCursorIter := buffer.IterAtMark(cursor)
		offset := oldCursorIter.LineOffset()

		// 移动到相同的列位置
		for i := 0; i < offset && !newCursorIter.IsEnd() && !newCursorIter.EndsLine(); i++ {
			newCursorIter.ForwardChar()
		}

		buffer.PlaceCursor(newCursorIter)
	}

	buffer.EndUserAction()

	// 更新自定义文本缓冲区
	e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
}

// increaseIndent increases the indentation of the current line or selected lines
func (e *EditorWindow) increaseIndent() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	// 开始用户操作，以便撤销时作为一个整体
	buffer.BeginUserAction()

	var startLine, endLine int

	if buffer.HasSelection() {
		// 如果有选择，增加选择的所有行的缩进
		startIter, endIter, _ := buffer.SelectionBounds()
		startLine = startIter.Line()
		endLine = endIter.Line()
	} else {
		// 如果没有选择，增加当前行的缩进
		cursor := buffer.GetInsert()
		cursorIter := buffer.IterAtMark(cursor)
		startLine = cursorIter.Line()
		endLine = startLine
	}

	// 记录是否有选择
	var hasSelection bool
	if buffer.HasSelection() {
		hasSelection = true
	}

	// 为每一行增加缩进
	for line := startLine; line <= endLine; line++ {
		// 获取行首位置
		lineStartIter := buffer.StartIter()
		for i := 0; i < line; i++ {
			lineStartIter.ForwardLine()
		}

		// 在行首添加制表符
		buffer.Insert(lineStartIter, "\t")
	}

	// 恢复选择
	if hasSelection {
		// 由于文本已经改变，我们需要重新计算选择范围
		newStartIter := buffer.StartIter()
		for i := 0; i < startLine; i++ {
			newStartIter.ForwardLine()
		}

		newEndIter := buffer.StartIter()
		for i := 0; i < endLine; i++ {
			newEndIter.ForwardLine()
		}
		newEndIter.ForwardToLineEnd()

		buffer.SelectRange(newStartIter, newEndIter)
	}

	buffer.EndUserAction()

	// 更新自定义文本缓冲区
	e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
}

// decreaseIndent decreases the indentation of the current line or selected lines
func (e *EditorWindow) decreaseIndent() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	// 开始用户操作，以便撤销时作为一个整体
	buffer.BeginUserAction()

	var startLine, endLine int

	if buffer.HasSelection() {
		// 如果有选择，减少选择的所有行的缩进
		startIter, endIter, _ := buffer.SelectionBounds()
		startLine = startIter.Line()
		endLine = endIter.Line()
	} else {
		// 如果没有选择，减少当前行的缩进
		cursor := buffer.GetInsert()
		cursorIter := buffer.IterAtMark(cursor)
		startLine = cursorIter.Line()
		endLine = startLine
	}

	// 记录是否有选择
	var hasSelection bool
	if buffer.HasSelection() {
		hasSelection = true
	}

	// 为每一行减少缩进
	for line := startLine; line <= endLine; line++ {
		// 获取行首位置
		lineStartIter := buffer.StartIter()
		for i := 0; i < line; i++ {
			lineStartIter.ForwardLine()
		}

		// 获取行的文本
		lineEndIter := lineStartIter.Copy()
		lineEndIter.ForwardToLineEnd()
		lineText := buffer.Text(lineStartIter, lineEndIter, false)

		// 检查行首是否有制表符或空格
		if len(lineText) > 0 {
			if lineText[0] == '\t' {
				// 删除行首的制表符
				deleteEndIter := lineStartIter.Copy()
				deleteEndIter.ForwardChar()
				buffer.Delete(lineStartIter, deleteEndIter)
			} else if lineText[0] == ' ' {
				// 计算行首有多少个连续空格
				spaceCount := 0
				for i := 0; i < len(lineText) && lineText[i] == ' '; i++ {
					spaceCount++
				}

				// 最多删除4个空格（一个制表符的宽度）
				deleteCount := min(spaceCount, 4)
				if deleteCount > 0 {
					deleteEndIter := lineStartIter.Copy()
					for i := 0; i < deleteCount; i++ {
						deleteEndIter.ForwardChar()
					}
					buffer.Delete(lineStartIter, deleteEndIter)
				}
			}
		}
	}

	// 恢复选择
	if hasSelection {
		// 由于文本已经改变，我们需要重新计算选择范围
		newStartIter := buffer.StartIter()
		for i := 0; i < startLine; i++ {
			newStartIter.ForwardLine()
		}

		newEndIter := buffer.StartIter()
		for i := 0; i < endLine; i++ {
			newEndIter.ForwardLine()
		}
		newEndIter.ForwardToLineEnd()

		buffer.SelectRange(newStartIter, newEndIter)
	}

	buffer.EndUserAction()

	// 更新自定义文本缓冲区
	e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// transformToUpper transforms the selected text to uppercase
func (e *EditorWindow) transformToUpper() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	if buffer.HasSelection() {
		startIter, endIter, _ := buffer.SelectionBounds()
		selectedText := buffer.Text(startIter, endIter, false)

		// 转换为大写
		upperText := strings.ToUpper(selectedText)

		// 替换选中的文本
		buffer.BeginUserAction()
		buffer.Delete(startIter, endIter)
		buffer.Insert(startIter, upperText)
		buffer.EndUserAction()

		// 更新自定义文本缓冲区
		e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
	} else {
		e.showInfoDialog("提示", "请先选择要转换的文本")
	}
}

// transformToLower transforms the selected text to lowercase
func (e *EditorWindow) transformToLower() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	if buffer.HasSelection() {
		startIter, endIter, _ := buffer.SelectionBounds()
		selectedText := buffer.Text(startIter, endIter, false)

		// 转换为小写
		lowerText := strings.ToLower(selectedText)

		// 替换选中的文本
		buffer.BeginUserAction()
		buffer.Delete(startIter, endIter)
		buffer.Insert(startIter, lowerText)
		buffer.EndUserAction()

		// 更新自定义文本缓冲区
		e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
	} else {
		e.showInfoDialog("提示", "请先选择要转换的文本")
	}
}

// transformToTitle transforms the selected text to title case
func (e *EditorWindow) transformToTitle() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	if buffer.HasSelection() {
		startIter, endIter, _ := buffer.SelectionBounds()
		selectedText := buffer.Text(startIter, endIter, false)

		// 转换为标题格式
		titleText := strings.Title(selectedText)

		// 替换选中的文本
		buffer.BeginUserAction()
		buffer.Delete(startIter, endIter)
		buffer.Insert(startIter, titleText)
		buffer.EndUserAction()

		// 更新自定义文本缓冲区
		e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
	} else {
		e.showInfoDialog("提示", "请先选择要转换的文本")
	}
}

// addLineComment adds line comments to the selected lines
func (e *EditorWindow) addLineComment() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	if buffer.HasSelection() {
		startIter, endIter, _ := buffer.SelectionBounds()
		startLine := startIter.Line()
		endLine := endIter.Line()

		buffer.BeginUserAction()

		// 为每一行添加注释
		for line := startLine; line <= endLine; line++ {
			// 获取行首位置
			lineIter := buffer.StartIter()
			for i := 0; i < line; i++ {
				lineIter.ForwardLine()
			}

			// 在行首添加注释
			buffer.Insert(lineIter, "// ")
		}

		buffer.EndUserAction()

		// 更新自定义文本缓冲区
		e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
	} else {
		// 如果没有选择，为当前行添加注释
		cursor := buffer.GetInsert()
		iter := buffer.IterAtMark(cursor)
		line := iter.Line()

		// 获取行首位置
		lineIter := buffer.StartIter()
		for i := 0; i < line; i++ {
			lineIter.ForwardLine()
		}

		// 在行首添加注释
		buffer.BeginUserAction()
		buffer.Insert(lineIter, "// ")
		buffer.EndUserAction()

		// 更新自定义文本缓冲区
		e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
	}
}

// removeLineComment removes line comments from the selected lines
func (e *EditorWindow) removeLineComment() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	if buffer.HasSelection() {
		startIter, endIter, _ := buffer.SelectionBounds()
		startLine := startIter.Line()
		endLine := endIter.Line()

		buffer.BeginUserAction()

		// 为每一行移除注释
		for line := startLine; line <= endLine; line++ {
			// 获取行首位置
			lineIter := buffer.StartIter()
			for i := 0; i < line; i++ {
				lineIter.ForwardLine()
			}

			// 检查行首是否有注释
			lineEndIter := lineIter.Copy()
			lineEndIter.ForwardToLineEnd()
			lineText := buffer.Text(lineIter, lineEndIter, false)

			// 移除注释
			if strings.HasPrefix(strings.TrimSpace(lineText), "//") {
				// 找到注释的位置
				commentPos := strings.Index(lineText, "//")
				if commentPos >= 0 {
					commentIter := lineIter.Copy()
					commentIter.ForwardChars(commentPos)

					// 删除注释符号
					endCommentIter := commentIter.Copy()
					endCommentIter.ForwardChars(2)
					buffer.Delete(commentIter, endCommentIter)

					// 如果注释后有空格，也删除它
					// 由于GTK4的TextIter没有GetChar方法，我们使用字符串检查
					if commentPos+2 < len(lineText) && lineText[commentPos+2] == ' ' {
						spaceIter := endCommentIter.Copy()
						spaceIter.ForwardChar()
						buffer.Delete(endCommentIter, spaceIter)
					}
				}
			}
		}

		buffer.EndUserAction()

		// 更新自定义文本缓冲区
		e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
	} else {
		// 如果没有选择，为当前行移除注释
		cursor := buffer.GetInsert()
		iter := buffer.IterAtMark(cursor)
		line := iter.Line()

		// 获取行首位置
		lineIter := buffer.StartIter()
		for i := 0; i < line; i++ {
			lineIter.ForwardLine()
		}

		// 检查行首是否有注释
		lineEndIter := lineIter.Copy()
		lineEndIter.ForwardToLineEnd()
		lineText := buffer.Text(lineIter, lineEndIter, false)

		// 移除注释
		if strings.HasPrefix(strings.TrimSpace(lineText), "//") {
			// 找到注释的位置
			commentPos := strings.Index(lineText, "//")
			if commentPos >= 0 {
				commentIter := lineIter.Copy()
				commentIter.ForwardChars(commentPos)

				// 删除注释符号
				endCommentIter := commentIter.Copy()
				endCommentIter.ForwardChars(2)
				buffer.Delete(commentIter, endCommentIter)

				// 如果注释后有空格，也删除它
				// 由于GTK4的TextIter没有GetChar方法，我们使用字符串检查
				if commentPos+2 < len(lineText) && lineText[commentPos+2] == ' ' {
					spaceIter := endCommentIter.Copy()
					spaceIter.ForwardChar()
					buffer.Delete(endCommentIter, spaceIter)
				}
			}
		}
	}
}

// addBlockComment adds block comments around the selected text
func (e *EditorWindow) addBlockComment() {
	buffer := e.bufferAdapter.GetGtkBuffer()

	if buffer.HasSelection() {
		startIter, endIter, _ := buffer.SelectionBounds()
		selectedText := buffer.Text(startIter, endIter, false)

		// 添加块注释
		commentedText := "/* " + selectedText + " */"

		// 替换选中的文本
		buffer.BeginUserAction()
		buffer.Delete(startIter, endIter)
		buffer.Insert(startIter, commentedText)
		buffer.EndUserAction()

		// 更新自定义文本缓冲区
		e.textBuffer.SetText(buffer.Text(buffer.StartIter(), buffer.EndIter(), false))
	} else {
		e.showInfoDialog("提示", "请先选择要添加注释的文本")
	}
}

// toggleAutoSave enables or disables auto-save
func (e *EditorWindow) toggleAutoSave(enabled bool) {
	e.autoSaveEnabled = enabled

	if enabled {
		// 如果启用自动保存，创建定时器
		if e.autoSaveTimer != nil {
			e.autoSaveTimer.Stop()
		}

		// 每5分钟自动保存一次
		e.autoSaveTimer = time.AfterFunc(5*time.Minute, func() {
			// 在主线程中执行保存操作
			glib.IdleAdd(func() {
				if e.filePath != "" {
					e.saveFile()
					e.showInfoDialog("自动保存", "文件已自动保存")
				}

				// 重新启动定时器
				e.autoSaveTimer = time.AfterFunc(5*time.Minute, func() {
					glib.IdleAdd(func() {
						if e.filePath != "" {
							e.saveFile()
						}
						e.toggleAutoSave(true) // 重新启动定时器
					})
				})
			})
		})
	} else {
		// 如果禁用自动保存，停止定时器
		if e.autoSaveTimer != nil {
			e.autoSaveTimer.Stop()
			e.autoSaveTimer = nil
		}
	}
}
