package textbuffer

import (
	"runtime"
	"strings"
	"time"
	"unsafe"
)

const (
	// 初始间隙大小
	initialGapSize = 4096
	// 最小间隙大小
	minGapSize = 1024
	// 最大间隙大小
	maxGapSize = 1024 * 1024 // 增加到1MB
	// 当间隙大小小于此值时进行扩展
	gapSizeThreshold = 1024
	// 行缓存的大小
	lineCacheSize = 10000
	// 大文本阈值，超过此值使用优化方法
	largeTextThreshold = 10 * 1024 * 1024 // 10MB
	// 分块大小，用于分块处理大文本
	chunkSize = 1024 * 1024 // 1MB
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
	// 缓冲区大小（不包括间隙）
	size int
	// 缓存的文本，避免重复构建
	cachedText string
	// 文本缓存是否有效
	textCacheValid bool
	// 行数缓存
	cachedLineCount int
	// 行数缓存是否有效
	lineCountCacheValid bool
	// 行起始位置缓存
	lineStartCache []int
	// 行起始位置缓存是否有效
	lineStartCacheValid bool
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
	const initialGapSize = 128
	buffer := make([]rune, initialGapSize)
	return &GapBuffer{
		buffer:              buffer,
		gapStart:            0,
		gapEnd:              initialGapSize,
		size:                0,
		cachedText:          "",
		textCacheValid:      true, // 空字符串是有效的
		cachedLineCount:     1,    // 空文本有一行
		lineCountCacheValid: true,
		lineStartCache:      []int{0},
		lineStartCacheValid: true,
	}
}

// NewGapBufferWithText 创建一个新的GapBuffer，并初始化文本内容
func NewGapBufferWithText(text string) *GapBuffer {
	gb := NewGapBuffer()
	if text != "" {
		gb.SetText(text)
	}
	return gb
}

// GetText 获取整个文本内容
func (gb *GapBuffer) GetText() string {
	if gb.size == 0 {
		return ""
	}

	var builder strings.Builder
	builder.Grow(gb.size)

	if gb.gapStart > 0 {
		builder.WriteString(string(gb.buffer[:gb.gapStart]))
	}

	if gb.gapEnd < len(gb.buffer) {
		builder.WriteString(string(gb.buffer[gb.gapEnd:]))
	}

	return builder.String()
}

// GetLength 获取文本总长度
func (gb *GapBuffer) GetLength() int {
	return gb.size
}

// GetLineCount 获取行数
func (gb *GapBuffer) GetLineCount() int {
	// 如果缓存有效，直接返回缓存
	if gb.lineCountCacheValid {
		return gb.cachedLineCount
	}

	// 使用lineStartCache获取行数
	gb.updateLineStartCache()
	count := len(gb.lineStartCache)

	gb.cachedLineCount = count
	gb.lineCountCacheValid = true
	return count
}

// GetLineContent 获取指定行的内容
func (gb *GapBuffer) GetLineContent(lineIndex int) string {
	if lineIndex < 0 {
		return ""
	}

	gb.updateLineStartCache()
	if lineIndex >= len(gb.lineStartCache) {
		return ""
	}

	text := gb.GetText()
	start := gb.lineStartCache[lineIndex]

	// 计算行结束位置
	end := len(text)
	if lineIndex < len(gb.lineStartCache)-1 {
		end = gb.lineStartCache[lineIndex+1]
	}

	// 返回行内容
	if start < end {
		return text[start:end]
	}

	return ""
}

// GetLines 获取所有行的内容
func (gb *GapBuffer) GetLines() []string {
	gb.updateLineStartCache()
	text := gb.GetText()

	if text == "" {
		return []string{""}
	}

	lines := make([]string, len(gb.lineStartCache))
	for i := 0; i < len(gb.lineStartCache); i++ {
		lines[i] = gb.GetLineContent(i)
	}

	return lines
}

