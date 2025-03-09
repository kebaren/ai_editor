package textbuffer

import (
	"testing"
)

func TestTextBuffer(t *testing.T) {
	// 测试创建TextBuffer
	buffer := NewTextBuffer()
	if buffer.GetText() != "" {
		t.Errorf("Expected empty text, got '%s'", buffer.GetText())
	}

	// 测试插入文本
	err := buffer.Insert(Position{Line: 0, Column: 0}, "Hello, World!")
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}
	if buffer.GetText() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", buffer.GetText())
	}

	// 测试在中间插入文本
	err = buffer.Insert(Position{Line: 0, Column: 7}, " Go")
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}
	if buffer.GetText() != "Hello,  GoWorld!" {
		t.Errorf("Expected 'Hello,  GoWorld!', got '%s'", buffer.GetText())
	}

	// 测试删除文本
	err = buffer.Delete(Range{
		Start: Position{Line: 0, Column: 7},
		End:   Position{Line: 0, Column: 10},
	})
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	if buffer.GetText() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", buffer.GetText())
	}

	// 测试撤销操作
	err = buffer.Undo()
	if err != nil {
		t.Errorf("Undo failed: %v", err)
	}
	if buffer.GetText() != "Hello,  GoWorld!" {
		t.Errorf("Expected 'Hello,  GoWorld!', got '%s'", buffer.GetText())
	}

	// 测试重做操作
	err = buffer.Redo()
	if err != nil {
		t.Errorf("Redo failed: %v", err)
	}
	if buffer.GetText() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", buffer.GetText())
	}

	// 测试替换文本
	err = buffer.Replace(Range{
		Start: Position{Line: 0, Column: 0},
		End:   Position{Line: 0, Column: 5},
	}, "Hi")
	if err != nil {
		t.Errorf("Replace failed: %v", err)
	}
	if buffer.GetText() != "Hi, World!" {
		t.Errorf("Expected 'Hi, World!', got '%s'", buffer.GetText())
	}

	// 测试获取行数
	if buffer.GetLineCount() != 1 {
		t.Errorf("Expected 1 line, got %d", buffer.GetLineCount())
	}

	// 测试获取文本长度
	if buffer.GetLength() != 10 {
		t.Errorf("Expected length 10, got %d", buffer.GetLength())
	}

	// 测试插入多行文本
	err = buffer.Insert(Position{Line: 0, Column: buffer.GetLength()}, "\nThis is a new line.\nAnd another line.")
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}
	expectedText := "Hi, World!\nThis is a new line.\nAnd another line."
	if buffer.GetText() != expectedText {
		t.Errorf("Expected '%s', got '%s'", expectedText, buffer.GetText())
	}

	// 测试获取行数
	if buffer.GetLineCount() != 3 {
		t.Errorf("Expected 3 lines, got %d", buffer.GetLineCount())
	}

	// 测试获取指定行的内容
	if buffer.GetLineContent(1) != "This is a new line.\n" {
		t.Errorf("Expected 'This is a new line.\\n', got '%s'", buffer.GetLineContent(1))
	}

	// 测试获取所有行的内容
	lines := buffer.GetLines()
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "Hi, World!\n" {
		t.Errorf("Expected 'Hi, World!\\n', got '%s'", lines[0])
	}
	if lines[1] != "This is a new line.\n" {
		t.Errorf("Expected 'This is a new line.\\n', got '%s'", lines[1])
	}
	if lines[2] != "And another line." {
		t.Errorf("Expected 'And another line.', got '%s'", lines[2])
	}

	// 测试清空文本
	buffer.Clear()
	if buffer.GetText() != "" {
		t.Errorf("Expected empty text after clear, got '%s'", buffer.GetText())
	}

	// 测试设置文本
	buffer.SetText("New text")
	if buffer.GetText() != "New text" {
		t.Errorf("Expected 'New text', got '%s'", buffer.GetText())
	}
}

func TestTextBufferWithInitialText(t *testing.T) {
	// 测试创建带有初始文本的TextBuffer
	buffer := NewTextBufferWithText("Initial text")
	if buffer.GetText() != "Initial text" {
		t.Errorf("Expected 'Initial text', got '%s'", buffer.GetText())
	}

	// 测试获取位置
	position := buffer.GetPositionAt(8)
	if position.Line != 0 || position.Column != 8 {
		t.Errorf("Expected position (0, 8), got (%d, %d)", position.Line, position.Column)
	}

	// 测试获取偏移量
	offset := buffer.GetOffsetAt(Position{Line: 0, Column: 8})
	if offset != 8 {
		t.Errorf("Expected offset 8, got %d", offset)
	}

	// 测试获取范围内的文本
	text := buffer.GetTextInRange(Range{
		Start: Position{Line: 0, Column: 0},
		End:   Position{Line: 0, Column: 7},
	})
	if text != "Initial" {
		t.Errorf("Expected 'Initial', got '%s'", text)
	}
}

func TestTextBufferWithMultilineText(t *testing.T) {
	// 测试创建带有多行文本的TextBuffer
	initialText := "Line 1\nLine 2\nLine 3"
	buffer := NewTextBufferWithText(initialText)
	if buffer.GetText() != initialText {
		t.Errorf("Expected '%s', got '%s'", initialText, buffer.GetText())
	}

	// 测试获取行数
	if buffer.GetLineCount() != 3 {
		t.Errorf("Expected 3 lines, got %d", buffer.GetLineCount())
	}

	// 测试获取指定行的内容
	if buffer.GetLineContent(1) != "Line 2\n" {
		t.Errorf("Expected 'Line 2\\n', got '%s'", buffer.GetLineContent(1))
	}

	// 测试获取位置
	position := buffer.GetPositionAt(8)
	if position.Line != 1 || position.Column != 1 {
		t.Errorf("Expected position (1, 1), got (%d, %d)", position.Line, position.Column)
	}

	// 测试获取偏移量
	offset := buffer.GetOffsetAt(Position{Line: 1, Column: 1})
	if offset != 8 {
		t.Errorf("Expected offset 8, got %d", offset)
	}

	// 测试跨行删除
	err := buffer.Delete(Range{
		Start: Position{Line: 0, Column: 6},
		End:   Position{Line: 1, Column: 3},
	})
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	if buffer.GetText() != "Line 1e 2\nLine 3" {
		t.Errorf("Expected 'Line 1e 2\\nLine 3', got '%s'", buffer.GetText())
	}

	// 测试跨行插入
	err = buffer.Insert(Position{Line: 1, Column: 0}, "New line\n")
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}
	if buffer.GetText() != "Line 1e 2\nNew line\nLine 3" {
		t.Errorf("Expected 'Line 1e 2\\nNew line\\nLine 3', got '%s'", buffer.GetText())
	}
}
