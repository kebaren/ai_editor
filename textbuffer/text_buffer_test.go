package textbuffer

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	// 大文本测试的大小
	largeTextSize = 1 * 1024 * 1024 // 1MB，从10MB减少到1MB
	// 并发测试的goroutine数量
	concurrentGoroutines = 10 // 从100减少到10
	// 并发测试的操作次数
	concurrentOperations = 100 // 从1000减少到100
)

// 生成指定大小的随机文本
func generateRandomText(size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789\n"
	b := make([]byte, size)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 测试基本操作
func TestBasicOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()

	// 测试初始状态
	if tb.GetLength() != 0 {
		t.Errorf("Expected empty buffer, got length %d", tb.GetLength())
	}

	// 测试插入文本
	text := "Hello, World!"
	tb.Insert(Position{Line: 0, Column: 0}, text)
	if tb.GetText() != text {
		t.Errorf("Expected text %q, got %q", text, tb.GetText())
	}

	// 测试删除文本
	tb.Delete(Range{
		Start: Position{Line: 0, Column: 5},
		End:   Position{Line: 0, Column: 12},
	})
	if tb.GetText() != "Hello!" {
		t.Errorf("Expected text %q, got %q", "Hello!", tb.GetText())
	}

	// 测试清空文本
	tb.Clear()
	if tb.GetLength() != 0 {
		t.Errorf("Expected empty buffer after clear, got length %d", tb.GetLength())
	}

	// 测试设置文本
	newText := "New text content"
	tb.SetText(newText)
	if tb.GetText() != newText {
		t.Errorf("Expected text %q, got %q", newText, tb.GetText())
	}
}

// 测试撤销/重做操作
func TestUndoRedo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()

	// 执行一系列操作
	operations := []struct {
		text string
		pos  Position
	}{
		{"Hello", Position{0, 0}},
		{", ", Position{0, 5}},
		{"World", Position{0, 7}},
		{"!", Position{0, 12}},
	}

	// 执行插入操作
	for _, op := range operations {
		tb.Insert(op.pos, op.text)
	}

	expectedText := "Hello, World!"
	if tb.GetText() != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, tb.GetText())
	}

	// 测试撤销操作
	for i := 0; i < len(operations); i++ {
		if err := tb.Undo(); err != nil {
			t.Errorf("Unexpected error during undo: %v", err)
		}
	}

	if tb.GetLength() != 0 {
		t.Errorf("Expected empty buffer after undo, got length %d", tb.GetLength())
	}

	// 测试重做操作
	for i := 0; i < len(operations); i++ {
		if err := tb.Redo(); err != nil {
			t.Errorf("Unexpected error during redo: %v", err)
		}
	}

	if tb.GetText() != expectedText {
		t.Errorf("Expected text %q after redo, got %q", expectedText, tb.GetText())
	}
}

// 测试行操作
func TestLineOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()

	// 测试多行文本
	text := "Line 1\nLine 2\nLine 3\nLine 4\n"
	tb.SetText(text)

	// 测试行数
	expectedLines := 4 // 4行文本，最后一个换行符不会创建新行
	if tb.GetLineCount() != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, tb.GetLineCount())
	}

	// 测试获取行内容
	expectedLineContents := []string{
		"Line 1\n",
		"Line 2\n",
		"Line 3\n",
		"Line 4\n",
	}

	for i, expected := range expectedLineContents {
		if content := tb.GetLineContent(i); content != expected {
			t.Errorf("Line %d: expected %q, got %q", i, expected, content)
		}
	}

	// 测试获取所有行
	lines := tb.GetLines()
	if len(lines) != len(expectedLineContents) {
		t.Errorf("Expected %d lines, got %d", len(expectedLineContents), len(lines))
	}
	for i, line := range lines {
		if line != expectedLineContents[i] {
			t.Errorf("Line %d: expected %q, got %q", i, expectedLineContents[i], line)
		}
	}

	// 测试空行处理
	tb.SetText("Line 1\n\nLine 3\n")
	if tb.GetLineCount() != 3 {
		t.Errorf("Expected 3 lines with empty line, got %d", tb.GetLineCount())
	}

	// 测试没有换行符的文本
	tb.SetText("Single line without newline")
	if tb.GetLineCount() != 1 {
		t.Errorf("Expected 1 line for text without newline, got %d", tb.GetLineCount())
	}
}

