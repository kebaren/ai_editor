package textbuffer

import (
	"errors"
	"sync"
)

// TextBuffer 是一个文本缓冲区，用于存储和操作文本
type TextBuffer struct {
	// 文本内容的数据结构（使用红黑树替代PieceTable）
	rbTree *RBTree
	// 互斥锁，用于并发访问
	mutex sync.RWMutex
	// 撤销/重做栈
	undoStack *UndoStack
}

// NewTextBuffer 创建一个新的TextBuffer
func NewTextBuffer() *TextBuffer {
	return NewTextBufferWithText("")
}

// NewTextBufferWithText 创建一个新的TextBuffer，并初始化文本内容
func NewTextBufferWithText(text string) *TextBuffer {
	rbTree := NewRBTree()
	if text != "" {
		rbTree.Insert(0, text)
	}

	return &TextBuffer{
		rbTree:    rbTree,
		mutex:     sync.RWMutex{},
		undoStack: NewUndoStack(),
	}
}

// GetText 获取整个文本内容
func (tb *TextBuffer) GetText() string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.rbTree.GetText()
}

// GetLength 获取文本总长度
func (tb *TextBuffer) GetLength() int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	if tb.rbTree.GetRoot() == nil {
		return 0
	}

	return tb.rbTree.GetRoot().totalLength
}

// GetLineCount 获取行数
func (tb *TextBuffer) GetLineCount() int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.rbTree.GetLineCount()
}

// GetLineContent 获取指定行的内容
func (tb *TextBuffer) GetLineContent(lineIndex int) string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.rbTree.GetLineContent(lineIndex)
}

// GetLines 获取所有行的内容
func (tb *TextBuffer) GetLines() []string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.rbTree.GetLines()
}

// GetPositionAt 获取指定偏移量对应的位置
func (tb *TextBuffer) GetPositionAt(offset int) Position {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.rbTree.GetPositionAt(offset)
}

// GetOffsetAt 获取指定位置对应的偏移量
func (tb *TextBuffer) GetOffsetAt(position Position) int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.rbTree.GetOffsetAt(position)
}

// GetTextInRange 获取指定范围内的文本
func (tb *TextBuffer) GetTextInRange(r Range) string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	startOffset := tb.rbTree.GetOffsetAt(r.Start)
	endOffset := tb.rbTree.GetOffsetAt(r.End)

	if startOffset >= endOffset {
		return ""
	}

	// 找到起始位置
	startNode, startNodeOffset := tb.rbTree.findNodeAtOffset(startOffset)
	if startNode == nil {
		return ""
	}

	// 找到结束位置
	endNode, endNodeOffset := tb.rbTree.findNodeAtOffset(endOffset)
	if endNode == nil {
		return ""
	}

	// 如果起始和结束位置在同一个节点内
	if startNode == endNode {
		runes := []rune(startNode.line)
		return string(runes[startNodeOffset:endNodeOffset])
	}

	// 如果起始和结束位置在不同的节点内
	var result string

	// 1. 处理起始节点
	startRunes := []rune(startNode.line)
	result += string(startRunes[startNodeOffset:])

	// 2. 处理中间节点
	inRange := false
	tb.rbTree.inOrderTraversal(tb.rbTree.root, func(node *TreeNode) {
		if node == startNode {
			inRange = true
			return
		}

		if inRange && node != endNode {
			result += node.line
		}

		if node == endNode {
			inRange = false
		}
	})

	// 3. 处理结束节点
	endRunes := []rune(endNode.line)
	result += string(endRunes[:endNodeOffset])

	return result
}

// Insert 在指定位置插入文本
func (tb *TextBuffer) Insert(position Position, text string) error {
	if text == "" {
		return nil
	}

	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	offset := tb.rbTree.GetOffsetAt(position)

	// 记录操作用于撤销
	tb.undoStack.Push(&TextOperation{
		Type:     OperationInsert,
		Position: position,
		Text:     text,
		OldText:  "",
	})

	// 执行插入操作
	tb.rbTree.Insert(offset, text)

	return nil
}

