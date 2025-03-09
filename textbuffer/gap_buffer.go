package textbuffer

import (
	"strings"
)

// GapBuffer 是一个基于Gap Buffer的文本缓冲区
// Gap Buffer是一种高效的文本编辑数据结构，它在文本中维护一个"间隙"，
// 使得在当前编辑位置附近的插入和删除操作可以在常数时间内完成
type GapBuffer struct {
	// 缓冲区内容
	buffer []rune
	// 间隙起始位置
	gapStart int
	// 间隙结束位置
	gapEnd int
	// 缓冲区大小
	size int
	// 行信息缓存
	lineCache []lineInfo
	// 行信息是否有效
	lineCacheValid bool
}

// lineInfo 存储行的信息
type lineInfo struct {
	// 行的起始位置（不包括间隙）
	start int
	// 行的长度（不包括间隙和换行符）
	length int
	// 是否包含换行符
	hasNewline bool
}

// NewGapBuffer 创建一个新的GapBuffer
func NewGapBuffer() *GapBuffer {
	initialGapSize := 128
	buffer := make([]rune, initialGapSize)
	return &GapBuffer{
		buffer:         buffer,
		gapStart:       0,
		gapEnd:         initialGapSize,
		size:           0,
		lineCache:      nil,
		lineCacheValid: false,
	}
}

// NewGapBufferWithText 创建一个新的GapBuffer，并初始化文本内容
func NewGapBufferWithText(text string) *GapBuffer {
	if text == "" {
		return NewGapBuffer()
	}

	runes := []rune(text)
	textLength := len(runes)
	initialGapSize := 128
	buffer := make([]rune, textLength+initialGapSize)

	// 复制文本到缓冲区
	copy(buffer, runes)

	gb := &GapBuffer{
		buffer:         buffer,
		gapStart:       textLength,
		gapEnd:         textLength + initialGapSize,
		size:           textLength,
		lineCache:      nil,
		lineCacheValid: false,
	}

	// 更新行信息
	gb.updateLineCache()

	return gb
}

// GetText 获取整个文本内容
func (gb *GapBuffer) GetText() string {
	if gb.size == 0 {
		return ""
	}

	var builder strings.Builder
	builder.Grow(gb.size)

	// 添加间隙前的文本
	for i := 0; i < gb.gapStart; i++ {
		builder.WriteRune(gb.buffer[i])
	}

	// 添加间隙后的文本
	for i := gb.gapEnd; i < len(gb.buffer); i++ {
		builder.WriteRune(gb.buffer[i])
	}

	return builder.String()
}

// GetLength 获取文本总长度
func (gb *GapBuffer) GetLength() int {
	return gb.size
}

// GetLineCount 获取行数
func (gb *GapBuffer) GetLineCount() int {
	if !gb.lineCacheValid {
		gb.updateLineCache()
	}
	return len(gb.lineCache)
}

// GetLineContent 获取指定行的内容
func (gb *GapBuffer) GetLineContent(lineIndex int) string {
	if !gb.lineCacheValid {
		gb.updateLineCache()
	}

	if lineIndex < 0 || lineIndex >= len(gb.lineCache) {
		return ""
	}

	lineInfo := gb.lineCache[lineIndex]
	var builder strings.Builder
	builder.Grow(lineInfo.length + 1) // +1 for potential newline

	// 获取行内容（考虑间隙）
	start := lineInfo.start
	length := lineInfo.length

	// 如果行起始位置在间隙之前
	if start < gb.gapStart {
		// 如果整行都在间隙之前
		if start+length <= gb.gapStart {
			for i := start; i < start+length; i++ {
				builder.WriteRune(gb.buffer[i])
			}
		} else {
			// 行跨越了间隙
			for i := start; i < gb.gapStart; i++ {
				builder.WriteRune(gb.buffer[i])
			}
			remaining := length - (gb.gapStart - start)
			for i := gb.gapEnd; i < gb.gapEnd+remaining; i++ {
				builder.WriteRune(gb.buffer[i])
			}
		}
	} else {
		// 行起始位置在间隙之后
		adjustedStart := start + (gb.gapEnd - gb.gapStart)
		for i := adjustedStart; i < adjustedStart+length; i++ {
			builder.WriteRune(gb.buffer[i])
		}
	}

	// 添加换行符（如果有）
	if lineInfo.hasNewline {
		builder.WriteRune('\n')
	}

	return builder.String()
}

// GetLines 获取所有行的内容
func (gb *GapBuffer) GetLines() []string {
	lineCount := gb.GetLineCount()
	lines := make([]string, lineCount)

	for i := 0; i < lineCount; i++ {
		lines[i] = gb.GetLineContent(i)
	}

	return lines
}