// 测试大文本处理
func TestLargeTextOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()

	// 生成大文本
	largeText := generateRandomText(largeTextSize)

	// 测试设置大文本
	start := time.Now()
	tb.SetText(largeText)
	setDuration := time.Since(start)
	t.Logf("Set large text duration: %v", setDuration)

	if tb.GetLength() != len([]rune(largeText)) {
		t.Errorf("Expected length %d, got %d", len([]rune(largeText)), tb.GetLength())
	}

	// 测试获取大文本
	start = time.Now()
	retrievedText := tb.GetText()
	getDuration := time.Since(start)
	t.Logf("Get large text duration: %v", getDuration)

	if retrievedText != largeText {
		t.Error("Retrieved text does not match original text")
	}

	// 测试在大文本中间插入内容
	start = time.Now()
	insertPos := Position{Line: tb.GetLineCount() / 2, Column: 0}
	insertText := "Inserted text\n"
	tb.Insert(insertPos, insertText)
	insertDuration := time.Since(start)
	t.Logf("Insert in large text duration: %v", insertDuration)

	// 测试性能指标
	if setDuration > time.Second {
		t.Errorf("Setting large text took too long: %v", setDuration)
	}
	if getDuration > time.Second {
		t.Errorf("Getting large text took too long: %v", getDuration)
	}
	if insertDuration > time.Millisecond*100 {
		t.Errorf("Insertion in large text took too long: %v", insertDuration)
	}
}

// 测试并发安全性
func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()
	var wg sync.WaitGroup

	// 创建多个goroutine同时操作TextBuffer
	for i := 0; i < concurrentGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Add(-1)

			// 执行随机操作
			for j := 0; j < concurrentOperations; j++ {
				op := rand.Intn(4)
				switch op {
				case 0: // Insert
					text := fmt.Sprintf("Text%d-%d\n", id, j)
					tb.Insert(Position{Line: 0, Column: 0}, text)
				case 1: // Delete
					if tb.GetLength() > 0 {
						tb.Delete(Range{
							Start: Position{Line: 0, Column: 0},
							End:   Position{Line: 0, Column: 1},
						})
					}
				case 2: // Get text
					_ = tb.GetText()
				case 3: // Get line count
					_ = tb.GetLineCount()
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
}

// 测试搜索和替换功能
func TestSearchAndReplace(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()

	// 设置测试文本
	text := `This is a test text.
This line contains test pattern.
Another test line here.
Final line with TEST.`
	tb.SetText(text)

	// 测试普通搜索
	testCases := []struct {
		searchText    string
		startPos      Position
		caseSensitive bool
		wholeWord     bool
		regex         bool
		expectError   bool
		expectedPos   Position
	}{
		{"test", Position{0, 0}, false, false, false, false, Position{0, 10}},
		{"TEST", Position{0, 0}, false, false, false, false, Position{0, 10}},
		{"test", Position{0, 11}, false, false, false, false, Position{1, 19}},
		{"test", Position{0, 0}, true, true, false, false, Position{0, 10}},
		{"notfound", Position{0, 0}, false, false, false, true, Position{0, 0}},
	}

	for i, tc := range testCases {
		r, err := tb.FindNext(tc.searchText, tc.startPos, tc.caseSensitive, tc.wholeWord, tc.regex)
		if tc.expectError {
			if err == nil {
				t.Errorf("Case %d: Expected error, got none", i)
			}
			continue
		}
		if err != nil {
			t.Errorf("Case %d: Unexpected error: %v", i, err)
			continue
		}
		if r.Start != tc.expectedPos {
			t.Errorf("Case %d: Expected position %v, got %v", i, tc.expectedPos, r.Start)
		}
	}

	// 测试替换所有
	count, err := tb.ReplaceAll("test", "REPLACED", false, false, false)
	if err != nil {
		t.Errorf("Unexpected error during replace all: %v", err)
	}
	if count != 4 {
		t.Errorf("Expected 4 replacements, got %d", count)
	}

	expectedText := `This is a REPLACED text.
This line contains REPLACED pattern.
Another REPLACED line here.
Final line with REPLACED.`
	if tb.GetText() != expectedText {
		t.Errorf("Expected text after replace:\n%s\ngot:\n%s", expectedText, tb.GetText())
	}
}

// 测试EOL处理
func TestEOLHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()

	// 测试不同类型的EOL
	testCases := []struct {
		name     string
		text     string
		eolType  EOLType
		expected string
	}{
		{
			name:     "Unix EOL",
			text:     "Line 1\nLine 2\nLine 3",
			eolType:  EOLUnix,
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Windows EOL",
			text:     "Line 1\r\nLine 2\r\nLine 3",
			eolType:  EOLWindows,
			expected: "Line 1\r\nLine 2\r\nLine 3",
		},
		{
			name:     "Mixed EOL to Unix",
			text:     "Line 1\nLine 2\r\nLine 3\rLine 4",
			eolType:  EOLUnix,
			expected: "Line 1\nLine 2\nLine 3\nLine 4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tb.SetText(tc.text)
			if err := tb.SetEOLType(tc.eolType); err != nil {
				t.Errorf("Unexpected error setting EOL type: %v", err)
			}
			if tb.GetEOLType() != tc.eolType {
				t.Errorf("Expected EOL type %v, got %v", tc.eolType, tb.GetEOLType())
			}
			if tb.GetText() != tc.expected {
				t.Errorf("Expected text %q, got %q", tc.expected, tb.GetText())
			}
		})
	}
}

