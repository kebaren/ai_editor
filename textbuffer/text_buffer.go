package textbuffer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"
)

// EOLType represents the type of end-of-line character(s)
type EOLType int

const (
	// EOLUnix represents Unix-style line endings (\n)
	EOLUnix EOLType = iota
	// EOLWindows represents Windows-style line endings (\r\n)
	EOLWindows
	// EOLMac represents old Mac-style line endings (\r)
	EOLMac
)

// EventType represents the type of event
type EventType int

const (
	// EventTextChanged is triggered when the text is changed
	EventTextChanged EventType = iota
	// EventCursorMoved is triggered when the cursor is moved
	EventCursorMoved
	// EventSelectionChanged is triggered when the selection is changed
	EventSelectionChanged
	// EventLanguageChanged is triggered when the language is changed
	EventLanguageChanged
	// EventModifiedChanged is triggered when the modified flag is changed
	EventModifiedChanged
	// EventFilePathChanged is triggered when the file path is changed
	EventFilePathChanged
)

// TextBuffer 是一个文本缓冲区，用于存储和操作文本
type TextBuffer struct {
	// 文本内容的数据结构（使用GapBuffer替代LineBuffer）
	gapBuffer *GapBuffer
	// 互斥锁，用于并发访问
	mutex sync.RWMutex
	// 撤销/重做栈
	undoStack *UndoStack
	// EOL type for this buffer
	eolType EOLType
	// Language ID for syntax highlighting and LSP
	languageID string
	// File path associated with this buffer
	filePath string
	// Modified flag
	modified bool
	// Event listeners
	eventListeners map[EventType][]EventListener
	// Lua plugin manager
	luaPluginManager *LuaPluginManager
	// LSP manager
	lspManager *LSPManager
}

// EventListener is a function that handles events
type EventListener func(event Event)

// Event represents an event
type Event struct {
	// Type of event
	Type EventType
	// TextBuffer that triggered the event
	TextBuffer *TextBuffer
	// Additional data
	Data interface{}
}

// TextChangedEventData contains data for text changed events
type TextChangedEventData struct {
	// Position where the change occurred
	Position Position
	// Text that was inserted or deleted
	Text string
	// Old text before the change
	OldText string
	// Operation type (insert, delete, replace, clear, setText)
	OperationType OperationType
}

// CursorMovedEventData contains data for cursor moved events
type CursorMovedEventData struct {
	// Old position
	OldPosition Position
	// New position
	NewPosition Position
}

// SelectionChangedEventData contains data for selection changed events
type SelectionChangedEventData struct {
	// Old selection
	OldSelection Range
	// New selection
	NewSelection Range
}

// NewTextBuffer 创建一个新的TextBuffer
func NewTextBuffer() *TextBuffer {
	return NewTextBufferWithText("")
}

// NewTextBufferWithText 创建一个新的TextBuffer，并初始化文本内容
func NewTextBufferWithText(text string) *TextBuffer {
	gapBuffer := NewGapBufferWithText(text)

	// Detect EOL type from text
	eolType := detectEOLType(text)

	tb := &TextBuffer{
		gapBuffer:      gapBuffer,
		mutex:          sync.RWMutex{},
		undoStack:      NewUndoStack(),
		eolType:        eolType,
		languageID:     "plaintext",
		filePath:       "",
		modified:       false,
		eventListeners: make(map[EventType][]EventListener),
	}

	return tb
}

// detectEOLType detects the EOL type from the given text
func detectEOLType(text string) EOLType {
	// Default to Unix-style
	eolType := EOLUnix

	// Check for Windows-style line endings
	if strings.Contains(text, "\r\n") {
		eolType = EOLWindows
	} else if strings.Contains(text, "\r") && !strings.Contains(text, "\n") {
		// Check for old Mac-style line endings
		eolType = EOLMac
	}

	return eolType
}

// GetEOLType returns the current EOL type
func (tb *TextBuffer) GetEOLType() EOLType {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.eolType
}

// SetEOLType sets the EOL type and converts all line endings in the buffer
func (tb *TextBuffer) SetEOLType(eolType EOLType) error {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.eolType == eolType {
		return nil // No change needed
	}

	// Get current text
	text := tb.gapBuffer.GetText()

	// Convert line endings
	var newText string
	switch eolType {
	case EOLUnix:
		// Convert to \n
		newText = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	case EOLWindows:
		// Convert to \r\n
		// First convert all to \n, then to \r\n
		newText = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
		newText = strings.ReplaceAll(newText, "\n", "\r\n")
	case EOLMac:
		// Convert to \r
		newText = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\r"), "\n", "\r")
	default:
		return errors.New("invalid EOL type")
	}

	// Record operation for undo
	tb.undoStack.Push(&TextOperation{
		Type:     OperationReplace,
		Position: Position{Line: 0, Column: 0},
		Text:     newText,
		OldText:  text,
	})

	// Update text and EOL type
	tb.gapBuffer.SetText(newText)
	tb.eolType = eolType

	return nil
}

// GetEOLString returns the string representation of the current EOL type
func (tb *TextBuffer) GetEOLString() string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	switch tb.eolType {
	case EOLWindows:
		return "\r\n"
	case EOLMac:
		return "\r"
	default:
		return "\n"
	}
}

// GetText 获取整个文本内容
func (tb *TextBuffer) GetText() string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.gapBuffer.GetText()
}

// GetLength 获取文本总长度
func (tb *TextBuffer) GetLength() int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.gapBuffer.GetLength()
}

// GetLineCount 获取行数
func (tb *TextBuffer) GetLineCount() int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.gapBuffer.GetLineCount()
}

// GetLineContent 获取指定行的内容
func (tb *TextBuffer) GetLineContent(lineIndex int) string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.gapBuffer.GetLineContent(lineIndex)
}

// GetLines 获取所有行的内容
func (tb *TextBuffer) GetLines() []string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.gapBuffer.GetLines()
}

// GetPositionAt 获取指定偏移量对应的位置
func (tb *TextBuffer) GetPositionAt(offset int) Position {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.gapBuffer.GetPositionAt(offset)
}