// GetPositionAt 获取指定偏移量对应的位置
func (gb *GapBuffer) GetPositionAt(offset int) Position {
	if !gb.lineCacheValid {
		gb.updateLineCache()
	}

	if offset <= 0 {
		return Position{Line: 0, Column: 0}
	}

	if offset >= gb.size {
		// 如果偏移量超出文本长度，返回最后一个位置
		lastLine := len(gb.lineCache) - 1
		lastColumn := gb.lineCache[lastLine].length
		return Position{Line: lastLine, Column: lastColumn}
	}

	// 将偏移量转换为实际缓冲区索引
	realOffset := offset
	if offset > gb.gapStart {
		realOffset += (gb.gapEnd - gb.gapStart)
	}

	// 查找包含偏移量的行
	currentOffset := 0
	for i, line := range gb.lineCache {
		lineLength := line.length
		if line.hasNewline {
			lineLength++
		}

		if currentOffset+lineLength > offset {
			// 找到了包含偏移量的行
			column := offset - currentOffset
			return Position{Line: i, Column: column}
		}

		currentOffset += lineLength
	}

	// 如果没有找到，返回最后一个位置
	lastLine := len(gb.lineCache) - 1
	lastColumn := gb.lineCache[lastLine].length
	return Position{Line: lastLine, Column: lastColumn}
}

// GetOffsetAt 获取指定位置对应的偏移量
func (gb *GapBuffer) GetOffsetAt(position Position) int {
	if !gb.lineCacheValid {
		gb.updateLineCache()
	}

	if position.Line < 0 {
		return 0
	}

	if position.Line >= len(gb.lineCache) {
		return gb.size
	}

	// 计算偏移量
	offset := 0
	for i := 0; i < position.Line; i++ {
		offset += gb.lineCache[i].length
		if gb.lineCache[i].hasNewline {
			offset++
		}
	}

	// 添加列偏移
	if position.Column > gb.lineCache[position.Line].length {
		offset += gb.lineCache[position.Line].length
	} else {
		offset += position.Column
	}

	return offset
}

// GetTextInRange 获取指定范围内的文本
func (gb *GapBuffer) GetTextInRange(r Range) string {
	startOffset := gb.GetOffsetAt(r.Start)
	endOffset := gb.GetOffsetAt(r.End)

	if startOffset >= endOffset {
		return ""
	}

	// 将偏移量转换为实际缓冲区索引
	realStartOffset := startOffset
	if startOffset > gb.gapStart {
		realStartOffset += (gb.gapEnd - gb.gapStart)
	}

	realEndOffset := endOffset
	if endOffset > gb.gapStart {
		realEndOffset += (gb.gapEnd - gb.gapStart)
	}

	var builder strings.Builder
	builder.Grow(endOffset - startOffset)

	// 如果范围不跨越间隙
	if (startOffset < gb.gapStart && endOffset <= gb.gapStart) ||
		(startOffset >= gb.gapStart && endOffset > gb.gapStart) {
		for i := realStartOffset; i < realEndOffset; i++ {
			builder.WriteRune(gb.buffer[i])
		}
	} else {
		// 范围跨越了间隙
		for i := realStartOffset; i < gb.gapStart; i++ {
			builder.WriteRune(gb.buffer[i])
		}
		for i := gb.gapEnd; i < realEndOffset; i++ {
			builder.WriteRune(gb.buffer[i])
		}
	}

	return builder.String()
}

// moveGap 将间隙移动到指定位置
func (gb *GapBuffer) moveGap(offset int) {
	if offset == gb.gapStart {
		return
	}

	// 确保偏移量在有效范围内
	if offset < 0 {
		offset = 0
	} else if offset > gb.size {
		offset = gb.size
	}

	// 计算实际偏移量（考虑间隙）
	realOffset := offset
	if offset > gb.gapStart {
		realOffset += (gb.gapEnd - gb.gapStart)
	}

	// 移动间隙
	if offset < gb.gapStart {
		// 向左移动间隙
		// 将间隙左侧的数据移动到间隙右侧
		gapSize := gb.gapEnd - gb.gapStart
		moveLength := gb.gapStart - offset

		// 将数据从间隙左侧移动到间隙右侧
		copy(gb.buffer[gb.gapEnd-moveLength:], gb.buffer[offset:gb.gapStart])

		// 更新间隙位置
		gb.gapStart = offset
		gb.gapEnd = offset + gapSize
	} else {
		// 向右移动间隙
		// 将间隙右侧的数据移动到间隙左侧
		gapSize := gb.gapEnd - gb.gapStart
		moveLength := realOffset - gb.gapEnd

		// 将数据从间隙右侧移动到间隙左侧
		copy(gb.buffer[gb.gapStart:], gb.buffer[gb.gapEnd:gb.gapEnd+moveLength])

		// 更新间隙位置
		gb.gapStart = gb.gapStart + moveLength
		gb.gapEnd = gb.gapStart + gapSize
	}
}