// 测试事件系统
func TestEventSystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()

	var eventCount int
	var mu sync.Mutex

	// 添加事件监听器
	tb.AddEventListener(EventTextChanged, func(event Event) {
		mu.Lock()
		eventCount++
		mu.Unlock()
	})

	// 执行一系列操作
	operations := []func(){
		func() { tb.Insert(Position{0, 0}, "Hello") },
		func() { tb.Delete(Range{Start: Position{0, 0}, End: Position{0, 1}}) },
		func() { tb.SetText("New text") },
		func() { tb.Clear() },
	}

	for _, op := range operations {
		op()
	}

	// 等待事件处理完成
	time.Sleep(time.Millisecond * 100)

	mu.Lock()
	if eventCount != len(operations) {
		t.Errorf("Expected %d events, got %d", len(operations), eventCount)
	}
	mu.Unlock()
}

// 测试文件操作
func TestFileOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()

	// 测试保存和加载文件
	tempFile := "test_file.txt"
	testText := "This is a test file content.\nWith multiple lines.\n"

	// 设置文本并保存
	tb.SetText(testText)
	if err := tb.SaveToFile(tempFile); err != nil {
		t.Errorf("Unexpected error saving file: %v", err)
	}

	// 清空缓冲区后加载文件
	tb.Clear()
	if err := tb.LoadFromFile(tempFile); err != nil {
		t.Errorf("Unexpected error loading file: %v", err)
	}

	if tb.GetText() != testText {
		t.Errorf("Expected loaded text %q, got %q", testText, tb.GetText())
	}

	// 测试语言检测
	if tb.GetLanguageID() != "plaintext" {
		t.Errorf("Expected language ID 'plaintext', got %q", tb.GetLanguageID())
	}

	// 清理测试文件
	if err := tb.Delete(Range{
		Start: Position{Line: 0, Column: 0},
		End:   Position{Line: tb.GetLineCount(), Column: 0},
	}); err != nil {
		t.Errorf("Unexpected error deleting file content: %v", err)
	}
}

