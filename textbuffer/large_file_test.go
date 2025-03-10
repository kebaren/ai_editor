package textbuffer

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

const (
	ONE_GB = 1024 * 1024 * 1024 // 1GB in bytes
)

// TestLargeFilePerformance 测试大文件性能
func TestLargeFilePerformance(t *testing.T) {
	// 生成1GB测试数据
	fmt.Println("生成1GB测试数据...")
	start := time.Now()
	testData := generateLargeText(ONE_GB)
	fmt.Printf("数据生成完成，耗时: %.2f秒\n", time.Since(start).Seconds())

	// 创建TextBuffer
	buffer := NewTextBufferWithText("")

	// 测试加载
	fmt.Println("\n测试加载1GB文本...")
	start = time.Now()
	buffer.SetText(testData)
	fmt.Printf("加载完成，耗时: %.2f秒\n", time.Since(start).Seconds())

	// 测试读取
	fmt.Println("\n测试读取全部文本...")
	start = time.Now()
	_ = buffer.GetText()
	fmt.Printf("读取完成，耗时: %.2f秒\n", time.Since(start).Seconds())

	// 测试随机插入
	fmt.Println("\n测试随机位置插入1KB文本...")
	insertText := strings.Repeat("X", 1024)
	randomOffset := rand.Intn(buffer.GetLength())
	pos := buffer.GetPositionAt(randomOffset)
	start = time.Now()
	buffer.Insert(pos, insertText)
	fmt.Printf("插入完成，耗时: %.2f秒\n", time.Since(start).Seconds())

	// 测试随机删除
	fmt.Println("\n测试随机位置删除1KB文本...")
	randomOffset = rand.Intn(buffer.GetLength() - 1024)
	startPos := buffer.GetPositionAt(randomOffset)
	endPos := buffer.GetPositionAt(randomOffset + 1024)
	start = time.Now()
	buffer.Delete(Range{Start: startPos, End: endPos})
	fmt.Printf("删除完成，耗时: %.2f秒\n", time.Since(start).Seconds())

	// 测试搜索
	fmt.Println("\n测试搜索文本...")
	searchText := "test"
	start = time.Now()
	_, err := buffer.FindNextLarge(searchText, Position{Line: 0, Column: 0}, true, false, false)
	fmt.Printf("搜索完成，耗时: %.2f秒\n", time.Since(start).Seconds())
	if err != nil {
		fmt.Printf("搜索结果: %v\n", err)
	}

	// 测试替换
	fmt.Println("\n测试替换所有匹配...")
	start = time.Now()
	count, err := buffer.ReplaceAllLarge("test", "TEST", true, false, false)
	fmt.Printf("替换完成，耗时: %.2f秒, 替换数量: %d\n", time.Since(start).Seconds(), count)
	if err != nil {
		fmt.Printf("替换结果: %v\n", err)
	}

	// 打印内存使用统计
	memStats := buffer.GetMemoryStats()
	fmt.Printf("\n内存使用统计:\n")
	fmt.Printf("当前使用: %.2f MB\n", float64(memStats.CurrentUsage)/(1024*1024))
	fmt.Printf("峰值使用: %.2f MB\n", float64(memStats.PeakUsage)/(1024*1024))
}

// generateLargeText 生成大文本用于测试
func generateLargeText(size int) string {
	const chunk = 1024 * 1024 // 1MB chunks
	var sb strings.Builder
	sb.Grow(size)

	// 生成一些随机单词作为基础文本
	words := []string{"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog", "test", "\n"}

	for sb.Len() < size {
		word := words[rand.Intn(len(words))]
		sb.WriteString(word + " ")

		// 每写入1MB检查一次大小
		if sb.Len() > chunk {
			if sb.Len() >= size {
				break
			}
		}
	}

	return sb.String()[:size]
}
