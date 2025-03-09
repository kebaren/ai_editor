package textbuffer

import (
	"strings"
)

// LineMap 用于管理文本的行信息
type LineMap struct {
	// 行结束位置的偏移量数组
	lineEndings []int
	// 文本总长度
	length int
}

// NewLineMap 创建一个新的LineMap
func NewLineMap(text string) *LineMap {
	lm := &LineMap{
		lineEndings: []int{},
		length:      len([]rune(text)),
	}
	lm.update(text)
	return lm
}

// update 更新行信息
func (lm *LineMap) update(text string) {
	runes := []rune(text)
	lm.lineEndings = []int{}
	lm.length = len(runes)

	for i, r := range runes {
		if r == '\n' {
			lm.lineEndings = append(lm.lineEndings, i)
		}
	}
	// 添加最后一行的结束位置
	if len(runes) > 0 {
		lm.lineEndings = append(lm.lineEndings, len(runes)-1)
	}
}

// GetLineCount 获取行数
func (lm *LineMap) GetLineCount() int {
	return len(lm.lineEndings)
}

// GetLineLength 获取指定行的长度
func (lm *LineMap) GetLineLength(lineIndex int) int {
	if lineIndex < 0 || lineIndex >= len(lm.lineEndings) {
		return 0
	}

	if lineIndex == 0 {
		return lm.lineEndings[0] + 1
	}

	return lm.lineEndings[lineIndex] - lm.lineEndings[lineIndex-1]
}

// GetLineContent 获取指定行的内容
func (lm *LineMap) GetLineContent(text string, lineIndex int) string {
	if lineIndex < 0 || lineIndex >= len(lm.lineEndings) {
		return ""
	}

	runes := []rune(text)
	var startOffset int
	if lineIndex == 0 {
		startOffset = 0
	} else {
		startOffset = lm.lineEndings[lineIndex-1] + 1
	}

	endOffset := lm.lineEndings[lineIndex] + 1
	if endOffset > len(runes) {
		endOffset = len(runes)
	}

	return string(runes[startOffset:endOffset])
}

// GetPositionAt 获取指定偏移量对应的位置
func (lm *LineMap) GetPositionAt(offset int) Position {
	if offset <= 0 {
		return Position{Line: 0, Column: 0}
	}

	if offset >= lm.length {
		// 如果偏移量超出文本长度，返回最后一个位置
		if len(lm.lineEndings) == 0 {
			return Position{Line: 0, Column: lm.length}
		}
		lastLine := len(lm.lineEndings) - 1
		lastColumn := lm.lineEndings[lastLine]
		if lastLine > 0 {
			lastColumn -= lm.lineEndings[lastLine-1] + 1
		}
		return Position{Line: lastLine, Column: lastColumn}
	}

	// 二分查找找到行号
	line := 0
	low, high := 0, len(lm.lineEndings)-1
	for low <= high {
		mid := (low + high) / 2
		if lm.lineEndings[mid] < offset {
			low = mid + 1
		} else if lm.lineEndings[mid] > offset {
			high = mid - 1
		} else {
			line = mid
			break
		}
	}

	if high < 0 {
		line = 0
	} else if low > high {
		line = low
	}

	// 计算列号
	var column int
	if line == 0 {
		column = offset
	} else {
		column = offset - lm.lineEndings[line-1] - 1
	}

	return Position{Line: line, Column: column}
}

// GetOffsetAt 获取指定位置对应的偏移量
func (lm *LineMap) GetOffsetAt(position Position) int {
	if position.Line < 0 {
		return 0
	}

	if position.Line >= len(lm.lineEndings) {
		return lm.length
	}

	var offset int
	if position.Line == 0 {
		offset = 0
	} else {
		offset = lm.lineEndings[position.Line-1] + 1
	}

	lineLength := lm.GetLineLength(position.Line)
	if position.Column >= lineLength {
		offset += lineLength
	} else {
		offset += position.Column
	}

	return offset
}

// GetTextInRange 获取指定范围内的文本
func (lm *LineMap) GetTextInRange(text string, r Range) string {
	startOffset := lm.GetOffsetAt(r.Start)
	endOffset := lm.GetOffsetAt(r.End)
	runes := []rune(text)

	if startOffset >= len(runes) || endOffset <= 0 || startOffset >= endOffset {
		return ""
	}

	if endOffset > len(runes) {
		endOffset = len(runes)
	}

	return string(runes[startOffset:endOffset])
}

// GetLines 获取所有行的内容
func (lm *LineMap) GetLines(text string) []string {
	lines := make([]string, lm.GetLineCount())
	for i := 0; i < lm.GetLineCount(); i++ {
		lines[i] = lm.GetLineContent(text, i)
	}
	return lines
}

// GetText 获取整个文本内容
func (lm *LineMap) GetText(text string) string {
	return text
}

// SplitLines 将文本分割成行
func SplitLines(text string) []string {
	return strings.Split(text, "\n")
}