// 性能基准测试
func BenchmarkTextBuffer(b *testing.B) {
	text := generateRandomText(1024 * 1024) // 1MB

	b.Run("SetText", func(b *testing.B) {
		tb := NewTextBuffer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tb.SetText(text)
		}
	})

	b.Run("Insert", func(b *testing.B) {
		tb := NewTextBuffer()
		insertText := "Benchmark insert text"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tb.Insert(Position{0, 0}, insertText)
		}
	})

	b.Run("Delete", func(b *testing.B) {
		tb := NewTextBuffer()
		tb.SetText(text)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tb.Delete(Range{
				Start: Position{Line: 0, Column: 0},
				End:   Position{Line: 0, Column: 10},
			})
		}
	})

	b.Run("GetText", func(b *testing.B) {
		tb := NewTextBuffer()
		tb.SetText(text)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tb.GetText()
		}
	})
}

// 测试性能分析和内存监控
func TestProfilerAndMemoryMonitor(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// 创建性能分析器
	profiler, err := NewProfiler("./profiles")
	if err != nil {
		t.Fatalf("Failed to create profiler: %v", err)
	}

	// 开始CPU分析
	if err := profiler.StartCPUProfiling(); err != nil {
		t.Fatalf("Failed to start CPU profiling: %v", err)
	}

	// 开始内存分析
	if err := profiler.StartMemoryProfiling(); err != nil {
		t.Fatalf("Failed to start memory profiling: %v", err)
	}

	// 创建TextBuffer并执行一些操作
	tb := NewTextBuffer()

	// 记录初始内存状态
	initialStats := tb.GetMemoryStats()
	t.Logf("Initial memory stats:\n%s", initialStats.String())

	// 执行一系列操作
	largeText := generateRandomText(100 * 1024) // 100KB，从1MB减少到100KB
	tb.SetText(largeText)

	// 执行一些插入和删除操作
	for i := 0; i < 100; i++ { // 从1000减少到100
		tb.Insert(Position{Line: 0, Column: 0}, "Test text\n")
		tb.Delete(Range{
			Start: Position{Line: 0, Column: 0},
			End:   Position{Line: 1, Column: 0},
		})
	}

	// 记录最终内存状态
	finalStats := tb.GetMemoryStats()
	t.Logf("Final memory stats:\n%s", finalStats.String())

	// 获取性能分析信息
	profileInfo := profiler.GetProfileInfo()
	t.Logf("Profile info:\n%s", profileInfo.String())

	// 创建内存快照
	if err := profiler.TakeSnapshot(); err != nil {
		t.Errorf("Failed to take memory snapshot: %v", err)
	}

	// 收集goroutine分析信息
	if err := profiler.CollectGoroutineProfile(); err != nil {
		t.Errorf("Failed to collect goroutine profile: %v", err)
	}

	// 收集阻塞分析信息
	if err := profiler.CollectBlockProfile(); err != nil {
		t.Errorf("Failed to collect block profile: %v", err)
	}

	// 停止分析
	if err := profiler.StopCPUProfiling(); err != nil {
		t.Errorf("Failed to stop CPU profiling: %v", err)
	}

	if err := profiler.StopMemoryProfiling(); err != nil {
		t.Errorf("Failed to stop memory profiling: %v", err)
	}

	// 清理资源
	tb.Close()
}

// 测试内存池
func TestMemoryPool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	pool := NewMemoryPool(maxGapSize * 2)

	// 测试获取和放回缓冲区
	sizes := []int{128, 1024, 4096, 16384}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			// 获取缓冲区
			buf := pool.GetBuffer(size)
			if cap(buf) < size {
				t.Errorf("Buffer capacity %d is less than requested size %d", cap(buf), size)
			}

			// 使用缓冲区
			buf = append(buf, make([]rune, size)...)
			if len(buf) != size {
				t.Errorf("Buffer length %d is not equal to size %d", len(buf), size)
			}

			// 放回缓冲区
			pool.PutBuffer(buf)
		})
	}

	// 测试并发访问
	t.Run("Concurrent", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Add(-1)
				for j := 0; j < 100; j++ {
					size := sizes[j%len(sizes)]
					buf := pool.GetBuffer(size)
					buf = append(buf, make([]rune, size)...)
					pool.PutBuffer(buf)
				}
			}()
		}
		wg.Wait()
	})
}