// GetOffsetAt 获取指定位置对应的偏移量
func (tb *TextBuffer) GetOffsetAt(position Position) int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.gapBuffer.GetOffsetAt(position)
}

// GetTextInRange 获取指定范围内的文本
func (tb *TextBuffer) GetTextInRange(r Range) string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	startOffset := tb.gapBuffer.GetOffsetAt(r.Start)
	endOffset := tb.gapBuffer.GetOffsetAt(r.End)

	return tb.gapBuffer.GetTextInRange(startOffset, endOffset)
}

// GetLanguageID returns the language ID
func (tb *TextBuffer) GetLanguageID() string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.languageID
}

// SetLanguageID sets the language ID
func (tb *TextBuffer) SetLanguageID(languageID string) {
	tb.mutex.Lock()
	oldLanguageID := tb.languageID
	tb.languageID = languageID
	tb.mutex.Unlock()

	if oldLanguageID != languageID {
		tb.triggerEvent(EventLanguageChanged, languageID)
	}
}

// GetFilePath returns the file path
func (tb *TextBuffer) GetFilePath() string {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.filePath
}

// SetFilePath sets the file path
func (tb *TextBuffer) SetFilePath(filePath string) {
	tb.mutex.Lock()
	oldFilePath := tb.filePath
	tb.filePath = filePath
	tb.mutex.Unlock()

	if oldFilePath != filePath {
		tb.triggerEvent(EventFilePathChanged, filePath)
	}
}

// IsModified returns whether the buffer has been modified
func (tb *TextBuffer) IsModified() bool {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	return tb.modified
}

// SetModified sets the modified flag
func (tb *TextBuffer) SetModified(modified bool) {
	tb.mutex.Lock()
	oldModified := tb.modified
	tb.modified = modified
	tb.mutex.Unlock()

	if oldModified != modified {
		tb.triggerEvent(EventModifiedChanged, modified)
	}
}

// GetLuaPluginManager returns the Lua plugin manager
func (tb *TextBuffer) GetLuaPluginManager() *LuaPluginManager {
	return tb.luaPluginManager
}

// GetLSPManager returns the LSP manager
func (tb *TextBuffer) GetLSPManager() *LSPManager {
	return tb.lspManager
}

// AddEventListener adds an event listener
func (tb *TextBuffer) AddEventListener(eventType EventType, listener EventListener) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.eventListeners[eventType] == nil {
		tb.eventListeners[eventType] = make([]EventListener, 0)
	}

	tb.eventListeners[eventType] = append(tb.eventListeners[eventType], listener)
}

// RemoveEventListener removes an event listener
func (tb *TextBuffer) RemoveEventListener(eventType EventType, listener EventListener) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	if tb.eventListeners[eventType] == nil {
		return
	}

	for i, l := range tb.eventListeners[eventType] {
		if &l == &listener {
			tb.eventListeners[eventType] = append(tb.eventListeners[eventType][:i], tb.eventListeners[eventType][i+1:]...)
			break
		}
	}
}

// triggerEvent 触发事件
func (tb *TextBuffer) triggerEvent(eventType EventType, data interface{}) {
	// 创建事件对象
	event := Event{
		Type:       eventType,
		TextBuffer: tb,
		Data:       data,
	}

	// 获取事件监听器列表的副本，避免在回调中修改监听器列表
	var listeners []EventListener
	tb.mutex.RLock()
	if tb.eventListeners != nil {
		if eventListeners, ok := tb.eventListeners[eventType]; ok {
			listeners = make([]EventListener, len(eventListeners))
			copy(listeners, eventListeners)
		}
	}
	tb.mutex.RUnlock()

	// 异步触发事件，避免阻塞
	go func() {
		// 调用事件监听器
		for _, listener := range listeners {
			if listener != nil {
				listener(event)
			}
		}
	}()
}

// Insert 在指定位置插入文本
func (tb *TextBuffer) Insert(position Position, text string) error {
	if text == "" {
		return nil
	}

	tb.mutex.Lock()

	// 获取偏移量
	offset := tb.gapBuffer.GetOffsetAt(position)

	// 记录操作用于撤销
	tb.undoStack.Push(&TextOperation{
		Type:     OperationInsert,
		Position: position,
		Text:     text,
		OldText:  "",
	})

	// 执行插入操作
	tb.gapBuffer.Insert(offset, text)

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      position,
		Text:          text,
		OldText:       "",
		OperationType: OperationInsert,
	})

	return nil
}

// Delete 删除指定范围的文本
func (tb *TextBuffer) Delete(r Range) error {
	tb.mutex.Lock()

	// 获取偏移量
	startOffset := tb.gapBuffer.GetOffsetAt(r.Start)
	endOffset := tb.gapBuffer.GetOffsetAt(r.End)

	if startOffset >= endOffset {
		tb.mutex.Unlock()
		return errors.New("invalid range")
	}

	// 获取要删除的文本
	oldText := tb.gapBuffer.GetTextInRange(startOffset, endOffset)

	// 记录操作用于撤销
	tb.undoStack.Push(&TextOperation{
		Type:     OperationDelete,
		Position: r.Start,
		Text:     oldText,
		OldText:  oldText,
	})

	// 执行删除操作
	tb.gapBuffer.Delete(startOffset, endOffset)

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      r.Start,
		Text:          oldText,
		OldText:       oldText,
		OperationType: OperationDelete,
	})

	return nil
}

// Replace 替换指定范围的文本
func (tb *TextBuffer) Replace(r Range, text string) error {
	tb.mutex.Lock()

	// 获取偏移量
	startOffset := tb.gapBuffer.GetOffsetAt(r.Start)
	endOffset := tb.gapBuffer.GetOffsetAt(r.End)

	if startOffset > endOffset {
		tb.mutex.Unlock()
		return errors.New("invalid range")
	}

	// 获取要替换的文本
	oldText := tb.gapBuffer.GetTextInRange(startOffset, endOffset)

	// 记录操作用于撤销
	tb.undoStack.Push(&TextOperation{
		Type:     OperationReplace,
		Position: r.Start,
		Text:     text,
		OldText:  oldText,
	})

	// 执行替换操作
	tb.gapBuffer.Delete(startOffset, endOffset)
	tb.gapBuffer.Insert(startOffset, text)

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      r.Start,
		Text:          text,
		OldText:       oldText,
		OperationType: OperationReplace,
	})

	return nil
}

