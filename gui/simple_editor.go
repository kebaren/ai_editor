package main

import (
	"fmt"
	"log"
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/example/gotextbuffer/textbuffer"
)

// SimpleEditor 是一个简化版的文本编辑器
type SimpleEditor struct {
	window        *gtk.ApplicationWindow
	textView      *gtk.TextView
	gtkBuffer     *gtk.TextBuffer
	textBuffer    *textbuffer.TextBuffer
	bufferAdapter *BufferAdapter
	statusBar     *gtk.Statusbar
	contextID     uint
}

// BufferAdapter 处理GTK文本缓冲区和我们自定义的textbuffer之间的集成
type BufferAdapter struct {
	gtkBuffer  *gtk.TextBuffer
	textBuffer *textbuffer.TextBuffer
	updating   bool
}

// NewBufferAdapter 创建一个新的缓冲区适配器
func NewBufferAdapter(gtkBuffer *gtk.TextBuffer, textBuffer *textbuffer.TextBuffer) *BufferAdapter {
	adapter := &BufferAdapter{
		gtkBuffer:  gtkBuffer,
		textBuffer: textBuffer,
		updating:   false,
	}

	// 连接GTK缓冲区的changed信号
	gtkBuffer.ConnectChanged(func() {
		if !adapter.updating {
			adapter.onGtkBufferChanged()
		}
	})

	// 使用文本缓冲区的内容初始化GTK缓冲区
	adapter.updateGtkBuffer()

	return adapter
}

// onGtkBufferChanged 在GTK缓冲区更改时调用
func (a *BufferAdapter) onGtkBufferChanged() {
	// 从GTK缓冲区获取文本
	text := a.gtkBuffer.Text(a.gtkBuffer.StartIter(), a.gtkBuffer.EndIter(), false)

	// 更新文本缓冲区
	a.textBuffer.SetText(text)
}

// updateGtkBuffer 从文本缓冲区更新GTK缓冲区
func (a *BufferAdapter) updateGtkBuffer() {
	a.updating = true
	text := a.textBuffer.GetText()
	a.gtkBuffer.SetText(text)
	a.updating = false
}

// NewSimpleEditor 创建一个新的简单编辑器
func NewSimpleEditor(app *gtk.Application) *SimpleEditor {
	// 创建一个新窗口
	win := gtk.NewApplicationWindow(app)
	win.SetTitle("Simple TextBuffer Editor")
	win.SetDefaultSize(800, 600)

	// 创建文本缓冲区
	textBuffer := textbuffer.NewTextBuffer()

	// 创建GTK文本缓冲区和视图
	gtkBuffer := gtk.NewTextBuffer(nil)
	textView := gtk.NewTextViewWithBuffer(gtkBuffer)
	textView.SetWrapMode(gtk.WrapWord)
	textView.SetMonospace(true)

	// 创建缓冲区适配器
	bufferAdapter := NewBufferAdapter(gtkBuffer, textBuffer)

	// 创建状态栏
	statusBar := gtk.NewStatusbar()
	contextID := statusBar.GetContextId("editor")

	// 创建编辑器
	editor := &SimpleEditor{
		window:        win,
		textView:      textView,
		gtkBuffer:     gtkBuffer,
		textBuffer:    textBuffer,
		bufferAdapter: bufferAdapter,
		statusBar:     statusBar,
		contextID:     contextID,
	}

	// 设置UI
	editor.setupUI()

	// 设置快捷键
	editor.setupActions(app)

	return editor
}

// setupUI 设置编辑器的UI
func (e *SimpleEditor) setupUI() {
	// 创建滚动窗口
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetChild(e.textView)
	scrolledWindow.SetVExpand(true)
	scrolledWindow.SetHExpand(true)

	// 创建主布局
	mainBox := gtk.NewBox(gtk.OrientationVertical, 0)
	mainBox.Append(e.createMenuBar())
	mainBox.Append(scrolledWindow)
	mainBox.Append(e.statusBar)

	// 设置窗口内容
	e.window.SetChild(mainBox)

	// 更新状态栏
	e.updateStatusBar()

	// 连接光标位置变化信号
	e.gtkBuffer.ConnectMarkSet(func(iter *gtk.TextIter, mark *gtk.TextMark) {
		if mark.Name() == "insert" {
			e.updateStatusBar()
		}
	})
}