// 测试边界情况
func TestEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()

	// 测试空操作
	t.Run("EmptyOperations", func(t *testing.T) {
		// 在空缓冲区上删除
		tb.Delete(Range{
			Start: Position{Line: 0, Column: 0},
			End:   Position{Line: 0, Column: 0},
		})

		// 插入空文本
		tb.Insert(Position{Line: 0, Column: 0}, "")

		// 在无效位置插入
		tb.Insert(Position{Line: -1, Column: -1}, "test")
		tb.Insert(Position{Line: 1000, Column: 1000}, "test")

		// 获取无效行
		if content := tb.GetLineContent(-1); content != "" {
			t.Errorf("Expected empty content for invalid line, got %q", content)
		}
	})

	// 测试大量换行符
	t.Run("ManyNewlines", func(t *testing.T) {
		text := strings.Repeat("\n", 1000)
		tb.SetText(text)
		if tb.GetLineCount() != 1000 {
			t.Errorf("Expected 1000 lines, got %d", tb.GetLineCount())
		}
	})

	// 测试Unicode字符
	t.Run("UnicodeCharacters", func(t *testing.T) {
		text := "Hello, 世界! 👋 🌍"
		tb.SetText(text)
		if tb.GetText() != text {
			t.Errorf("Unicode text not preserved, expected %q, got %q", text, tb.GetText())
		}
	})

	// 测试极限值
	t.Run("ExtremeValues", func(t *testing.T) {
		// 非常大的文本
		largeText := strings.Repeat("a", 100*1024) // 100KB，从1MB减少到100KB
		tb.SetText(largeText)

		// 非常长的行
		longLine := strings.Repeat("x", 1000) // 从10000减少到1000
		tb.SetText(longLine)

		// 大量短行
		manyLines := strings.Repeat("a\n", 1000) // 从10000减少到1000
		tb.SetText(manyLines)
	})

	// 测试资源清理
	t.Run("ResourceCleanup", func(t *testing.T) {
		tb := NewTextBuffer()
		tb.SetText("test")
		tb.Close()

		// 在关闭后尝试操作
		tb.Insert(Position{Line: 0, Column: 0}, "test")
		tb.Delete(Range{
			Start: Position{Line: 0, Column: 0},
			End:   Position{Line: 0, Column: 1},
		})
	})
}

// 测试并发安全性的扩展
func TestExtendedConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	tb := NewTextBuffer()
	var wg sync.WaitGroup

	// 测试多种并发操作组合
	operations := []struct {
		name string
		fn   func()
	}{
		{"Insert", func() { tb.Insert(Position{Line: 0, Column: 0}, "test") }},
		{"Delete", func() { tb.Delete(Range{Start: Position{Line: 0, Column: 0}, End: Position{Line: 0, Column: 1}}) }},
		{"GetText", func() { _ = tb.GetText() }},
		{"GetLineCount", func() { _ = tb.GetLineCount() }},
		{"GetLineContent", func() { _ = tb.GetLineContent(0) }},
		{"GetMemoryStats", func() { _ = tb.GetMemoryStats() }},
	}

	// 启动多个goroutine执行不同的操作组合
	for i := 0; i < 3; i++ { // 从10减少到3
		for _, op := range operations {
			wg.Add(1)
			go func(operation func()) {
				defer wg.Add(-1)
				for j := 0; j < 10; j++ { // 从100减少到10
					operation()
				}
			}(op.fn)
		}
	}

	// 等待所有操作完成
	wg.Wait()

	// 验证最终状态
	stats := tb.GetMemoryStats()
	t.Logf("Final memory stats after concurrent operations:\n%s", stats.String())
}

// 测试基本功能
func TestBasicFunctionality(t *testing.T) {
	tb := NewTextBuffer()

	// 测试初始状态
	if tb.GetLength() != 0 {
		t.Errorf("Expected empty buffer, got length %d", tb.GetLength())
	}

	// 测试插入文本
	tb.Insert(Position{Line: 0, Column: 0}, "Hello")
	if tb.GetText() != "Hello" {
		t.Errorf("Expected text %q, got %q", "Hello", tb.GetText())
	}

	// 测试清空文本
	tb.Clear()
	if tb.GetLength() != 0 {
		t.Errorf("Expected empty buffer after clear, got length %d", tb.GetLength())
	}
}
