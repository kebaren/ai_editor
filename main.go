package main

import (
	"fmt"

	"github.com/example/gotextbuffer/textbuffer"
)

func main() {
	// 创建一个新的TextBuffer
	buffer := textbuffer.NewTextBuffer()

	// 插入文本
	err := buffer.Insert(textbuffer.Position{Line: 0, Column: 0}, "Hello, World!")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("Text after insert:", buffer.GetText())

	// 在指定位置插入文本
	err = buffer.Insert(textbuffer.Position{Line: 0, Column: 7}, " Go")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("Text after second insert:", buffer.GetText())

	// 删除文本
	err = buffer.Delete(textbuffer.Range{
		Start: textbuffer.Position{Line: 0, Column: 7},
		End:   textbuffer.Position{Line: 0, Column: 10},
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("Text after delete:", buffer.GetText())

	// 撤销操作
	err = buffer.Undo()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("Text after undo:", buffer.GetText())

	// 重做操作
	err = buffer.Redo()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("Text after redo:", buffer.GetText())

	// 替换文本
	err = buffer.Replace(textbuffer.Range{
		Start: textbuffer.Position{Line: 0, Column: 0},
		End:   textbuffer.Position{Line: 0, Column: 5},
	}, "Hi")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("Text after replace:", buffer.GetText())

	// 获取行数
	fmt.Println("Line count:", buffer.GetLineCount())

	// 获取文本长度
	fmt.Println("Text length:", buffer.GetLength())

	// 插入多行文本
	err = buffer.Insert(textbuffer.Position{Line: 0, Column: buffer.GetLength()}, "\nThis is a new line.\nAnd another line.")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("Text after multi-line insert:", buffer.GetText())

	// 获取行数
	fmt.Println("Line count after multi-line insert:", buffer.GetLineCount())

	// 获取指定行的内容
	fmt.Println("Line 1 content:", buffer.GetLineContent(1))

	// 获取所有行的内容
	lines := buffer.GetLines()
	fmt.Println("All lines:")
	for i, line := range lines {
		fmt.Printf("Line %d: %s\n", i, line)
	}

	// 演示换行符管理
	fmt.Println("\n--- 换行符管理演示 ---")

	// 创建一个新的TextBuffer，包含多行文本
	multiLineBuffer := textbuffer.NewTextBufferWithText("Line 1\nLine 2\nLine 3")

	// 打印文本内容
	fmt.Println("Original text:")
	fmt.Println(multiLineBuffer.GetText())

	// 获取行数
	fmt.Println("Line count:", multiLineBuffer.GetLineCount())

	// 获取指定行的内容
	fmt.Println("Line 0:", multiLineBuffer.GetLineContent(0))
	fmt.Println("Line 1:", multiLineBuffer.GetLineContent(1))
	fmt.Println("Line 2:", multiLineBuffer.GetLineContent(2))

	// 在行尾插入文本
	err = multiLineBuffer.Insert(textbuffer.Position{Line: 0, Column: 6}, " modified")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("\nAfter inserting at line end:")
	fmt.Println(multiLineBuffer.GetText())

	// 在行首插入文本
	err = multiLineBuffer.Insert(textbuffer.Position{Line: 1, Column: 0}, "Modified ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("\nAfter inserting at line start:")
	fmt.Println(multiLineBuffer.GetText())

	// 跨行删除
	err = multiLineBuffer.Delete(textbuffer.Range{
		Start: textbuffer.Position{Line: 0, Column: 5},
		End:   textbuffer.Position{Line: 1, Column: 5},
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("\nAfter cross-line delete:")
	fmt.Println(multiLineBuffer.GetText())

	// 插入多个换行符
	err = multiLineBuffer.Insert(textbuffer.Position{Line: 0, Column: 4}, "\n\n")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("\nAfter inserting multiple line breaks:")
	fmt.Println(multiLineBuffer.GetText())

	// 获取行数
	fmt.Println("Line count:", multiLineBuffer.GetLineCount())
}
