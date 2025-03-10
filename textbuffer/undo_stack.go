package textbuffer

import (
	"errors"
)

// OperationType 表示文本操作的类型
type OperationType int

const (
	// OperationInsert 表示插入操作
	OperationInsert OperationType = iota
	// OperationDelete 表示删除操作
	OperationDelete
	// OperationReplace 表示替换操作
	OperationReplace
	// OperationClear 表示清空操作
	OperationClear
	// OperationSetText 表示设置整个文本内容的操作
	OperationSetText
)

// TextOperation 表示一个文本操作
type TextOperation struct {
	// 操作类型
	Type OperationType
	// 操作位置
	Position Position
	// 操作的文本
	Text string
	// 操作前的文本（用于撤销）
	OldText string
}

// UndoStack 是一个撤销/重做栈
type UndoStack struct {
	// 撤销栈
	undoStack []*TextOperation
	// 重做栈
	redoStack []*TextOperation
	// 最大栈大小
	maxStackSize int
}

// NewUndoStack 创建一个新的UndoStack
func NewUndoStack() *UndoStack {
	return &UndoStack{
		undoStack:    []*TextOperation{},
		redoStack:    []*TextOperation{},
		maxStackSize: 100, // 默认最大栈大小
	}
}

// Push 将一个操作推入撤销栈
func (us *UndoStack) Push(operation *TextOperation) {
	// 清空重做栈
	us.redoStack = []*TextOperation{}

	// 将操作推入撤销栈
	us.undoStack = append(us.undoStack, operation)

	// 如果撤销栈大小超过最大值，移除最早的操作
	if len(us.undoStack) > us.maxStackSize {
		us.undoStack = us.undoStack[1:]
	}
}

// Undo 撤销上一次操作
func (us *UndoStack) Undo() (*TextOperation, error) {
	if len(us.undoStack) == 0 {
		return nil, errors.New("no operation to undo")
	}

	// 弹出最后一个操作
	lastIndex := len(us.undoStack) - 1
	operation := us.undoStack[lastIndex]
	us.undoStack = us.undoStack[:lastIndex]

	// 将操作推入重做栈
	us.redoStack = append(us.redoStack, operation)

	return operation, nil
}

// Redo 重做上一次撤销的操作
func (us *UndoStack) Redo() (*TextOperation, error) {
	if len(us.redoStack) == 0 {
		return nil, errors.New("no operation to redo")
	}

	// 弹出最后一个操作
	lastIndex := len(us.redoStack) - 1
	operation := us.redoStack[lastIndex]
	us.redoStack = us.redoStack[:lastIndex]

	// 将操作推入撤销栈
	us.undoStack = append(us.undoStack, operation)

	return operation, nil
}

// CanUndo 判断是否可以撤销
func (us *UndoStack) CanUndo() bool {
	return len(us.undoStack) > 0
}

// CanRedo 判断是否可以重做
func (us *UndoStack) CanRedo() bool {
	return len(us.redoStack) > 0
}

// Clear 清空撤销/重做栈
func (us *UndoStack) Clear() {
	us.undoStack = []*TextOperation{}
	us.redoStack = []*TextOperation{}
}