// FindNext finds the next occurrence of a string
func (tb *TextBuffer) FindNext(searchText string, startPosition Position, caseSensitive, wholeWord, regex bool) (Range, error) {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	if searchText == "" {
		return Range{}, errors.New("search text cannot be empty")
	}

	// Get text after start position
	startOffset := tb.gapBuffer.GetOffsetAt(startPosition)
	text := tb.gapBuffer.GetText()
	searchArea := text[startOffset:]

	// Prepare search
	var searchIndex int
	var matchLength int

	if regex {
		// Use regex search
		var regexPattern string
		if wholeWord {
			regexPattern = fmt.Sprintf("\\b%s\\b", searchText)
		} else {
			regexPattern = searchText
		}

		re, err := regexp.Compile(regexPattern)
		if err != nil {
			return Range{}, fmt.Errorf("invalid regex pattern: %v", err)
		}

		if !caseSensitive {
			re, err = regexp.Compile("(?i)" + regexPattern)
			if err != nil {
				return Range{}, fmt.Errorf("invalid regex pattern: %v", err)
			}
		}

		match := re.FindStringIndex(searchArea)
		if match == nil {
			// Try from the beginning
			match = re.FindStringIndex(text)
			if match == nil {
				return Range{}, errors.New("text not found")
			}
			searchIndex = match[0]
		} else {
			searchIndex = startOffset + match[0]
		}
		matchLength = match[1] - match[0]
	} else {
		// Use simple string search
		var searchFunc func(string, string) int
		if caseSensitive {
			searchFunc = strings.Index
		} else {
			searchFunc = func(s, substr string) int {
				return strings.Index(strings.ToLower(s), strings.ToLower(substr))
			}
		}

		index := searchFunc(searchArea, searchText)
		if index == -1 {
			// Try from the beginning
			index = searchFunc(text, searchText)
			if index == -1 {
				return Range{}, errors.New("text not found")
			}
			searchIndex = index
		} else {
			searchIndex = startOffset + index
		}
		matchLength = len(searchText)

		// Check for whole word
		if wholeWord {
			// Check if the match is a whole word
			isWordBoundaryBefore := searchIndex == 0 || !isWordChar(rune(text[searchIndex-1]))
			isWordBoundaryAfter := searchIndex+matchLength >= len(text) || !isWordChar(rune(text[searchIndex+matchLength]))

			if !isWordBoundaryBefore || !isWordBoundaryAfter {
				// Not a whole word, try to find the next occurrence
				nextStartPosition := tb.gapBuffer.GetPositionAt(searchIndex + 1)
				return tb.FindNext(searchText, nextStartPosition, caseSensitive, wholeWord, regex)
			}
		}
	}

	// Convert offsets to positions
	startPos := tb.gapBuffer.GetPositionAt(searchIndex)
	endPos := tb.gapBuffer.GetPositionAt(searchIndex + matchLength)

	return Range{Start: startPos, End: endPos}, nil
}

// FindPrevious finds the previous occurrence of a string
func (tb *TextBuffer) FindPrevious(searchText string, startPosition Position, caseSensitive, wholeWord, regex bool) (Range, error) {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	if searchText == "" {
		return Range{}, errors.New("search text cannot be empty")
	}

	// Get text before start position
	startOffset := tb.gapBuffer.GetOffsetAt(startPosition)
	text := tb.gapBuffer.GetText()
	searchArea := text[:startOffset]

	// Prepare search
	var searchIndex int
	var matchLength int

	if regex {
		// Use regex search
		var regexPattern string
		if wholeWord {
			regexPattern = fmt.Sprintf("\\b%s\\b", searchText)
		} else {
			regexPattern = searchText
		}

		re, err := regexp.Compile(regexPattern)
		if err != nil {
			return Range{}, fmt.Errorf("invalid regex pattern: %v", err)
		}

		if !caseSensitive {
			re, err = regexp.Compile("(?i)" + regexPattern)
			if err != nil {
				return Range{}, fmt.Errorf("invalid regex pattern: %v", err)
			}
		}

		// Find all matches in the search area
		matches := re.FindAllStringIndex(searchArea, -1)
		if len(matches) == 0 {
			// Try from the end
			matches = re.FindAllStringIndex(text, -1)
			if len(matches) == 0 {
				return Range{}, errors.New("text not found")
			}
			// Get the last match
			lastMatch := matches[len(matches)-1]
			searchIndex = lastMatch[0]
			matchLength = lastMatch[1] - lastMatch[0]
		} else {
			// Get the last match
			lastMatch := matches[len(matches)-1]
			searchIndex = lastMatch[0]
			matchLength = lastMatch[1] - lastMatch[0]
		}
	} else {
		// Use simple string search
		var searchFunc func(string, string) int
		if caseSensitive {
			searchFunc = strings.LastIndex
		} else {
			searchFunc = func(s, substr string) int {
				return strings.LastIndex(strings.ToLower(s), strings.ToLower(substr))
			}
		}

		index := searchFunc(searchArea, searchText)
		if index == -1 {
			// Try from the end
			index = searchFunc(text, searchText)
			if index == -1 {
				return Range{}, errors.New("text not found")
			}
			searchIndex = index
		} else {
			searchIndex = index
		}
		matchLength = len(searchText)

		// Check for whole word
		if wholeWord {
			// Check if the match is a whole word
			isWordBoundaryBefore := searchIndex == 0 || !isWordChar(rune(text[searchIndex-1]))
			isWordBoundaryAfter := searchIndex+matchLength >= len(text) || !isWordChar(rune(text[searchIndex+matchLength]))

			if !isWordBoundaryBefore || !isWordBoundaryAfter {
				// Not a whole word, try to find the previous occurrence
				nextStartPosition := tb.gapBuffer.GetPositionAt(searchIndex)
				return tb.FindPrevious(searchText, nextStartPosition, caseSensitive, wholeWord, regex)
			}
		}
	}

	// Convert offsets to positions
	startPos := tb.gapBuffer.GetPositionAt(searchIndex)
	endPos := tb.gapBuffer.GetPositionAt(searchIndex + matchLength)

	return Range{Start: startPos, End: endPos}, nil
}