// Delete 删除指定范围的文本
func (tb *TextBuffer) Delete(r Range) error {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	startOffset := tb.rbTree.GetOffsetAt(r.Start)
	endOffset := tb.rbTree.GetOffsetAt(r.End)

	if startOffset >= endOffset {
		return errors.New("invalid range")
	}

	// 获取要删除的文本
	oldText := tb.GetTextInRange(r)

	// 记录操作用于撤销
	tb.undoStack.Push(&TextOperation{
		Type:     OperationDelete,
		Position: r.Start,
		Text:     "",
		OldText:  oldText,
	})

	// 执行删除操作
	tb.rbTree.Delete(startOffset, endOffset)

	return nil
}

// Replace 替换指定范围的文本
func (tb *TextBuffer) Replace(r Range, text string) error {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	startOffset := tb.rbTree.GetOffsetAt(r.Start)
	endOffset := tb.rbTree.GetOffsetAt(r.End)

	if startOffset > endOffset {
		return errors.New("invalid range")
	}

	// 获取要替换的文本
	oldText := tb.GetTextInRange(r)

	// 记录操作用于撤销
	tb.undoStack.Push(&TextOperation{
		Type:     OperationReplace,
		Position: r.Start,
		Text:     text,
		OldText:  oldText,
	})

	// 执行替换操作
	tb.rbTree.Delete(startOffset, endOffset)
	tb.rbTree.Insert(startOffset, text)

	return nil
}

// Undo 撤销上一次操作
func (tb *TextBuffer) Undo() error {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	operation, err := tb.undoStack.Undo()
	if err != nil {
		return err
	}

	switch operation.Type {
	case OperationInsert:
		// 撤销插入操作，需要删除插入的文本
		startOffset := tb.rbTree.GetOffsetAt(operation.Position)
		endOffset := startOffset + len([]rune(operation.Text))
		tb.rbTree.Delete(startOffset, endOffset)
	case OperationDelete:
		// 撤销删除操作，需要重新插入删除的文本
		offset := tb.rbTree.GetOffsetAt(operation.Position)
		tb.rbTree.Insert(offset, operation.OldText)
	case OperationReplace:
		// 撤销替换操作，需要恢复原来的文本
		startOffset := tb.rbTree.GetOffsetAt(operation.Position)
		endOffset := startOffset + len([]rune(operation.Text))
		tb.rbTree.Delete(startOffset, endOffset)
		tb.rbTree.Insert(startOffset, operation.OldText)
	}

	return nil
}

// Redo 重做上一次撤销的操作
func (tb *TextBuffer) Redo() error {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	operation, err := tb.undoStack.Redo()
	if err != nil {
		return err
	}

	switch operation.Type {
	case OperationInsert:
		// 重做插入操作
		offset := tb.rbTree.GetOffsetAt(operation.Position)
		tb.rbTree.Insert(offset, operation.Text)
	case OperationDelete:
		// 重做删除操作
		startOffset := tb.rbTree.GetOffsetAt(operation.Position)
		endOffset := startOffset + len([]rune(operation.OldText))
		tb.rbTree.Delete(startOffset, endOffset)
	case OperationReplace:
		// 重做替换操作
		startOffset := tb.rbTree.GetOffsetAt(operation.Position)
		endOffset := startOffset + len([]rune(operation.OldText))
		tb.rbTree.Delete(startOffset, endOffset)
		tb.rbTree.Insert(startOffset, operation.Text)
	}

	return nil
}

// Clear 清空文本缓冲区
func (tb *TextBuffer) Clear() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// 记录操作用于撤销
	tb.undoStack.Push(&TextOperation{
		Type:     OperationDelete,
		Position: Position{Line: 0, Column: 0},
		Text:     "",
		OldText:  tb.rbTree.GetText(),
	})

	// 清空红黑树
	tb.rbTree.Clear()
}

// SetText 设置整个文本内容
func (tb *TextBuffer) SetText(text string) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// 记录操作用于撤销
	tb.undoStack.Push(&TextOperation{
		Type:     OperationReplace,
		Position: Position{Line: 0, Column: 0},
		Text:     text,
		OldText:  tb.rbTree.GetText(),
	})

	// 清空红黑树
	tb.rbTree.Clear()

	// 插入新文本
	if text != "" {
		tb.rbTree.Insert(0, text)
	}
}
