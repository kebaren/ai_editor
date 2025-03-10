package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/gotextbuffer/textbuffer"
)

func main() {
	// 创建一个新的TextBuffer
	buffer := textbuffer.NewTextBuffer()

	// 演示基本功能
	fmt.Println("=== 基本功能演示 ===")
	demoBasicFeatures(buffer)

	// 演示EOL功能
	fmt.Println("\n=== EOL功能演示 ===")
	demoEOLFeatures(buffer)

	// 演示增强的文本编辑器API
	fmt.Println("\n=== 增强的文本编辑器API演示 ===")
	demoEnhancedAPI(buffer)

	// 演示Lua插件系统
	fmt.Println("\n=== Lua插件系统演示 ===")
	demoLuaPluginSystem(buffer)

	// 演示LSP系统
	fmt.Println("\n=== LSP系统演示 ===")
	demoLSPSystem(buffer)
}

// 演示基本功能
func demoBasicFeatures(buffer *textbuffer.TextBuffer) {
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
}

// 演示EOL功能
func demoEOLFeatures(buffer *textbuffer.TextBuffer) {
	// 清空缓冲区
	buffer.Clear()

	// 插入多行文本，使用Unix风格的换行符
	err := buffer.Insert(textbuffer.Position{Line: 0, Column: 0}, "Line 1\nLine 2\nLine 3")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 获取当前EOL类型
	eolType := buffer.GetEOLType()
	fmt.Println("Current EOL type:", eolTypeToString(eolType))

	// 获取EOL字符串
	eolString := buffer.GetEOLString()
	fmt.Printf("EOL string: %q\n", eolString)

	// 转换为Windows风格的换行符
	err = buffer.SetEOLType(textbuffer.EOLWindows)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 获取当前EOL类型
	eolType = buffer.GetEOLType()
	fmt.Println("EOL type after conversion:", eolTypeToString(eolType))

	// 获取EOL字符串
	eolString = buffer.GetEOLString()
	fmt.Printf("EOL string after conversion: %q\n", eolString)

	// 打印文本内容
	fmt.Println("Text with Windows-style line endings:", buffer.GetText())
}

// 演示增强的文本编辑器API
func demoEnhancedAPI(buffer *textbuffer.TextBuffer) {
	// 清空缓冲区
	buffer.Clear()

	// 插入示例文本
	err := buffer.Insert(textbuffer.Position{Line: 0, Column: 0}, "This is a sample text.\nIt contains multiple lines.\nWe can search and replace text in it.")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 打印文本内容
	fmt.Println("Original text:", buffer.GetText())

	// 设置语言ID
	buffer.SetLanguageID("markdown")
	fmt.Println("Language ID:", buffer.GetLanguageID())

	// 设置文件路径
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, "sample.md")
	buffer.SetFilePath(filePath)
	fmt.Println("File path:", buffer.GetFilePath())

	// 查找文本
	searchText := "sample"
	fmt.Printf("Finding text %q:\n", searchText)
	foundRange, err := buffer.FindNext(searchText, textbuffer.Position{Line: 0, Column: 0}, true, false, false)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Found at: Line %d, Column %d to Line %d, Column %d\n",
			foundRange.Start.Line, foundRange.Start.Column,
			foundRange.End.Line, foundRange.End.Column)
	}

	// 替换所有文本
	searchText = "text"
	replaceText := "content"
	fmt.Printf("Replacing all occurrences of %q with %q:\n", searchText, replaceText)
	count, err := buffer.ReplaceAll(searchText, replaceText, true, false, false)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Replaced %d occurrences\n", count)
		fmt.Println("Text after replace:", buffer.GetText())
	}

	// 保存到文件
	fmt.Println("Saving to file:", filePath)
	err = buffer.SaveToFile(filePath)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("File saved successfully")
		fmt.Println("Modified flag:", buffer.IsModified())
	}

	// 从文件加载
	fmt.Println("Loading from file:", filePath)
	err = buffer.LoadFromFile(filePath)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("File loaded successfully")
		fmt.Println("Text after load:", buffer.GetText())
	}

	// 删除临时文件
	os.Remove(filePath)
}