// ReplaceAll replaces all occurrences of a string
func (tb *TextBuffer) ReplaceAll(searchText, replaceText string, caseSensitive, wholeWord, regex bool) (int, error) {
	if searchText == "" {
		return 0, errors.New("search text cannot be empty")
	}

	tb.mutex.Lock()

	// Get text
	text := tb.gapBuffer.GetText()

	// Prepare search
	var newText string
	var count int

	if regex {
		// Use regex search
		var regexPattern string
		if wholeWord {
			regexPattern = fmt.Sprintf("\\b%s\\b", searchText)
		} else {
			regexPattern = searchText
		}

		re, err := regexp.Compile(regexPattern)
		if err != nil {
			tb.mutex.Unlock()
			return 0, fmt.Errorf("invalid regex pattern: %v", err)
		}

		if !caseSensitive {
			re, err = regexp.Compile("(?i)" + regexPattern)
			if err != nil {
				tb.mutex.Unlock()
				return 0, fmt.Errorf("invalid regex pattern: %v", err)
			}
		}

		// Replace all occurrences
		newText = re.ReplaceAllString(text, replaceText)
		count = strings.Count(newText, replaceText)
	} else {
		// Use simple string search
		if !caseSensitive {
			// Case-insensitive search
			searchTextLower := strings.ToLower(searchText)
			textLower := strings.ToLower(text)

			var lastIndex int
			var sb strings.Builder

			for {
				index := strings.Index(textLower[lastIndex:], searchTextLower)
				if index == -1 {
					// No more occurrences
					sb.WriteString(text[lastIndex:])
					break
				}

				index += lastIndex

				// Check for whole word
				if wholeWord {
					isWordBoundaryBefore := index == 0 || !isWordChar(rune(text[index-1]))
					isWordBoundaryAfter := index+len(searchText) >= len(text) || !isWordChar(rune(text[index+len(searchText)]))

					if !isWordBoundaryBefore || !isWordBoundaryAfter {
						// Not a whole word, skip this occurrence
						sb.WriteString(text[lastIndex : index+1])
						lastIndex = index + 1
						continue
					}
				}

				// Replace this occurrence
				sb.WriteString(text[lastIndex:index])
				sb.WriteString(replaceText)
				lastIndex = index + len(searchText)
				count++
			}

			newText = sb.String()
		} else {
			// Case-sensitive search
			if wholeWord {
				// Whole word search
				var lastIndex int
				var sb strings.Builder

				for {
					index := strings.Index(text[lastIndex:], searchText)
					if index == -1 {
						// No more occurrences
						sb.WriteString(text[lastIndex:])
						break
					}

					index += lastIndex

					// Check for whole word
					isWordBoundaryBefore := index == 0 || !isWordChar(rune(text[index-1]))
					isWordBoundaryAfter := index+len(searchText) >= len(text) || !isWordChar(rune(text[index+len(searchText)]))

					if !isWordBoundaryBefore || !isWordBoundaryAfter {
						// Not a whole word, skip this occurrence
						sb.WriteString(text[lastIndex : index+1])
						lastIndex = index + 1
						continue
					}

					// Replace this occurrence
					sb.WriteString(text[lastIndex:index])
					sb.WriteString(replaceText)
					lastIndex = index + len(searchText)
					count++
				}

				newText = sb.String()
			} else {
				// Simple search and replace
				newText = strings.ReplaceAll(text, searchText, replaceText)
				count = strings.Count(text, searchText)
			}
		}
	}

	if count == 0 {
		tb.mutex.Unlock()
		return 0, errors.New("text not found")
	}

	// Record operation for undo
	tb.undoStack.Push(&TextOperation{
		Type:     OperationReplace,
		Position: Position{Line: 0, Column: 0},
		Text:     newText,
		OldText:  text,
	})

	// Update text
	tb.gapBuffer.SetText(newText)

	// Set modified flag
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// Trigger event
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      Position{Line: 0, Column: 0},
		Text:          newText,
		OldText:       text,
		OperationType: OperationReplace,
	})

	return count, nil
}

// isWordChar checks if a character is a word character
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}

// SaveToFile 保存文本到文件
func (tb *TextBuffer) SaveToFile(filePath string) error {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	// 如果文件路径为空，使用当前文件路径
	if filePath == "" {
		if tb.filePath == "" {
			return errors.New("no file path specified")
		}
		filePath = tb.filePath
	}

	// 创建目录（如果不存在）
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp(dir, "temp_*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	tempFilePath := tempFile.Name()
	defer os.Remove(tempFilePath) // 确保在出错时删除临时文件

	// 获取文本大小
	textSize := tb.gapBuffer.GetLength()

	// 对于大文件，使用分块写入
	if textSize > largeTextThreshold {
		// 使用缓冲写入器
		bufWriter := bufio.NewWriterSize(tempFile, chunkSize)

		// 分块处理
		for offset := 0; offset < textSize; offset += chunkSize {
			end := offset + chunkSize
			if end > textSize {
				end = textSize
			}

			// 获取文本块
			chunk := tb.gapBuffer.GetTextChunk(offset, end)

			// 写入文本块
			if _, err := bufWriter.WriteString(chunk); err != nil {
				tempFile.Close()
				return fmt.Errorf("failed to write to file: %v", err)
			}
		}

		// 刷新缓冲区
		if err := bufWriter.Flush(); err != nil {
			tempFile.Close()
			return fmt.Errorf("failed to flush buffer: %v", err)
		}
	} else {
		// 对于小文件，直接写入
		text := tb.gapBuffer.GetText()

		// 根据EOL类型转换换行符
		if tb.eolType == EOLWindows {
			text = strings.ReplaceAll(text, "\n", "\r\n")
		} else if tb.eolType == EOLMac {
			text = strings.ReplaceAll(text, "\n", "\r")
		}

		if _, err := tempFile.WriteString(text); err != nil {
			tempFile.Close()
			return fmt.Errorf("failed to write to file: %v", err)
		}
	}

	// 关闭临时文件
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %v", err)
	}

	// 重命名临时文件为目标文件
	if err := os.Rename(tempFilePath, filePath); err != nil {
		return fmt.Errorf("failed to rename temporary file: %v", err)
	}

	// 更新文件路径
	if tb.filePath != filePath {
		oldPath := tb.filePath
		tb.filePath = filePath

		// 触发文件路径变更事件
		tb.triggerEvent(EventFilePathChanged, struct {
			OldPath string
			NewPath string
		}{
			OldPath: oldPath,
			NewPath: filePath,
		})
	}

	// 重置修改标志
	tb.modified = false

	// 触发修改标志变更事件
	tb.triggerEvent(EventModifiedChanged, false)

	return nil
}