// GetPositionAt 获取指定偏移量对应的位置
func (gb *GapBuffer) GetPositionAt(offset int) Position {
	if offset < 0 {
		return Position{Line: 0, Column: 0}
	}

	if offset >= gb.size {
		// 如果偏移量超出文本长度，返回最后一个位置
		lineCount := gb.GetLineCount()
		if lineCount == 0 {
			return Position{Line: 0, Column: 0}
		}

		lastLine := lineCount - 1
		lastLineStart := gb.getLineStart(lastLine)
		lastLineLength := gb.size - lastLineStart

		return Position{Line: lastLine, Column: lastLineLength}
	}

	// 二分查找找到行号
	gb.updateLineStartCache()
	line := 0
	left, right := 0, len(gb.lineStartCache)-1

	for left <= right {
		mid := (left + right) / 2
		if gb.lineStartCache[mid] <= offset {
			line = mid
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	// 计算列号
	column := offset - gb.lineStartCache[line]

	return Position{Line: line, Column: column}
}

// getLineStart 获取指定行的起始位置
func (gb *GapBuffer) getLineStart(line int) int {
	gb.updateLineStartCache()
	if line < 0 || line >= len(gb.lineStartCache) {
		return 0
	}
	return gb.lineStartCache[line]
}

// GetOffsetAt 获取指定位置对应的偏移量
func (gb *GapBuffer) GetOffsetAt(position Position) int {
	if position.Line < 0 || position.Column < 0 {
		return 0
	}

	gb.updateLineStartCache()
	if position.Line >= len(gb.lineStartCache) {
		return gb.size
	}

	lineStart := gb.lineStartCache[position.Line]

	// 计算行结束位置
	lineEnd := gb.size
	if position.Line < len(gb.lineStartCache)-1 {
		lineEnd = gb.lineStartCache[position.Line+1] - 1 // -1 to exclude newline
	}

	// 确保列号不超过行长度
	column := position.Column
	if lineStart+column > lineEnd {
		column = lineEnd - lineStart
	}

	return lineStart + column
}

// GetTextInRange 获取指定范围的文本
func (gb *GapBuffer) GetTextInRange(start, end int) string {
	if start < 0 || end > gb.size || start >= end {
		return ""
	}

	var builder strings.Builder
	builder.Grow(end - start)

	if end <= gb.gapStart {
		// 范围在间隙之前
		builder.WriteString(string(gb.buffer[start:end]))
	} else if start >= gb.gapStart {
		// 范围在间隙之后
		gapSize := gb.gapEnd - gb.gapStart
		builder.WriteString(string(gb.buffer[start+gapSize : end+gapSize]))
	} else {
		// 范围跨越间隙
		builder.WriteString(string(gb.buffer[start:gb.gapStart]))
		builder.WriteString(string(gb.buffer[gb.gapEnd : gb.gapEnd+(end-gb.gapStart)]))
	}

	return builder.String()
}

// moveGap 将间隙移动到指定位置
func (gb *GapBuffer) moveGap(pos int) {
	if pos == gb.gapStart {
		return
	}

	// 确保位置在有效范围内
	if pos < 0 {
		pos = 0
	} else if pos > gb.size {
		pos = gb.size
	}

	// 计算间隙大小
	gapSize := gb.gapEnd - gb.gapStart

	if pos < gb.gapStart {
		// 向左移动间隙
		// 将[pos, gapStart)的内容移动到间隙后面
		moveLen := gb.gapStart - pos
		if moveLen > 0 && gb.gapEnd+moveLen <= len(gb.buffer) {
			for i := 0; i < moveLen; i++ {
				gb.buffer[gb.gapEnd+i] = gb.buffer[pos+i]
			}
		}
	} else {
		// 向右移动间隙
		// 将[gapEnd, pos+gapSize)的内容移动到间隙前面
		moveLen := pos - gb.gapStart
		if moveLen > 0 && gb.gapEnd+moveLen <= len(gb.buffer) {
			for i := 0; i < moveLen; i++ {
				gb.buffer[gb.gapStart+i] = gb.buffer[gb.gapEnd+i]
			}
		}
	}

	// 更新间隙位置
	gb.gapStart = pos
	gb.gapEnd = pos + gapSize
}

// ensureGapSize 确保间隙有足够的空间
func (gb *GapBuffer) ensureGapSize(minSize int) {
	gapSize := gb.gapEnd - gb.gapStart
	if gapSize >= minSize {
		return
	}

	// 计算新的间隙大小
	newGapSize := initialGapSize
	for newGapSize < minSize {
		newGapSize *= 2
	}
	if newGapSize > maxGapSize {
		newGapSize = maxGapSize
	}

	// 创建新的缓冲区
	newSize := len(gb.buffer) - gapSize + newGapSize
	newBuffer := make([]rune, newSize)

	// 复制间隙前的内容
	copy(newBuffer, gb.buffer[:gb.gapStart])

	// 复制间隙后的内容
	copy(newBuffer[gb.gapStart+newGapSize:], gb.buffer[gb.gapEnd:])

	// 更新缓冲区
	gb.buffer = newBuffer
	gb.gapEnd = gb.gapStart + newGapSize
}

// Insert 在指定位置插入文本
func (gb *GapBuffer) Insert(position int, text string) {
	if text == "" {
		return
	}

	// 确保位置在有效范围内
	if position < 0 {
		position = 0
	} else if position > gb.size {
		position = gb.size
	}

	// 将间隙移动到插入位置
	gb.moveGap(position)

	// 计算需要插入的文本长度
	insertRunes := []rune(text)
	insertSize := len(insertRunes)

	// 确保间隙有足够的空间
	gb.ensureGapSize(insertSize)

	// 插入文本
	copy(gb.buffer[gb.gapStart:gb.gapStart+insertSize], insertRunes)
	gb.gapStart += insertSize
	gb.size += insertSize

	// 使行缓存失效
	gb.lineStartCacheValid = false
}

// Delete 删除指定范围的文本
func (gb *GapBuffer) Delete(start, end int) {
	if start < 0 || end > gb.size || start >= end {
		return
	}

	// 将间隙移动到删除范围的起始位置
	gb.moveGap(start)

	// 更新间隙大小
	deleteSize := end - start
	gb.gapEnd += deleteSize
	gb.size -= deleteSize

	// 如果间隙太大，缩小它
	gapSize := gb.gapEnd - gb.gapStart
	if gapSize > maxGapSize {
		// 创建新的缓冲区
		newGapSize := initialGapSize
		newSize := len(gb.buffer) - gapSize + newGapSize
		newBuffer := make([]rune, newSize)

		// 复制间隙前的内容
		copy(newBuffer, gb.buffer[:gb.gapStart])

		// 复制间隙后的内容
		copy(newBuffer[gb.gapStart+newGapSize:], gb.buffer[gb.gapEnd:])

		// 更新缓冲区
		gb.buffer = newBuffer
		gb.gapEnd = gb.gapStart + newGapSize
	}

	// 使行缓存失效
	gb.lineStartCacheValid = false
}

// Clear 清空文本缓冲区
func (gb *GapBuffer) Clear() {
	// 创建新的缓冲区
	gb.buffer = make([]rune, initialGapSize)
	gb.gapStart = 0
	gb.gapEnd = initialGapSize
	gb.size = 0
	gb.lineStartCacheValid = false
}

// SetText 设置整个文本内容
func (gb *GapBuffer) SetText(text string) {
	runes := []rune(text)
	size := len(runes)

	// 如果新文本大小超过当前缓冲区大小，创建新的缓冲区
	if size > len(gb.buffer)-(gb.gapEnd-gb.gapStart) {
		bufferSize := size + initialGapSize
		gb.buffer = make([]rune, bufferSize)
		if size > 0 {
			copy(gb.buffer, runes)
		}
		gb.gapStart = size
		gb.gapEnd = bufferSize
	} else {
		// 将间隙移动到开始位置
		gb.moveGap(0)
		// 复制文本到缓冲区
		if size > 0 {
			copy(gb.buffer[0:size], runes)
		}
		// 更新间隙位置
		gb.gapStart = size
		gb.gapEnd = len(gb.buffer)
	}

	gb.size = size
	gb.lineStartCacheValid = false
}

// updateLineStartCache 更新行起始位置缓存
func (gb *GapBuffer) updateLineStartCache() {
	if gb.lineStartCacheValid {
		return
	}

	gb.lineStartCache = gb.lineStartCache[:0]

	// 第一行总是从0开始
	gb.lineStartCache = append(gb.lineStartCache, 0)

	// 如果缓冲区为空，只有一行，起始位置为0
	if gb.size == 0 {
		gb.lineStartCacheValid = true
		return
	}

	// 处理间隙前的内容
	for i := 0; i < gb.gapStart; i++ {
		if gb.buffer[i] == '\n' {
			// 找到一个换行符，下一行的起始位置是当前位置+1
			gb.lineStartCache = append(gb.lineStartCache, i+1)
		}
	}

	// 处理间隙后的内容
	for i := gb.gapEnd; i < len(gb.buffer); i++ {
		if gb.buffer[i] == '\n' {
			// 找到一个换行符，下一行的起始位置是当前位置+1（需要调整间隙）
			realPos := i - (gb.gapEnd - gb.gapStart)
			gb.lineStartCache = append(gb.lineStartCache, realPos+1)
		}
	}

	gb.lineStartCacheValid = true
}

// GetLineStart 获取指定行的起始位置
func (gb *GapBuffer) GetLineStart(line int) int {
	return gb.getLineStart(line)
}

// GetLineLength 获取指定行的长度
func (gb *GapBuffer) GetLineLength(line int) int {
	gb.updateLineStartCache()
	if line < 0 || line >= len(gb.lineStartCache) {
		return 0
	}

	lineStart := gb.lineStartCache[line]
	lineEnd := gb.size

	if line < len(gb.lineStartCache)-1 {
		lineEnd = gb.lineStartCache[line+1] - 1 // -1 to exclude newline
	}

	return lineEnd - lineStart
}

// Close 关闭GapBuffer，释放资源
func (gb *GapBuffer) Close() {
	// 不需要释放资源，因为不使用MemoryMonitor和MemoryPool
}

// GetMemoryStats 获取内存使用统计信息
func (gb *GapBuffer) GetMemoryStats() MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 计算当前GapBuffer使用的内存
	bufferSize := len(gb.buffer) * int(unsafe.Sizeof(rune(0)))
	lineStartCacheSize := len(gb.lineStartCache) * int(unsafe.Sizeof(int(0)))

	// 估算总内存使用
	currentUsage := bufferSize + lineStartCacheSize

	return MemoryStats{
		CurrentUsage:   uint64(currentUsage),
		PeakUsage:      uint64(memStats.TotalAlloc),
		TotalAllocated: uint64(memStats.TotalAlloc),
		Allocations:    uint64(memStats.Mallocs),
		Deallocations:  uint64(memStats.Frees),
		UptimeSeconds:  uint64(time.Since(startTime).Seconds()),
	}
}

// GetTextChunk 获取指定范围的文本块，适用于大文本处理
func (gb *GapBuffer) GetTextChunk(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > gb.size {
		end = gb.size
	}
	if start >= end {
		return ""
	}

	// 计算实际缓冲区索引
	realStart := start
	if start > gb.gapStart {
		realStart += (gb.gapEnd - gb.gapStart)
	}

	realEnd := end
	if end > gb.gapStart {
		realEnd += (gb.gapEnd - gb.gapStart)
	}

	// 如果范围不跨越间隙，直接返回
	if start <= gb.gapStart && end <= gb.gapStart || start >= gb.gapStart && end >= gb.gapStart {
		return string(gb.buffer[realStart:realEnd])
	}

	// 范围跨越间隙，需要拼接
	var builder strings.Builder
	builder.Grow(end - start)

	// 添加间隙前的部分
	if start < gb.gapStart {
		builder.WriteString(string(gb.buffer[start:gb.gapStart]))
	}

	// 添加间隙后的部分
	if end > gb.gapStart {
		builder.WriteString(string(gb.buffer[gb.gapEnd:realEnd]))
	}

	return builder.String()
}

// InsertChunk 在指定位置插入大块文本，针对大文本优化
func (gb *GapBuffer) InsertChunk(pos int, text string) {
	if text == "" {
		return
	}

	// 确保位置在有效范围内
	if pos < 0 {
		pos = 0
	} else if pos > gb.size {
		pos = gb.size
	}

	// 计算需要插入的文本长度
	insertRunes := []rune(text)
	insertSize := len(insertRunes)

	// 对于超大文本，直接创建新的缓冲区
	if insertSize > largeTextThreshold {
		// 获取当前文本的前半部分和后半部分
		beforeText := gb.GetTextChunk(0, pos)
		afterText := gb.GetTextChunk(pos, gb.size)

		// 创建新文本
		newText := beforeText + text + afterText

		// 设置新文本
		gb.SetText(newText)
		return
	}

	// 对于中等大小的文本，使用分块插入
	if insertSize > maxGapSize {
		// 将间隙移动到插入位置
		gb.moveGap(pos)

		// 如果间隙不够大，扩展它
		if gb.gapEnd-gb.gapStart < insertSize {
			// 计算新的缓冲区大小
			newSize := len(gb.buffer) + insertSize - (gb.gapEnd - gb.gapStart) + initialGapSize
			newBuffer := make([]rune, newSize)

			// 复制间隙前的内容
			copy(newBuffer, gb.buffer[:gb.gapStart])

			// 复制间隙后的内容
			gapSize := initialGapSize
			copy(newBuffer[gb.gapStart+insertSize+gapSize:], gb.buffer[gb.gapEnd:])

			// 更新缓冲区
			gb.buffer = newBuffer
			gb.gapEnd = gb.gapStart + insertSize + gapSize
		}

		// 分块插入文本
		chunkSize := maxGapSize / 2
		for i := 0; i < insertSize; i += chunkSize {
			end := i + chunkSize
			if end > insertSize {
				end = insertSize
			}

			// 计算当前块的大小
			currentChunkSize := end - i

			// 在间隙起始位置插入文本块
			copy(gb.buffer[gb.gapStart:gb.gapStart+currentChunkSize], insertRunes[i:end])

			// 更新间隙起始位置
			gb.gapStart += currentChunkSize
		}

		// 更新文本大小
		gb.size += insertSize

		// 使缓存失效
		gb.textCacheValid = false
		gb.lineCountCacheValid = false
		gb.lineStartCacheValid = false
		return
	}

	// 对于小文本，使用标准的间隙缓冲区方法
	gb.Insert(pos, text)
}

// DeleteChunk 删除指定范围的大块文本，针对大文本优化
func (gb *GapBuffer) DeleteChunk(start, end int) {
	// 确保范围在有效范围内
	if start < 0 {
		start = 0
	}
	if end > gb.size {
		end = gb.size
	}
	if start >= end {
		return
	}

	deleteSize := end - start

	// 对于超大范围删除，直接创建新缓冲区
	if deleteSize > largeTextThreshold {
		// 获取当前文本的前半部分和后半部分
		beforeText := gb.GetTextChunk(0, start)
		afterText := gb.GetTextChunk(end, gb.size)

		// 创建新文本
		newText := beforeText + afterText

		// 设置新文本
		gb.SetText(newText)
		return
	}

	// 对于中等大小的删除，使用优化的方法
	if deleteSize > maxGapSize {
		// 将间隙移动到删除范围的起始位置
		gb.moveGap(start)

		// 更新间隙大小
		gb.gapEnd += deleteSize
		gb.size -= deleteSize

		// 如果间隙太大，缩小它
		gapSize := gb.gapEnd - gb.gapStart
		if gapSize > maxGapSize*2 {
			// 创建新的缓冲区
			newGapSize := initialGapSize
			newSize := len(gb.buffer) - gapSize + newGapSize
			newBuffer := make([]rune, newSize)

			// 复制间隙前的内容
			copy(newBuffer, gb.buffer[:gb.gapStart])

			// 复制间隙后的内容
			copy(newBuffer[gb.gapStart+newGapSize:], gb.buffer[gb.gapEnd:])

			// 更新缓冲区
			gb.buffer = newBuffer
			gb.gapEnd = gb.gapStart + newGapSize
		}

		// 使缓存失效
		gb.textCacheValid = false
		gb.lineCountCacheValid = false
		gb.lineStartCacheValid = false
		return
	}

	// 对于小范围删除，使用标准的间隙缓冲区方法
	gb.Delete(start, end)
}

// ReplaceChunk 替换指定范围的大块文本，针对大文本优化
func (gb *GapBuffer) ReplaceChunk(start, end int, text string) {
	// 确保范围在有效范围内
	if start < 0 {
		start = 0
	}
	if end > gb.size {
		end = gb.size
	}
	if start > end {
		start = end
	}

	insertRunes := []rune(text)
	insertSize := len(insertRunes)
	deleteSize := end - start

	// 对于超大文本操作，直接创建新缓冲区
	if insertSize > largeTextThreshold || deleteSize > largeTextThreshold {
		// 获取当前文本的前半部分和后半部分
		beforeText := gb.GetTextChunk(0, start)
		afterText := gb.GetTextChunk(end, gb.size)

		// 创建新文本
		newText := beforeText + text + afterText

		// 设置新文本
		gb.SetText(newText)
		return
	}

	// 对于中等大小的操作，使用优化的方法
	if insertSize > maxGapSize || deleteSize > maxGapSize {
		// 删除旧文本
		gb.DeleteChunk(start, end)

		// 插入新文本
		gb.InsertChunk(start, text)
		return
	}

	// 对于小范围替换，使用标准的间隙缓冲区方法
	gb.Delete(start, end)
	gb.Insert(start, text)
}

// FindTextForward 向前搜索文本，针对大文本优化
func (gb *GapBuffer) FindTextForward(searchText string, startPos int, caseSensitive bool) (int, int) {
	if searchText == "" || startPos >= gb.size {
		return -1, -1
	}

	if startPos < 0 {
		startPos = 0
	}

	// 获取文本内容
	text := gb.GetText()

	// 如果不区分大小写，转换为小写
	if !caseSensitive {
		text = strings.ToLower(text)
		searchText = strings.ToLower(searchText)
	}

	// 搜索文本
	pos := strings.Index(text[startPos:], searchText)
	if pos == -1 {
		return -1, -1
	}

	// 计算实际位置
	start := startPos + pos
	end := start + len([]rune(searchText))

	return start, end
}

// FindTextBackward 向后搜索文本，针对大文本优化
func (gb *GapBuffer) FindTextBackward(searchText string, startPos int, caseSensitive bool) (int, int) {
	if searchText == "" || startPos <= 0 {
		return -1, -1
	}

	if startPos > gb.size {
		startPos = gb.size
	}

	// 获取文本内容
	text := gb.GetText()

	// 如果不区分大小写，转换为小写
	if !caseSensitive {
		text = strings.ToLower(text)
		searchText = strings.ToLower(searchText)
	}

	// 搜索文本
	pos := strings.LastIndex(text[:startPos], searchText)
	if pos == -1 {
		return -1, -1
	}

	// 计算实际位置
	start := pos
	end := start + len([]rune(searchText))

	return start, end
}