// 演示Lua插件系统
func demoLuaPluginSystem(buffer *textbuffer.TextBuffer) {
	// 获取Lua插件管理器
	pluginManager := buffer.GetLuaPluginManager()
	fmt.Println("Lua plugin manager initialized")

	// 打印插件管理器的地址，以确保变量被使用
	fmt.Printf("Plugin manager address: %p\n", pluginManager)

	// 在实际应用中，这里会加载和运行Lua插件
	fmt.Println("In a real application, Lua plugins would be loaded and executed here")
	fmt.Println("Example Lua plugin API functions:")
	fmt.Println("- textbuffer.get_text()")
	fmt.Println("- textbuffer.get_line_count()")
	fmt.Println("- textbuffer.get_line_content(line_index)")
	fmt.Println("- textbuffer.insert(line, column, text)")
	fmt.Println("- textbuffer.delete(start_line, start_column, end_line, end_column)")
	fmt.Println("- textbuffer.replace(start_line, start_column, end_line, end_column, text)")
	fmt.Println("- textbuffer.undo()")
	fmt.Println("- textbuffer.redo()")
	fmt.Println("- textbuffer.get_eol_type()")
	fmt.Println("- textbuffer.set_eol_type(eol_type)")

	// 打印插件管理器的方法
	fmt.Println("\nPlugin manager methods:")
	fmt.Println("- LoadPlugin(path)")
	fmt.Println("- UnloadPlugin(name)")
	fmt.Println("- EnablePlugin(name)")
	fmt.Println("- DisablePlugin(name)")
	fmt.Println("- GetPlugin(name)")
	fmt.Println("- GetPlugins()")
	fmt.Println("- CallPluginFunction(pluginName, functionName, ...args)")
}

// 演示LSP系统
func demoLSPSystem(buffer *textbuffer.TextBuffer) {
	// 获取LSP管理器
	lspManager := buffer.GetLSPManager()
	fmt.Println("LSP manager initialized")

	// 打印LSP管理器的地址，以确保变量被使用
	fmt.Printf("LSP manager address: %p\n", lspManager)

	// 在实际应用中，这里会启动和连接到LSP服务器
	fmt.Println("In a real application, LSP servers would be started and connected here")
	fmt.Println("Example LSP server commands:")
	fmt.Println("- For Go: gopls")
	fmt.Println("- For JavaScript/TypeScript: typescript-language-server --stdio")
	fmt.Println("- For Python: pyls")
	fmt.Println("- For Rust: rust-analyzer")

	// 打印LSP管理器的方法
	fmt.Println("\nLSP manager methods:")
	fmt.Println("- StartServer(languageID, command, ...args)")
	fmt.Println("- StopServer(languageID)")
	fmt.Println("- GetServer(languageID)")
	fmt.Println("- GetServers()")

	// 打印LSP服务器的方法
	fmt.Println("\nLSP server methods:")
	fmt.Println("- DidOpen(uri, languageID, text)")
	fmt.Println("- DidChange(uri, version, changes)")
	fmt.Println("- DidClose(uri)")
	fmt.Println("- Completion(uri, line, character)")
	fmt.Println("- Hover(uri, line, character)")
	fmt.Println("- Definition(uri, line, character)")
	fmt.Println("- References(uri, line, character, includeDeclaration)")
	fmt.Println("- DocumentSymbol(uri)")
	fmt.Println("- Formatting(uri, tabSize, insertSpaces)")
	fmt.Println("- RangeFormatting(uri, startLine, startCharacter, endLine, endCharacter, tabSize, insertSpaces)")
	fmt.Println("- CodeAction(uri, startLine, startCharacter, endLine, endCharacter, diagnostics, only)")
	fmt.Println("- Rename(uri, line, character, newName)")
}

// 辅助函数：将EOL类型转换为字符串
func eolTypeToString(eolType textbuffer.EOLType) string {
	switch eolType {
	case textbuffer.EOLUnix:
		return "Unix (\\n)"
	case textbuffer.EOLWindows:
		return "Windows (\\r\\n)"
	case textbuffer.EOLMac:
		return "Mac (\\r)"
	default:
		return "Unknown"
	}
}