// LoadFromFile 从文件加载文本
func (tb *TextBuffer) LoadFromFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// 获取文件大小
	fileSize := fileInfo.Size()

	// 检测文件类型和语言
	extension := strings.ToLower(filepath.Ext(filePath))
	languageID, _ := detectLanguageFromExtension(extension)

	// 锁定文本缓冲区
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// 对于大文件，使用分块加载
	if fileSize > int64(largeTextThreshold) {
		// 创建一个新的GapBuffer
		newBuffer := NewGapBuffer()

		// 使用缓冲读取器
		bufReader := bufio.NewReaderSize(file, chunkSize)

		// 分块读取文件
		buffer := make([]byte, chunkSize)
		var totalText strings.Builder
		totalText.Grow(int(fileSize))

		for {
			n, err := bufReader.Read(buffer)
			if err != nil && err != io.EOF {
				return fmt.Errorf("failed to read file: %v", err)
			}

			if n == 0 {
				break
			}

			// 将字节转换为字符串并添加到缓冲区
			chunk := string(buffer[:n])
			totalText.WriteString(chunk)
		}

		// 检测EOL类型
		text := totalText.String()
		tb.eolType = detectEOLType(text)

		// 标准化换行符为\n
		if tb.eolType == EOLWindows {
			text = strings.ReplaceAll(text, "\r\n", "\n")
		} else if tb.eolType == EOLMac {
			text = strings.ReplaceAll(text, "\r", "\n")
		}

		// 设置文本
		newBuffer.SetText(text)

		// 替换旧的GapBuffer
		tb.gapBuffer = newBuffer
	} else {
		// 对于小文件，直接读取
		content, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		// 转换为字符串
		text := string(content)

		// 检测EOL类型
		tb.eolType = detectEOLType(text)

		// 标准化换行符为\n
		if tb.eolType == EOLWindows {
			text = strings.ReplaceAll(text, "\r\n", "\n")
		} else if tb.eolType == EOLMac {
			text = strings.ReplaceAll(text, "\r", "\n")
		}

		// 设置文本
		tb.gapBuffer.SetText(text)
	}

	// 清空撤销/重做栈
	tb.undoStack.Clear()

	// 更新文件路径和语言ID
	oldPath := tb.filePath
	tb.filePath = filePath

	oldLanguageID := tb.languageID
	tb.languageID = languageID

	// 重置修改标志
	tb.modified = false

	// 触发事件
	if oldPath != filePath {
		tb.triggerEvent(EventFilePathChanged, struct {
			OldPath string
			NewPath string
		}{
			OldPath: oldPath,
			NewPath: filePath,
		})
	}

	if oldLanguageID != languageID {
		tb.triggerEvent(EventLanguageChanged, languageID)
	}

	tb.triggerEvent(EventModifiedChanged, false)

	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      Position{Line: 0, Column: 0},
		Text:          tb.gapBuffer.GetText(),
		OldText:       "",
		OperationType: OperationSetText,
	})

	return nil
}

// detectLanguageFromExtension detects the language ID from a file extension
func detectLanguageFromExtension(extension string) (string, bool) {
	// Common language mappings
	languageMap := map[string]string{
		".txt":   "plaintext",
		".md":    "markdown",
		".go":    "go",
		".js":    "javascript",
		".ts":    "typescript",
		".jsx":   "javascriptreact",
		".tsx":   "typescriptreact",
		".html":  "html",
		".css":   "css",
		".json":  "json",
		".xml":   "xml",
		".yaml":  "yaml",
		".yml":   "yaml",
		".py":    "python",
		".rb":    "ruby",
		".java":  "java",
		".c":     "c",
		".cpp":   "cpp",
		".h":     "c",
		".hpp":   "cpp",
		".cs":    "csharp",
		".php":   "php",
		".rs":    "rust",
		".swift": "swift",
		".lua":   "lua",
		".sh":    "shellscript",
		".bat":   "bat",
		".ps1":   "powershell",
	}

	if languageID, ok := languageMap[extension]; ok {
		return languageID, true
	}

	return "plaintext", false
}

// Undo 撤销上一次操作
func (tb *TextBuffer) Undo() error {
	tb.mutex.Lock()

	operation, err := tb.undoStack.Undo()
	if err != nil {
		tb.mutex.Unlock()
		return err
	}

	var newText string
	var position Position

	switch operation.Type {
	case OperationInsert:
		// 撤销插入操作，需要删除插入的文本
		startOffset := tb.gapBuffer.GetOffsetAt(operation.Position)
		endOffset := startOffset + len([]rune(operation.Text))
		newText = ""
		position = operation.Position
		tb.gapBuffer.Delete(startOffset, endOffset)
	case OperationDelete:
		// 撤销删除操作，需要重新插入删除的文本
		offset := tb.gapBuffer.GetOffsetAt(operation.Position)
		newText = operation.OldText
		position = operation.Position
		tb.gapBuffer.Insert(offset, operation.OldText)
	case OperationReplace:
		// 撤销替换操作，需要恢复原来的文本
		startOffset := tb.gapBuffer.GetOffsetAt(operation.Position)
		endOffset := startOffset + len([]rune(operation.Text))
		newText = operation.OldText
		position = operation.Position
		tb.gapBuffer.Delete(startOffset, endOffset)
		tb.gapBuffer.Insert(startOffset, operation.OldText)
	}

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      position,
		Text:          newText,
		OldText:       operation.OldText,
		OperationType: operation.Type,
	})

	return nil
}

