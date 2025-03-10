package editor

import (
	"github.com/example/gotextbuffer/textbuffer"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// BufferAdapter 处理Qt文本编辑器和我们自定义的TextBuffer之间的集成
type BufferAdapter struct {
	textEdit   *widgets.QPlainTextEdit
	textBuffer *textbuffer.TextBuffer
	updating   bool
}

// NewBufferAdapter 创建一个新的缓冲区适配器
func NewBufferAdapter(textEdit *widgets.QPlainTextEdit, textBuffer *textbuffer.TextBuffer) *BufferAdapter {
	adapter := &BufferAdapter{
		textEdit:   textEdit,
		textBuffer: textBuffer,
		updating:   false,
	}

	// 连接文本编辑器的文本变化信号
	textEdit.Document().ConnectContentsChanged(func() {
		if !adapter.updating {
			adapter.onTextEditChanged()
		}
	})

	// 初始化文本编辑器
	adapter.updateTextEdit()

	return adapter
}

// onTextEditChanged 当文本编辑器内容变化时调用
func (a *BufferAdapter) onTextEditChanged() {
	// 获取文本编辑器的文本
	text := a.textEdit.ToPlainText()

	// 更新文本缓冲区
	a.textBuffer.SetText(text)
}

// updateTextEdit 从文本缓冲区更新文本编辑器
func (a *BufferAdapter) updateTextEdit() {
	a.updating = true
	text := a.textBuffer.GetText()
	a.textEdit.SetPlainText(text)
	a.updating = false
}

// GetTextBuffer 获取文本缓冲区
func (a *BufferAdapter) GetTextBuffer() *textbuffer.TextBuffer {
	return a.textBuffer
}

// GetTextEdit 获取文本编辑器
func (a *BufferAdapter) GetTextEdit() *widgets.QPlainTextEdit {
	return a.textEdit
}

// SetCursorPosition 设置光标位置
func (a *BufferAdapter) SetCursorPosition(position textbuffer.Position) {
	cursor := a.textEdit.TextCursor()
	block := a.textEdit.Document().FindBlockByLineNumber(position.Line)
	if !block.IsValid() {
		return
	}

	pos := block.Position() + position.Column
	cursor.SetPosition(pos, gui.QTextCursor__MoveAnchor)
	a.textEdit.SetTextCursor(cursor)
}

// GetCursorPosition 获取光标位置
func (a *BufferAdapter) GetCursorPosition() textbuffer.Position {
	cursor := a.textEdit.TextCursor()
	line := cursor.BlockNumber()
	column := cursor.PositionInBlock()
	return textbuffer.Position{
		Line:   int(line),
		Column: int(column),
	}
}

// SelectRange 选择一个范围
func (a *BufferAdapter) SelectRange(r textbuffer.Range) {
	startBlock := a.textEdit.Document().FindBlockByLineNumber(r.Start.Line)
	endBlock := a.textEdit.Document().FindBlockByLineNumber(r.End.Line)

	if !startBlock.IsValid() || !endBlock.IsValid() {
		return
	}

	startPos := startBlock.Position() + r.Start.Column
	endPos := endBlock.Position() + r.End.Column

	cursor := a.textEdit.TextCursor()
	cursor.SetPosition(startPos, gui.QTextCursor__MoveAnchor)
	cursor.SetPosition(endPos, gui.QTextCursor__KeepAnchor)
	a.textEdit.SetTextCursor(cursor)
}

// GetSelection 获取当前选择
func (a *BufferAdapter) GetSelection() (textbuffer.Range, bool) {
	cursor := a.textEdit.TextCursor()
	if !cursor.HasSelection() {
		return textbuffer.Range{}, false
	}

	startPos := cursor.SelectionStart()
	endPos := cursor.SelectionEnd()

	startBlock := a.textEdit.Document().FindBlock(startPos)
	endBlock := a.textEdit.Document().FindBlock(endPos)

	startLine := startBlock.BlockNumber()
	startColumn := startPos - startBlock.Position()

	endLine := endBlock.BlockNumber()
	endColumn := endPos - endBlock.Position()

	return textbuffer.Range{
		Start: textbuffer.Position{
			Line:   int(startLine),
			Column: int(startColumn),
		},
		End: textbuffer.Position{
			Line:   int(endLine),
			Column: int(endColumn),
		},
	}, true
}
