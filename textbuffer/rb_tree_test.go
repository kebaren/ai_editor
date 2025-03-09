package textbuffer

import (
	"testing"
)

func TestRBTree(t *testing.T) {
	// 测试创建RBTree
	tree := NewRBTree()
	if !tree.IsEmpty() {
		t.Errorf("Expected empty tree, got size %d", tree.Size())
	}

	// 测试插入文本
	tree.Insert(0, "Hello, World!")
	if tree.GetText() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", tree.GetText())
	}

	// 测试在中间插入文本
	tree.Insert(7, " Go")
	if tree.GetText() != "Hello, Go World!" {
		t.Errorf("Expected 'Hello, Go World!', got '%s'", tree.GetText())
	}

	// 测试删除文本
	tree.Delete(7, 10)
	if tree.GetText() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", tree.GetText())
	}

	// 测试获取行数
	if tree.GetLineCount() != 1 {
		t.Errorf("Expected 1 line, got %d", tree.GetLineCount())
	}

	// 测试插入包含换行符的文本
	tree.Insert(13, "\nSecond line\nThird line")
	if tree.GetText() != "Hello, World!\nSecond line\nThird line" {
		t.Errorf("Expected 'Hello, World!\\nSecond line\\nThird line', got '%s'", tree.GetText())
	}

	// 测试获取行数
	if tree.GetLineCount() != 3 {
		t.Errorf("Expected 3 lines, got %d", tree.GetLineCount())
	}

	// 测试获取指定行的内容
	if tree.GetLineContent(1) != "Second line\n" {
		t.Errorf("Expected 'Second line\\n', got '%s'", tree.GetLineContent(1))
	}

	// 测试获取位置
	position := tree.GetPositionAt(15)
	if position.Line != 1 || position.Column != 1 {
		t.Errorf("Expected position (1, 1), got (%d, %d)", position.Line, position.Column)
	}

	// 测试获取偏移量
	offset := tree.GetOffsetAt(Position{Line: 1, Column: 1})
	if offset != 15 {
		t.Errorf("Expected offset 15, got %d", offset)
	}

	// 测试跨行删除
	tree.Delete(12, 20)
	if tree.GetText() != "Hello, World!ond line\nThird line" {
		t.Errorf("Expected 'Hello, World!ond line\\nThird line', got '%s'", tree.GetText())
	}

	// 测试清空树
	tree.Clear()
	if !tree.IsEmpty() {
		t.Errorf("Expected empty tree after clear, got size %d", tree.Size())
	}
}

func TestRBTreeWithLineBreaks(t *testing.T) {
	// 测试创建带有多行文本的RBTree
	tree := NewRBTree()
	tree.Insert(0, "Line 1\nLine 2\nLine 3")

	// 测试获取行数
	if tree.GetLineCount() != 3 {
		t.Errorf("Expected 3 lines, got %d", tree.GetLineCount())
	}

	// 测试获取指定行的内容
	if tree.GetLineContent(0) != "Line 1\n" {
		t.Errorf("Expected 'Line 1\\n', got '%s'", tree.GetLineContent(0))
	}

	if tree.GetLineContent(1) != "Line 2\n" {
		t.Errorf("Expected 'Line 2\\n', got '%s'", tree.GetLineContent(1))
	}

	if tree.GetLineContent(2) != "Line 3" {
		t.Errorf("Expected 'Line 3', got '%s'", tree.GetLineContent(2))
	}

	// 测试在行尾插入文本
	tree.Insert(6, " modified")
	if tree.GetText() != "Line 1 modified\nLine 2\nLine 3" {
		t.Errorf("Expected 'Line 1 modified\\nLine 2\\nLine 3', got '%s'", tree.GetText())
	}

	// 测试在行首插入文本
	tree.Insert(16, "Modified ")
	if tree.GetText() != "Line 1 modified\nModified Line 2\nLine 3" {
		t.Errorf("Expected 'Line 1 modified\\nModified Line 2\\nLine 3', got '%s'", tree.GetText())
	}

	// 测试跨行删除
	tree.Delete(5, 20)
	if tree.GetText() != "Line Line 2\nLine 3" {
		t.Errorf("Expected 'Line Line 2\\nLine 3', got '%s'", tree.GetText())
	}

	// 测试插入多个换行符
	tree.Insert(4, "\n\n")
	if tree.GetText() != "Line\n\n Line 2\nLine 3" {
		t.Errorf("Expected 'Line\\n\\n Line 2\\nLine 3', got '%s'", tree.GetText())
	}

	// 测试获取行数
	if tree.GetLineCount() != 4 {
		t.Errorf("Expected 4 lines, got %d", tree.GetLineCount())
	}
}

func TestRBTreeWithLargeText(t *testing.T) {
	// 测试创建带有大量文本的RBTree
	tree := NewRBTree()

	// 插入1000行文本
	for i := 0; i < 1000; i++ {
		tree.Insert(tree.GetRoot().totalLength, "Line "+string(rune(i+48))+"\n")
	}

	// 测试获取行数
	if tree.GetLineCount() != 1000 {
		t.Errorf("Expected 1000 lines, got %d", tree.GetLineCount())
	}

	// 测试随机访问行
	if tree.GetLineContent(500) != "Line "+string(rune(500+48))+"\n" {
		t.Errorf("Expected 'Line %c\\n', got '%s'", rune(500+48), tree.GetLineContent(500))
	}

	// 测试随机插入
	tree.Insert(2500, "Inserted Text")

	// 测试随机删除
	tree.Delete(2000, 3000)
}