// Redo 重做上一次撤销的操作
func (tb *TextBuffer) Redo() error {
	tb.mutex.Lock()

	operation, err := tb.undoStack.Redo()
	if err != nil {
		tb.mutex.Unlock()
		return err
	}

	var newText string
	var position Position

	switch operation.Type {
	case OperationInsert:
		// 重做插入操作
		offset := tb.gapBuffer.GetOffsetAt(operation.Position)
		newText = operation.Text
		position = operation.Position
		tb.gapBuffer.Insert(offset, operation.Text)
	case OperationDelete:
		// 重做删除操作
		startOffset := tb.gapBuffer.GetOffsetAt(operation.Position)
		endOffset := startOffset + len([]rune(operation.OldText))
		newText = ""
		position = operation.Position
		tb.gapBuffer.Delete(startOffset, endOffset)
	case OperationReplace:
		// 重做替换操作
		startOffset := tb.gapBuffer.GetOffsetAt(operation.Position)
		endOffset := startOffset + len([]rune(operation.OldText))
		newText = operation.Text
		position = operation.Position
		tb.gapBuffer.Delete(startOffset, endOffset)
		tb.gapBuffer.Insert(startOffset, operation.Text)
	}

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      position,
		Text:          newText,
		OldText:       operation.OldText,
		OperationType: operation.Type,
	})

	return nil
}

// Clear 清空文本缓冲区
func (tb *TextBuffer) Clear() {
	tb.mutex.Lock()

	// 获取旧文本
	oldText := tb.gapBuffer.GetText()

	// 清空文本
	tb.gapBuffer.Clear()

	// 清空撤销/重做栈
	tb.undoStack.Clear()

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      Position{Line: 0, Column: 0},
		Text:          "",
		OldText:       oldText,
		OperationType: OperationClear,
	})
}

// SetText 设置整个文本内容
func (tb *TextBuffer) SetText(text string) {
	tb.mutex.Lock()

	// 获取旧文本
	oldText := tb.gapBuffer.GetText()

	// 设置文本
	tb.gapBuffer.SetText(text)

	// 清空撤销/重做栈
	tb.undoStack.Clear()

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      Position{Line: 0, Column: 0},
		Text:          text,
		OldText:       oldText,
		OperationType: OperationSetText,
	})
}

// GetMemoryStats 获取内存使用统计信息
func (tb *TextBuffer) GetMemoryStats() MemoryStats {
	// 简化版本，只返回基本信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		CurrentUsage:   m.Alloc,
		PeakUsage:      m.Sys,
		TotalAllocated: m.TotalAlloc,
		Allocations:    m.Mallocs,
		Deallocations:  m.Frees,
		UptimeSeconds:  uint64(time.Since(time.Time{}).Seconds()),
	}
}

// Close 关闭TextBuffer，释放资源
func (tb *TextBuffer) Close() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// 清理事件监听器
	tb.eventListeners = nil

	// 清理缓冲区
	if tb.gapBuffer != nil {
		tb.gapBuffer.Close()
		tb.gapBuffer = nil
	}

	// 清理其他资源
	tb.luaPluginManager = nil
	tb.lspManager = nil
}

// 大文本处理优化方法

// InsertLargeText 插入大块文本，针对大文本优化
func (tb *TextBuffer) InsertLargeText(position Position, text string) error {
	if text == "" {
		return nil
	}

	tb.mutex.Lock()

	// 获取偏移量
	offset := tb.gapBuffer.GetOffsetAt(position)

	// 记录操作用于撤销（对于超大文本，可能需要特殊处理）
	if len(text) < maxGapSize {
		tb.undoStack.Push(&TextOperation{
			Type:     OperationInsert,
			Position: position,
			Text:     text,
			OldText:  "",
		})
	} else {
		// 对于超大文本，清空撤销栈以避免内存问题
		tb.undoStack.Clear()
	}

	// 使用优化的大文本插入方法
	tb.gapBuffer.InsertChunk(offset, text)

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      position,
		Text:          text,
		OldText:       "",
		OperationType: OperationInsert,
	})

	return nil
}

// DeleteLargeRange 删除大范围文本，针对大文本优化
func (tb *TextBuffer) DeleteLargeRange(r Range) error {
	tb.mutex.Lock()

	// 获取偏移量
	startOffset := tb.gapBuffer.GetOffsetAt(r.Start)
	endOffset := tb.gapBuffer.GetOffsetAt(r.End)

	if startOffset >= endOffset {
		tb.mutex.Unlock()
		return nil
	}

	// 获取要删除的文本
	var oldText string
	deleteSize := endOffset - startOffset

	// 对于超大范围，不保存完整的删除文本以避免内存问题
	if deleteSize < maxGapSize {
		oldText = tb.gapBuffer.GetTextInRange(startOffset, endOffset)

		// 记录操作用于撤销
		tb.undoStack.Push(&TextOperation{
			Type:     OperationDelete,
			Position: r.Start,
			Text:     oldText,
			OldText:  oldText,
		})
	} else {
		// 对于超大范围，只保存一个标记，并清空撤销栈
		oldText = fmt.Sprintf("[Large text: %d bytes]", deleteSize)
		tb.undoStack.Clear()
	}

	// 使用优化的大文本删除方法
	tb.gapBuffer.DeleteChunk(startOffset, endOffset)

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      r.Start,
		Text:          oldText,
		OldText:       oldText,
		OperationType: OperationDelete,
	})

	return nil
}

