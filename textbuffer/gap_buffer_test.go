package textbuffer

import (
	"strings"
	"testing"
)

func TestGapBuffer(t *testing.T) {
	// 测试创建GapBuffer
	buffer := NewGapBuffer()
	if buffer.GetText() != "" {
		t.Errorf("Expected empty text, got '%s'", buffer.GetText())
	}

	// 测试插入文本
	buffer.Insert(0, "Hello, World!")
	if buffer.GetText() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", buffer.GetText())
	}

	// 测试在中间插入文本
	buffer.Insert(7, " Go")
	expectedText := "Hello,  GoWorld!"
	if buffer.GetText() != expectedText {
		t.Errorf("Expected '%s', got '%s'", expectedText, buffer.GetText())
	}

	// 测试删除文本
	buffer.Delete(7, 10)
	if buffer.GetText() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", buffer.GetText())
	}

	// 测试获取行数
	if buffer.GetLineCount() != 1 {
		t.Errorf("Expected 1 line, got %d", buffer.GetLineCount())
	}

	// 测试插入包含换行符的文本
	buffer.Insert(13, "\nSecond line\nThird line")
	expectedText = "Hello, World!\nSecond line\nThird line"
	if buffer.GetText() != expectedText {
		t.Errorf("Expected '%s', got '%s'", expectedText, buffer.GetText())
	}

	// 测试获取行数
	if buffer.GetLineCount() != 3 {
		t.Errorf("Expected 3 lines, got %d", buffer.GetLineCount())
	}

	// 测试获取指定行的内容
	if buffer.GetLineContent(1) != "Second line\n" {
		t.Errorf("Expected 'Second line\\n', got '%s'", buffer.GetLineContent(1))
	}

	// 测试获取位置
	position := buffer.GetPositionAt(15)
	if position.Line != 1 || position.Column != 1 {
		t.Errorf("Expected position (1, 1), got (%d, %d)", position.Line, position.Column)
	}

	// 测试获取偏移量
	offset := buffer.GetOffsetAt(Position{Line: 1, Column: 1})
	if offset != 15 {
		t.Errorf("Expected offset 15, got %d", offset)
	}

	// 测试跨行删除
	buffer.Delete(12, 20)
	expectedText = "Hello, World line\nThird line"
	if buffer.GetText() != expectedText {
		t.Errorf("Expected '%s', got '%s'", expectedText, buffer.GetText())
	}

	// 测试清空缓冲区
	buffer.Clear()
	if buffer.GetText() != "" {
		t.Errorf("Expected empty text after clear, got '%s'", buffer.GetText())
	}
}

func TestGapBufferWithLineBreaks(t *testing.T) {
	// 测试创建带有多行文本的GapBuffer
	buffer := NewGapBufferWithText("Line 1\nLine 2\nLine 3")

	// 测试获取行数
	if buffer.GetLineCount() != 3 {
		t.Errorf("Expected 3 lines, got %d", buffer.GetLineCount())
	}

	// 测试获取指定行的内容
	if buffer.GetLineContent(0) != "Line 1\n" {
		t.Errorf("Expected 'Line 1\\n', got '%s'", buffer.GetLineContent(0))
	}

	if buffer.GetLineContent(1) != "Line 2\n" {
		t.Errorf("Expected 'Line 2\\n', got '%s'", buffer.GetLineContent(1))
	}

	if buffer.GetLineContent(2) != "Line 3" {
		t.Errorf("Expected 'Line 3', got '%s'", buffer.GetLineContent(2))
	}

	// 测试在行尾插入文本
	buffer.Insert(6, " modified")
	if buffer.GetText() != "Line 1 modified\nLine 2\nLine 3" {
		t.Errorf("Expected 'Line 1 modified\\nLine 2\\nLine 3', got '%s'", buffer.GetText())
	}

	// 测试在行首插入文本
	buffer.Insert(16, "Modified ")
	if buffer.GetText() != "Line 1 modified\nModified Line 2\nLine 3" {
		t.Errorf("Expected 'Line 1 modified\\nModified Line 2\\nLine 3', got '%s'", buffer.GetText())
	}

	// 测试跨行删除
	buffer.Delete(5, 20)
	expectedText := "Line fied Line 2\nLine 3"
	if buffer.GetText() != expectedText {
		t.Errorf("Expected '%s', got '%s'", expectedText, buffer.GetText())
	}

	// 测试插入多个换行符
	buffer.Insert(4, "\n\n")
	expectedText = "Line\n\n fied Line 2\nLine 3"
	if buffer.GetText() != expectedText {
		t.Errorf("Expected '%s', got '%s'", expectedText, buffer.GetText())
	}

	// 测试获取行数
	if buffer.GetLineCount() != 4 {
		t.Errorf("Expected 4 lines, got %d", buffer.GetLineCount())
	}
}

func TestGapBufferWithLargeText(t *testing.T) {
	// 测试创建带有大量文本的GapBuffer
	buffer := NewGapBuffer()

	// 插入100行文本
	for i := 0; i < 100; i++ {
		buffer.Insert(buffer.GetLength(), "Line "+string(rune(i+48))+"\n")
	}

	// 测试获取行数
	if buffer.GetLineCount() != 100 {
		t.Errorf("Expected 100 lines, got %d", buffer.GetLineCount())
	}

	// 测试随机访问行
	content := buffer.GetLineContent(50)
	if content == "" {
		t.Errorf("Expected non-empty content for line 50")
	}

	// 测试随机插入
	buffer.Insert(50, "Inserted Text")

	// 测试随机删除
	buffer.Delete(20, 30)

	// 测试间隙移动
	buffer.Insert(0, "Start")
	buffer.Insert(buffer.GetLength(), "End")

	// 测试大量插入导致的缓冲区扩容
	largeText := strings.Repeat("Large Text ", 1000)
	buffer.Insert(buffer.GetLength()/2, largeText)
}
