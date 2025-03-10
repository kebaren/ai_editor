package editor

import (
	"github.com/example/gotextbuffer/textbuffer"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// EditorWindow 表示编辑器主窗口
type EditorWindow struct {
	*widgets.QMainWindow
	textEdit       *widgets.QPlainTextEdit
	bufferAdapter  *BufferAdapter
	textBuffer     *textbuffer.TextBuffer
	filePath       string
	searchDialog   *widgets.QDialog
	searchLineEdit *widgets.QLineEdit
	statusBar      *widgets.QStatusBar
	caseSensitive  bool
	wholeWord      bool
	useRegex       bool
}

// NewEditorWindow 创建一个新的编辑器窗口
func NewEditorWindow(parent widgets.QWidget_ITF) *EditorWindow {
	// 创建主窗口
	window := widgets.NewQMainWindow(parent, 0)
	window.SetWindowTitle("GoTextBuffer Editor")
	window.Resize2(800, 600)

	// 创建编辑器窗口
	editor := &EditorWindow{
		QMainWindow:   window,
		textBuffer:    textbuffer.NewTextBuffer(),
		caseSensitive: false,
		wholeWord:     false,
		useRegex:      false,
	}

	// 设置UI
	editor.setupUI()

	// 设置快捷键
	editor.setupShortcuts()

	return editor
}

// setupUI 设置UI组件
func (e *EditorWindow) setupUI() {
	// 创建中央部件
	centralWidget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	centralWidget.SetLayout(layout)
	e.SetCentralWidget(centralWidget)

	// 创建菜单栏
	menuBar := e.createMenuBar()
	e.SetMenuBar(menuBar)

	// 创建工具栏
	toolbar := e.createToolbar()
	e.AddToolBar(core.Qt__TopToolBarArea, toolbar)

	// 创建文本编辑器
	e.textEdit = widgets.NewQPlainTextEdit(nil)
	e.textEdit.SetFont(gui.NewQFont2("Monospace", 10, int(gui.QFont__Normal), false))
	layout.AddWidget(e.textEdit, 0, 0)

	// 创建缓冲区适配器
	e.bufferAdapter = NewBufferAdapter(e.textEdit, e.textBuffer)

	// 创建状态栏
	e.statusBar = widgets.NewQStatusBar(nil)
	e.SetStatusBar(e.statusBar)

	// 更新状态栏
	e.updateStatusBar()

	// 连接光标位置变化信号
	e.textEdit.ConnectCursorPositionChanged(func() {
		e.updateStatusBar()
	})
}

// createMenuBar 创建菜单栏
func (e *EditorWindow) createMenuBar() *widgets.QMenuBar {
	menuBar := widgets.NewQMenuBar(nil)

	// 文件菜单
	fileMenu := menuBar.AddMenu2("文件")

	newAction := fileMenu.AddAction("新建")
	newAction.SetShortcut(gui.NewQKeySequence2("Ctrl+N", gui.QKeySequence__NativeText))
	newAction.ConnectTriggered(func(checked bool) {
		e.newFile()
	})

	openAction := fileMenu.AddAction("打开")
	openAction.SetShortcut(gui.NewQKeySequence2("Ctrl+O", gui.QKeySequence__NativeText))
	openAction.ConnectTriggered(func(checked bool) {
		e.openFile()
	})

	saveAction := fileMenu.AddAction("保存")
	saveAction.SetShortcut(gui.NewQKeySequence2("Ctrl+S", gui.QKeySequence__NativeText))
	saveAction.ConnectTriggered(func(checked bool) {
		e.saveFile()
	})

	saveAsAction := fileMenu.AddAction("另存为")
	saveAsAction.SetShortcut(gui.NewQKeySequence2("Ctrl+Shift+S", gui.QKeySequence__NativeText))
	saveAsAction.ConnectTriggered(func(checked bool) {
		e.saveFileAs()
	})

	fileMenu.AddSeparator()

	exitAction := fileMenu.AddAction("退出")
	exitAction.SetShortcut(gui.NewQKeySequence2("Ctrl+Q", gui.QKeySequence__NativeText))
	exitAction.ConnectTriggered(func(checked bool) {
		e.Close()
	})

	// 编辑菜单
	editMenu := menuBar.AddMenu2("编辑")

	undoAction := editMenu.AddAction("撤销")
	undoAction.SetShortcut(gui.NewQKeySequence2("Ctrl+Z", gui.QKeySequence__NativeText))
	undoAction.ConnectTriggered(func(checked bool) {
		e.undo()
	})

	redoAction := editMenu.AddAction("重做")
	redoAction.SetShortcut(gui.NewQKeySequence2("Ctrl+Y", gui.QKeySequence__NativeText))
	redoAction.ConnectTriggered(func(checked bool) {
		e.redo()
	})

	editMenu.AddSeparator()

	cutAction := editMenu.AddAction("剪切")
	cutAction.SetShortcut(gui.NewQKeySequence2("Ctrl+X", gui.QKeySequence__NativeText))
	cutAction.ConnectTriggered(func(checked bool) {
		e.textEdit.Cut()
	})

	copyAction := editMenu.AddAction("复制")
	copyAction.SetShortcut(gui.NewQKeySequence2("Ctrl+C", gui.QKeySequence__NativeText))
	copyAction.ConnectTriggered(func(checked bool) {
		e.textEdit.Copy()
	})

	pasteAction := editMenu.AddAction("粘贴")
	pasteAction.SetShortcut(gui.NewQKeySequence2("Ctrl+V", gui.QKeySequence__NativeText))
	pasteAction.ConnectTriggered(func(checked bool) {
		e.textEdit.Paste()
	})

	// 搜索菜单
	searchMenu := menuBar.AddMenu2("搜索")

	findAction := searchMenu.AddAction("查找")
	findAction.SetShortcut(gui.NewQKeySequence2("Ctrl+F", gui.QKeySequence__NativeText))
	findAction.ConnectTriggered(func(checked bool) {
		e.showSearchDialog()
	})

	return menuBar
}

// createToolbar 创建工具栏
func (e *EditorWindow) createToolbar() *widgets.QToolBar {
	toolbar := widgets.NewQToolBar2(nil)
	toolbar.SetMovable(false)

	newAction := toolbar.AddAction("新建")
	newAction.ConnectTriggered(func(checked bool) {
		e.newFile()
	})

	openAction := toolbar.AddAction("打开")
	openAction.ConnectTriggered(func(checked bool) {
		e.openFile()
	})

	saveAction := toolbar.AddAction("保存")
	saveAction.ConnectTriggered(func(checked bool) {
		e.saveFile()
	})

	toolbar.AddSeparator()

	undoAction := toolbar.AddAction("撤销")
	undoAction.ConnectTriggered(func(checked bool) {
		e.undo()
	})

	redoAction := toolbar.AddAction("重做")
	redoAction.ConnectTriggered(func(checked bool) {
		e.redo()
	})

	toolbar.AddSeparator()

	findAction := toolbar.AddAction("查找")
	findAction.ConnectTriggered(func(checked bool) {
		e.showSearchDialog()
	})

	return toolbar
}

// setupShortcuts 设置快捷键
func (e *EditorWindow) setupShortcuts() {
	// 快捷键已经在菜单项中设置
}

// updateStatusBar 更新状态栏
func (e *EditorWindow) updateStatusBar() {
	cursor := e.textEdit.TextCursor()
	line := cursor.BlockNumber() + 1
	column := cursor.PositionInBlock() + 1
	message := "行: " + QString(int(line)) + ", 列: " + QString(int(column))
	e.statusBar.ShowMessage(message, 0)
}

// QString 辅助函数，将整数转换为字符串
func QString(n int) string {
	return core.NewQVariant1(n).ToString()
}

// newFile 创建新文件
func (e *EditorWindow) newFile() {
	// 如果当前文档已修改，询问是否保存
	if e.textEdit.Document().IsModified() {
		result := widgets.QMessageBox_Question(nil, "保存文件",
			"文档已修改，是否保存？",
			widgets.QMessageBox__Save|widgets.QMessageBox__Discard|widgets.QMessageBox__Cancel,
			widgets.QMessageBox__Save)

		switch result {
		case widgets.QMessageBox__Save:
			if !e.saveFile() {
				return
			}
		case widgets.QMessageBox__Cancel:
			return
		}
	}

	// 清空文本编辑器
	e.textEdit.Clear()
	e.textBuffer.SetText("")
	e.filePath = ""
	e.SetWindowTitle("Orange Editor - 未命名")
}

// openFile 打开文件
func (e *EditorWindow) openFile(filePath ...string) {
	// 如果当前文档已修改，询问是否保存
	if e.textEdit.Document().IsModified() {
		result := widgets.QMessageBox_Question(nil, "保存文件",
			"文档已修改，是否保存？",
			widgets.QMessageBox__Save|widgets.QMessageBox__Discard|widgets.QMessageBox__Cancel,
			widgets.QMessageBox__Save)

		switch result {
		case widgets.QMessageBox__Save:
			if !e.saveFile() {
				return
			}
		case widgets.QMessageBox__Cancel:
			return
		}
	}

	// 如果没有提供文件路径，则打开文件对话框
	var selectedFilePath string
	if len(filePath) > 0 && filePath[0] != "" {
		selectedFilePath = filePath[0]
	} else {
		fileDialog := widgets.NewQFileDialog2(e, "打开文件", "", "")
		fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptOpen)
		fileDialog.SetFileMode(widgets.QFileDialog__ExistingFile)
		fileDialog.SetNameFilter("文本文件 (*.txt);;所有文件 (*)")
		if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
			return
		}
		selectedFiles := fileDialog.SelectedFiles()
		if len(selectedFiles) == 0 {
			return
		}
		selectedFilePath = selectedFiles[0]
	}

	// 打开文件
	file := core.NewQFile2(selectedFilePath)
	if !file.Open(core.QIODevice__ReadOnly | core.QIODevice__Text) {
		widgets.QMessageBox_Critical(nil, "错误", "无法打开文件: "+selectedFilePath, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	// 读取文件内容
	textStream := core.NewQTextStream2(file)
	byteArray := core.NewQByteArray2("UTF-8", -1)
	textStream.SetCodec(core.QTextCodec_CodecForName(byteArray))
	text := textStream.ReadAll()
	file.Close()

	// 设置文本内容
	e.textEdit.SetPlainText(text)
	e.textBuffer.SetText(text)
	e.filePath = selectedFilePath
	e.SetWindowTitle("Orange Editor - " + selectedFilePath)
}

// saveFile 保存文件
func (e *EditorWindow) saveFile() bool {
	if e.filePath == "" {
		return e.saveFileAs()
	}

	// 打开文件
	file := core.NewQFile2(e.filePath)
	if !file.Open(core.QIODevice__WriteOnly | core.QIODevice__Text) {
		widgets.QMessageBox_Critical(nil, "错误", "无法保存文件: "+e.filePath, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return false
	}

	// 获取文本并写入
	text := e.textEdit.ToPlainText()
	textBytes := []byte(text)
	file.Write(textBytes, int64(len(textBytes)))
	file.Close()

	// 标记文档为未修改
	e.textEdit.Document().SetModified(false)
	return true
}

// saveFileAs 另存为文件
func (e *EditorWindow) saveFileAs() bool {
	// 打开保存文件对话框
	filePath := widgets.QFileDialog_GetSaveFileName(
		nil, "保存文件", "",
		"文本文件 (*.txt);;所有文件 (*)", "", 0)

	if filePath == "" {
		return false
	}

	e.filePath = filePath
	e.SetWindowTitle("Orange Editor - " + filePath)
	return e.saveFile()
}

// undo 撤销操作
func (e *EditorWindow) undo() {
	e.textBuffer.Undo()
	e.bufferAdapter.updateTextEdit()
}

// redo 重做操作
func (e *EditorWindow) redo() {
	e.textBuffer.Redo()
	e.bufferAdapter.updateTextEdit()
}

// showSearchDialog 显示搜索对话框
func (e *EditorWindow) showSearchDialog() {
	if e.searchDialog == nil {
		e.searchDialog = widgets.NewQDialog(nil, 0)
		e.searchDialog.SetWindowTitle("查找")
		e.searchDialog.Resize2(400, 150)

		layout := widgets.NewQVBoxLayout()
		e.searchDialog.SetLayout(layout)

		// 搜索输入框
		formLayout := widgets.NewQFormLayout(nil)
		e.searchLineEdit = widgets.NewQLineEdit(nil)
		formLayout.AddRow3("查找内容:", e.searchLineEdit)
		layout.AddLayout(formLayout, 0)

		// 搜索选项
		optionsLayout := widgets.NewQHBoxLayout()

		caseSensitiveCheck := widgets.NewQCheckBox2("区分大小写", nil)
		caseSensitiveCheck.ConnectToggled(func(checked bool) {
			e.caseSensitive = checked
		})
		optionsLayout.AddWidget(caseSensitiveCheck, 0, 0)

		wholeWordCheck := widgets.NewQCheckBox2("全词匹配", nil)
		wholeWordCheck.ConnectToggled(func(checked bool) {
			e.wholeWord = checked
		})
		optionsLayout.AddWidget(wholeWordCheck, 0, 0)

		useRegexCheck := widgets.NewQCheckBox2("使用正则表达式", nil)
		useRegexCheck.ConnectToggled(func(checked bool) {
			e.useRegex = checked
		})
		optionsLayout.AddWidget(useRegexCheck, 0, 0)

		layout.AddLayout(optionsLayout, 0)

		// 按钮
		buttonLayout := widgets.NewQHBoxLayout()

		findNextButton := widgets.NewQPushButton2("查找下一个", nil)
		findNextButton.ConnectClicked(func(checked bool) {
			e.search(true)
		})
		buttonLayout.AddWidget(findNextButton, 0, 0)

		findPrevButton := widgets.NewQPushButton2("查找上一个", nil)
		findPrevButton.ConnectClicked(func(checked bool) {
			e.search(false)
		})
		buttonLayout.AddWidget(findPrevButton, 0, 0)

		closeButton := widgets.NewQPushButton2("关闭", nil)
		closeButton.ConnectClicked(func(checked bool) {
			e.searchDialog.Hide()
		})
		buttonLayout.AddWidget(closeButton, 0, 0)

		layout.AddLayout(buttonLayout, 0)
	}

	// 如果有选中的文本，设置为搜索内容
	cursor := e.textEdit.TextCursor()
	if cursor.HasSelection() {
		e.searchLineEdit.SetText(cursor.SelectedText())
	}

	e.searchDialog.Show()
}

// search 搜索文本
func (e *EditorWindow) search(forward bool) {
	searchText := e.searchLineEdit.Text()
	if searchText == "" {
		return
	}

	// 设置搜索选项
	options := gui.QTextDocument__FindFlag(0)
	if !forward {
		options |= gui.QTextDocument__FindBackward
	}
	if e.caseSensitive {
		options |= gui.QTextDocument__FindCaseSensitively
	}
	if e.wholeWord {
		options |= gui.QTextDocument__FindWholeWords
	}

	// 执行搜索
	found := false
	if e.useRegex {
		regex := core.NewQRegExp()
		regex.SetPattern(searchText)
		regex.SetCaseSensitivity(core.Qt__CaseSensitive)
		regex.SetPatternSyntax(core.QRegExp__RegExp)
		found = e.textEdit.Find2(regex, options)
	} else {
		found = e.textEdit.Find(searchText, options)
	}

	// 如果没有找到，且不是向后搜索，则从文档开始处继续搜索
	if !found && forward {
		cursor := e.textEdit.TextCursor()
		cursor.MovePosition(gui.QTextCursor__Start, gui.QTextCursor__MoveAnchor, 1)
		e.textEdit.SetTextCursor(cursor)

		if e.useRegex {
			regex := core.NewQRegExp()
			regex.SetPattern(searchText)
			regex.SetCaseSensitivity(core.Qt__CaseSensitive)
			regex.SetPatternSyntax(core.QRegExp__RegExp)
			found = e.textEdit.Find2(regex, options)
		} else {
			found = e.textEdit.Find(searchText, options)
		}
	}

	if !found {
		widgets.QMessageBox_Information(nil, "查找", "找不到 \""+searchText+"\"", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	}
}

// SetTextBuffer 设置文本缓冲区
func (e *EditorWindow) SetTextBuffer(buffer *textbuffer.TextBuffer) {
	e.textBuffer = buffer

	// 创建缓冲区适配器
	e.bufferAdapter = NewBufferAdapter(e.textEdit, buffer)
}

// OpenFile 打开文件
func (e *EditorWindow) OpenFile(filePath string) {
	e.openFile(filePath)
}