// ReplaceLargeRange 替换大范围文本，针对大文本优化
func (tb *TextBuffer) ReplaceLargeRange(r Range, text string) error {
	tb.mutex.Lock()

	// 获取偏移量
	startOffset := tb.gapBuffer.GetOffsetAt(r.Start)
	endOffset := tb.gapBuffer.GetOffsetAt(r.End)

	if startOffset > endOffset {
		tb.mutex.Unlock()
		return errors.New("invalid range")
	}

	// 获取要替换的文本
	var oldText string
	deleteSize := endOffset - startOffset
	insertSize := len([]rune(text))

	// 对于超大范围或超大文本，不保存完整的文本以避免内存问题
	if deleteSize < maxGapSize && insertSize < maxGapSize {
		oldText = tb.gapBuffer.GetTextInRange(startOffset, endOffset)

		// 记录操作用于撤销
		tb.undoStack.Push(&TextOperation{
			Type:     OperationReplace,
			Position: r.Start,
			Text:     text,
			OldText:  oldText,
		})
	} else {
		// 对于超大范围或超大文本，只保存一个标记，并清空撤销栈
		oldText = fmt.Sprintf("[Large text: %d bytes]", deleteSize)
		tb.undoStack.Clear()
	}

	// 使用优化的大文本替换方法
	tb.gapBuffer.ReplaceChunk(startOffset, endOffset, text)

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      r.Start,
		Text:          text,
		OldText:       oldText,
		OperationType: OperationReplace,
	})

	return nil
}

// GetTextInRangeByOffset 获取指定偏移量范围内的文本
func (tb *TextBuffer) GetTextInRangeByOffset(start, end int) string {
	if start < 0 || end > tb.gapBuffer.GetLength() || start >= end {
		return ""
	}
	return tb.gapBuffer.GetTextInRange(start, end)
}

// FindNextLarge 使用优化的搜索方法查找下一个匹配
func (tb *TextBuffer) FindNextLarge(searchText string, startPosition Position, caseSensitive, wholeWord, regex bool) (Range, error) {
	if searchText == "" {
		return Range{}, fmt.Errorf("search text cannot be empty")
	}

	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	// 如果需要正则表达式搜索，使用原始方法
	if regex {
		return tb.FindNext(searchText, startPosition, caseSensitive, wholeWord, regex)
	}

	// 获取起始偏移量
	startOffset := tb.gapBuffer.GetOffsetAt(startPosition)

	// 使用优化的向前搜索方法
	start, end := tb.gapBuffer.FindTextForward(searchText, startOffset, caseSensitive)
	if start == -1 {
		return Range{}, fmt.Errorf("pattern not found")
	}

	// 如果需要全词匹配，检查边界
	if wholeWord {
		// 检查前一个字符
		if start > 0 {
			prevChar := []rune(tb.GetTextInRangeByOffset(start-1, start))[0]
			if isWordChar(prevChar) {
				// 不是词的开始，继续搜索
				nextPos := tb.gapBuffer.GetPositionAt(end)
				return tb.FindNextLarge(searchText, nextPos, caseSensitive, wholeWord, regex)
			}
		}

		// 检查后一个字符
		if end < tb.gapBuffer.GetLength() {
			nextChar := []rune(tb.GetTextInRangeByOffset(end, end+1))[0]
			if isWordChar(nextChar) {
				// 不是词的结束，继续搜索
				nextPos := tb.gapBuffer.GetPositionAt(end)
				return tb.FindNextLarge(searchText, nextPos, caseSensitive, wholeWord, regex)
			}
		}
	}

	// 转换为位置范围
	startPos := tb.gapBuffer.GetPositionAt(start)
	endPos := tb.gapBuffer.GetPositionAt(end)

	return Range{Start: startPos, End: endPos}, nil
}

// FindPreviousLarge 使用优化的搜索方法查找上一个匹配
func (tb *TextBuffer) FindPreviousLarge(searchText string, startPosition Position, caseSensitive, wholeWord, regex bool) (Range, error) {
	if searchText == "" {
		return Range{}, fmt.Errorf("search text cannot be empty")
	}

	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	// 如果需要正则表达式搜索，使用原始方法
	if regex {
		return tb.FindPrevious(searchText, startPosition, caseSensitive, wholeWord, regex)
	}

	// 获取起始偏移量
	startOffset := tb.gapBuffer.GetOffsetAt(startPosition)

	// 使用优化的向后搜索方法
	start, end := tb.gapBuffer.FindTextBackward(searchText, startOffset, caseSensitive)
	if start == -1 {
		return Range{}, fmt.Errorf("pattern not found")
	}

	// 如果需要全词匹配，检查边界
	if wholeWord {
		// 检查前一个字符
		if start > 0 {
			prevChar := []rune(tb.GetTextInRangeByOffset(start-1, start))[0]
			if isWordChar(prevChar) {
				// 不是词的开始，继续搜索
				prevPos := tb.gapBuffer.GetPositionAt(start)
				return tb.FindPreviousLarge(searchText, prevPos, caseSensitive, wholeWord, regex)
			}
		}

		// 检查后一个字符
		if end < tb.gapBuffer.GetLength() {
			nextChar := []rune(tb.GetTextInRangeByOffset(end, end+1))[0]
			if isWordChar(nextChar) {
				// 不是词的结束，继续搜索
				prevPos := tb.gapBuffer.GetPositionAt(start)
				return tb.FindPreviousLarge(searchText, prevPos, caseSensitive, wholeWord, regex)
			}
		}
	}

	// 转换为位置范围
	startPos := tb.gapBuffer.GetPositionAt(start)
	endPos := tb.gapBuffer.GetPositionAt(end)

	return Range{Start: startPos, End: endPos}, nil
}