// ensureGapCapacity 确保间隙有足够的容量
func (gb *GapBuffer) ensureGapCapacity(needed int) {
	gapSize := gb.gapEnd - gb.gapStart
	if gapSize >= needed {
		return
	}

	// 计算新的缓冲区大小
	newCapacity := len(gb.buffer) * 2
	for newCapacity-len(gb.buffer) < needed-gapSize {
		newCapacity *= 2
	}

	// 创建新的缓冲区
	newBuffer := make([]rune, newCapacity)

	// 复制间隙前的数据
	copy(newBuffer, gb.buffer[:gb.gapStart])

	// 计算新的间隙结束位置
	newGapEnd := gb.gapStart + (newCapacity - len(gb.buffer) + gapSize)

	// 复制间隙后的数据
	copy(newBuffer[newGapEnd:], gb.buffer[gb.gapEnd:])

	// 更新缓冲区和间隙结束位置
	gb.buffer = newBuffer
	gb.gapEnd = newGapEnd
}

// Insert 在指定位置插入文本
func (gb *GapBuffer) Insert(offset int, text string) {
	if text == "" {
		return
	}

	// 确保偏移量在有效范围内
	if offset < 0 {
		offset = 0
	} else if offset > gb.size {
		offset = gb.size
	}

	// 将间隙移动到插入位置
	gb.moveGap(offset)

	// 获取要插入的文本
	runes := []rune(text)
	insertLength := len(runes)

	// 确保间隙有足够的容量
	gb.ensureGapCapacity(insertLength)

	// 将文本插入到间隙中
	copy(gb.buffer[gb.gapStart:], runes)

	// 更新间隙位置和文本大小
	gb.gapStart += insertLength
	gb.size += insertLength

	// 标记行信息缓存为无效
	gb.lineCacheValid = false
}

// Delete 删除指定范围的文本
func (gb *GapBuffer) Delete(startOffset, endOffset int) {
	if startOffset >= endOffset {
		return
	}

	// 确保偏移量在有效范围内
	if startOffset < 0 {
		startOffset = 0
	}
	if endOffset > gb.size {
		endOffset = gb.size
	}

	// 将间隙移动到删除范围的起始位置
	gb.moveGap(startOffset)

	// 计算删除的长度
	deleteLength := endOffset - startOffset

	// 更新间隙位置和文本大小
	gb.gapEnd += deleteLength
	gb.size -= deleteLength

	// 标记行信息缓存为无效
	gb.lineCacheValid = false
}

// Clear 清空文本缓冲区
func (gb *GapBuffer) Clear() {
	initialGapSize := 128
	gb.buffer = make([]rune, initialGapSize)
	gb.gapStart = 0
	gb.gapEnd = initialGapSize
	gb.size = 0
	gb.lineCache = nil
	gb.lineCacheValid = false
}

// SetText 设置整个文本内容
func (gb *GapBuffer) SetText(text string) {
	gb.Clear()
	gb.Insert(0, text)
}

// updateLineCache 更新行信息缓存
func (gb *GapBuffer) updateLineCache() {
	gb.lineCache = nil

	if gb.size == 0 {
		// 空文本至少有一行
		gb.lineCache = []lineInfo{{start: 0, length: 0, hasNewline: false}}
		gb.lineCacheValid = true
		return
	}

	// 计算行信息
	lineStart := 0
	lineLength := 0
	inGap := false

	for i := 0; i < len(gb.buffer); i++ {
		// 跳过间隙
		if i == gb.gapStart {
			inGap = true
			i = gb.gapEnd - 1 // -1 因为循环会 i++
			continue
		}

		if inGap && i < gb.gapEnd {
			continue
		}

		// 计算实际偏移量（考虑间隙）
		realOffset := i
		if inGap {
			realOffset -= (gb.gapEnd - gb.gapStart)
		}

		// 处理换行符
		if gb.buffer[i] == '\n' {
			// 添加当前行信息
			adjustedStart := lineStart
			if lineStart > gb.gapStart {
				adjustedStart -= (gb.gapEnd - gb.gapStart)
			}
			gb.lineCache = append(gb.lineCache, lineInfo{
				start:      adjustedStart,
				length:     lineLength,
				hasNewline: true,
			})

			// 开始新行
			lineStart = realOffset + 1
			lineLength = 0
		} else {
			lineLength++
		}
	}

	// 添加最后一行（如果没有以换行符结束）
	if lineLength > 0 || len(gb.lineCache) == 0 {
		adjustedStart := lineStart
		if lineStart > gb.gapStart {
			adjustedStart -= (gb.gapEnd - gb.gapStart)
		}
		gb.lineCache = append(gb.lineCache, lineInfo{
			start:      adjustedStart,
			length:     lineLength,
			hasNewline: false,
		})
	}

	gb.lineCacheValid = true
}