// createMenuBar 创建菜单栏
func (e *SimpleEditor) createMenuBar() *gtk.Box {
	// 创建菜单栏
	menuBar := gtk.NewBox(gtk.OrientationHorizontal, 0)
	menuBar.SetCSSClasses([]string{"toolbar"})

	// 文件菜单
	fileButton := gtk.NewButton()
	fileButton.SetLabel("文件")
	filePopover := gtk.NewPopover()
	filePopover.SetParent(fileButton)
	fileButton.ConnectClicked(func() {
		filePopover.Popup()
	})

	// 文件菜单项
	fileBox := gtk.NewBox(gtk.OrientationVertical, 0)

	newButton := gtk.NewButton()
	newButton.SetLabel("新建")
	newButton.ConnectClicked(func() {
		e.newFile()
		filePopover.Popdown()
	})
	fileBox.Append(newButton)

	openButton := gtk.NewButton()
	openButton.SetLabel("打开")
	openButton.ConnectClicked(func() {
		e.openFile()
		filePopover.Popdown()
	})
	fileBox.Append(openButton)

	saveButton := gtk.NewButton()
	saveButton.SetLabel("保存")
	saveButton.ConnectClicked(func() {
		e.saveFile()
		filePopover.Popdown()
	})
	fileBox.Append(saveButton)

	filePopover.SetChild(fileBox)
	menuBar.Append(fileButton)

	// 编辑菜单
	editButton := gtk.NewButton()
	editButton.SetLabel("编辑")
	editPopover := gtk.NewPopover()
	editPopover.SetParent(editButton)
	editButton.ConnectClicked(func() {
		editPopover.Popup()
	})

	// 编辑菜单项
	editBox := gtk.NewBox(gtk.OrientationVertical, 0)

	undoButton := gtk.NewButton()
	undoButton.SetLabel("撤销")
	undoButton.ConnectClicked(func() {
		e.undo()
		editPopover.Popdown()
	})
	editBox.Append(undoButton)

	redoButton := gtk.NewButton()
	redoButton.SetLabel("重做")
	redoButton.ConnectClicked(func() {
		e.redo()
		editPopover.Popdown()
	})
	editBox.Append(redoButton)

	editPopover.SetChild(editBox)
	menuBar.Append(editButton)

	return menuBar
}

// setupActions 设置编辑器的操作
func (e *SimpleEditor) setupActions(app *gtk.Application) {
	// 新建文件操作
	newAction := gio.NewSimpleAction("new", nil)
	newAction.ConnectActivate(func(parameter *gio.Variant) {
		e.newFile()
	})
	e.window.AddAction(newAction)

	// 打开文件操作
	openAction := gio.NewSimpleAction("open", nil)
	openAction.ConnectActivate(func(parameter *gio.Variant) {
		e.openFile()
	})
	e.window.AddAction(openAction)

	// 保存文件操作
	saveAction := gio.NewSimpleAction("save", nil)
	saveAction.ConnectActivate(func(parameter *gio.Variant) {
		e.saveFile()
	})
	e.window.AddAction(saveAction)

	// 撤销操作
	undoAction := gio.NewSimpleAction("undo", nil)
	undoAction.ConnectActivate(func(parameter *gio.Variant) {
		e.undo()
	})
	e.window.AddAction(undoAction)

	// 重做操作
	redoAction := gio.NewSimpleAction("redo", nil)
	redoAction.ConnectActivate(func(parameter *gio.Variant) {
		e.redo()
	})
	e.window.AddAction(redoAction)

	// 设置快捷键
	app.SetAccelsForAction("win.new", []string{"<Ctrl>n"})
	app.SetAccelsForAction("win.open", []string{"<Ctrl>o"})
	app.SetAccelsForAction("win.save", []string{"<Ctrl>s"})
	app.SetAccelsForAction("win.undo", []string{"<Ctrl>z"})
	app.SetAccelsForAction("win.redo", []string{"<Ctrl>y"})
}