// ReplaceAllLarge 使用优化的方法替换所有匹配项
func (tb *TextBuffer) ReplaceAllLarge(searchText, replaceText string, caseSensitive, wholeWord, regex bool) (int, error) {
	if searchText == "" {
		return 0, fmt.Errorf("search text cannot be empty")
	}

	// 如果需要正则表达式搜索，使用原始方法
	if regex {
		return tb.ReplaceAll(searchText, replaceText, caseSensitive, wholeWord, regex)
	}

	tb.mutex.Lock()

	// 获取文本大小
	textSize := tb.gapBuffer.GetLength()

	// 对于超大文本，使用分块处理
	if textSize > largeTextThreshold {
		// 计数器
		count := 0

		// 创建一个新的缓冲区来构建结果
		var resultBuilder strings.Builder
		resultBuilder.Grow(textSize) // 预分配空间

		// 上次匹配结束位置
		lastEnd := 0

		// 分块处理
		chunkSize := 10 * 1024 * 1024 // 10MB
		for offset := 0; offset < textSize; offset += chunkSize {
			end := offset + chunkSize
			if end > textSize {
				end = textSize
			}

			// 获取当前块
			chunk := tb.gapBuffer.GetTextChunk(offset, end)

			// 如果不区分大小写，转换为小写进行搜索
			var searchChunk, searchPattern string
			if caseSensitive {
				searchChunk = chunk
				searchPattern = searchText
			} else {
				searchChunk = strings.ToLower(chunk)
				searchPattern = strings.ToLower(searchText)
			}

			// 在当前块中搜索
			var lastChunkEnd int
			for {
				// 确保不会超出范围
				if lastChunkEnd >= len(searchChunk) {
					break
				}

				index := strings.Index(searchChunk[lastChunkEnd:], searchPattern)
				if index == -1 {
					break
				}

				// 计算实际位置
				matchStart := lastChunkEnd + index
				matchEnd := matchStart + len(searchPattern)
				lastChunkEnd = matchEnd

				// 确保不会超出范围
				if matchEnd > len(searchChunk) {
					break
				}

				// 如果需要全词匹配，检查边界
				if wholeWord {
					isWholeWord := true

					// 检查前一个字符
					if matchStart > 0 {
						prevChar := rune(searchChunk[matchStart-1])
						if isWordChar(prevChar) {
							isWholeWord = false
						}
					}

					// 检查后一个字符
					if matchEnd < len(searchChunk) {
						nextChar := rune(searchChunk[matchEnd])
						if isWordChar(nextChar) {
							isWholeWord = false
						}
					}

					if !isWholeWord {
						continue
					}
				}

				// 添加匹配前的文本
				resultBuilder.WriteString(chunk[lastEnd-offset : matchStart])

				// 添加替换文本
				resultBuilder.WriteString(replaceText)

				// 更新位置和计数器
				lastEnd = offset + matchEnd
				count++
			}

			// 如果没有找到匹配，添加整个块
			if lastEnd <= offset {
				resultBuilder.WriteString(chunk)
			} else if lastEnd < offset+len(chunk) {
				// 添加块中剩余的部分
				resultBuilder.WriteString(chunk[lastEnd-offset:])
			}

			// 更新lastEnd以确保它不会小于当前块的结束位置
			if lastEnd < offset+len(chunk) {
				lastEnd = offset + len(chunk)
			}
		}

		// 如果没有替换，直接返回
		if count == 0 {
			tb.mutex.Unlock()
			return 0, nil
		}

		// 设置新文本
		newText := resultBuilder.String()

		// 清空撤销栈（对于大文本替换）
		tb.undoStack.Clear()

		// 设置新文本
		tb.gapBuffer.SetText(newText)

		// 设置修改标志
		tb.modified = true

		// 释放锁，避免死锁
		tb.mutex.Unlock()

		// 触发事件
		tb.triggerEvent(EventTextChanged, TextChangedEventData{
			Position:      Position{Line: 0, Column: 0},
			Text:          newText,
			OldText:       "",
			OperationType: OperationReplace,
		})

		return count, nil
	}

	// 对于小文本，使用原始方法
	originalText := tb.gapBuffer.GetText()

	// 计数器
	count := 0

	// 从头开始搜索
	position := Position{Line: 0, Column: 0}

	// 创建一个新的缓冲区来构建结果
	var resultBuilder strings.Builder
	resultBuilder.Grow(len(originalText)) // 预分配空间
	lastEnd := 0

	for {
		// 获取当前位置的偏移量
		offset := tb.gapBuffer.GetOffsetAt(position)

		// 搜索下一个匹配
		start, end := tb.gapBuffer.FindTextForward(searchText, offset, caseSensitive)
		if start == -1 {
			break
		}

		// 如果需要全词匹配，检查边界
		if wholeWord {
			isWholeWord := true

			// 检查前一个字符
			if start > 0 {
				prevChar := []rune(tb.GetTextInRangeByOffset(start-1, start))[0]
				if isWordChar(prevChar) {
					isWholeWord = false
				}
			}

			// 检查后一个字符
			if end < tb.gapBuffer.GetLength() {
				nextChar := []rune(tb.GetTextInRangeByOffset(end, end+1))[0]
				if isWordChar(nextChar) {
					isWholeWord = false
				}
			}

			if !isWholeWord {
				// 不是全词匹配，继续搜索
				position = tb.gapBuffer.GetPositionAt(end)
				continue
			}
		}

		// 添加匹配前的文本
		resultBuilder.WriteString(tb.GetTextInRangeByOffset(lastEnd, start))

		// 添加替换文本
		resultBuilder.WriteString(replaceText)

		// 更新位置和计数器
		position = tb.gapBuffer.GetPositionAt(end)
		lastEnd = end
		count++
	}

	// 添加剩余文本
	if lastEnd < tb.gapBuffer.GetLength() {
		resultBuilder.WriteString(tb.GetTextInRangeByOffset(lastEnd, tb.gapBuffer.GetLength()))
	}

	// 如果没有替换，直接返回
	if count == 0 {
		tb.mutex.Unlock()
		return 0, nil
	}

	// 设置新文本
	newText := resultBuilder.String()

	// 记录操作用于撤销
	tb.undoStack.Push(&TextOperation{
		Type:     OperationReplace,
		Position: Position{Line: 0, Column: 0},
		Text:     newText,
		OldText:  originalText,
	})

	// 设置新文本
	tb.gapBuffer.SetText(newText)

	// 设置修改标志
	tb.modified = true

	// 释放锁，避免死锁
	tb.mutex.Unlock()

	// 触发事件
	tb.triggerEvent(EventTextChanged, TextChangedEventData{
		Position:      Position{Line: 0, Column: 0},
		Text:          newText,
		OldText:       originalText,
		OperationType: OperationReplace,
	})

	return count, nil
}