// newFile 创建新文件
func (e *SimpleEditor) newFile() {
	e.textBuffer.Clear()
	e.bufferAdapter.updateGtkBuffer()
	e.updateStatusBar()
}

// openFile 打开文件
func (e *SimpleEditor) openFile() {
	dialog := gtk.NewFileChooserNative(
		"打开文件",
		e.window,
		gtk.FileChooserActionOpen,
		"打开",
		"取消",
	)

	dialog.ConnectResponse(func(responseID int) {
		if responseID == gtk.ResponseAccept {
			file := dialog.File()
			if file != nil {
				filename := file.Path()
				err := e.textBuffer.LoadFromFile(filename)
				if err != nil {
					e.showError("打开文件失败", err.Error())
				} else {
					e.bufferAdapter.updateGtkBuffer()
					e.updateStatusBar()
				}
			}
		}
		dialog.Destroy()
	})

	dialog.Show()
}

// saveFile 保存文件
func (e *SimpleEditor) saveFile() {
	dialog := gtk.NewFileChooserNative(
		"保存文件",
		e.window,
		gtk.FileChooserActionSave,
		"保存",
		"取消",
	)

	dialog.ConnectResponse(func(responseID int) {
		if responseID == gtk.ResponseAccept {
			file := dialog.File()
			if file != nil {
				filename := file.Path()
				err := e.textBuffer.SaveToFile(filename)
				if err != nil {
					e.showError("保存文件失败", err.Error())
				} else {
					e.updateStatusBar()
				}
			}
		}
		dialog.Destroy()
	})

	dialog.Show()
}

// undo 撤销操作
func (e *SimpleEditor) undo() {
	err := e.textBuffer.Undo()
	if err != nil {
		e.showError("撤销失败", err.Error())
	} else {
		e.bufferAdapter.updateGtkBuffer()
	}
}

// redo 重做操作
func (e *SimpleEditor) redo() {
	err := e.textBuffer.Redo()
	if err != nil {
		e.showError("重做失败", err.Error())
	} else {
		e.bufferAdapter.updateGtkBuffer()
	}
}

// updateStatusBar 更新状态栏
func (e *SimpleEditor) updateStatusBar() {
	mark := e.gtkBuffer.GetMark("insert")
	iter := e.gtkBuffer.GetIterAtMark(mark)
	line := iter.Line() + 1
	column := iter.LineOffset() + 1

	// 清除之前的消息
	e.statusBar.Remove(e.contextID, e.statusBar.RemoveAll(e.contextID))

	// 显示光标位置
	e.statusBar.Push(e.contextID, fmt.Sprintf("行: %d, 列: %d", line, column))
}

// showError 显示错误对话框
func (e *SimpleEditor) showError(title, message string) {
	dialog := gtk.NewMessageDialog(
		e.window,
		gtk.DialogFlagsModal,
		gtk.MessageTypeError,
		gtk.ButtonsTypeOk,
		title,
	)
	dialog.SetProperty("secondary-text", message)
	dialog.ConnectResponse(func(responseID int) {
		dialog.Destroy()
	})
	dialog.Show()
}

// Show 显示编辑器窗口
func (e *SimpleEditor) Show() {
	e.window.Show()
}

func main() {
	// 创建应用程序
	app := gtk.NewApplication("com.example.gotextbuffer.simple", 0)

	// 当应用程序被激活时创建窗口
	app.ConnectActivate(func() {
		// 创建一个新的简单编辑器
		editor := NewSimpleEditor(app)

		// 显示窗口
		editor.Show()
	})

	// 运行应用程序
	if code := app.Run(os.Args); code > 0 {
		log.Fatalf("Application exited with code %d", code)
	}
}
