# GoTextBuffer

一个Go语言实现的文本缓冲区，参考VSCode的TextBuffer实现逻辑。

## 功能特性

- 高效的文本存储和操作
- 支持插入、删除、替换等基本文本操作
- 支持行和列的定位
- 支持撤销和重做操作
- 支持换行符管理，优化多行文本处理

## 实现方式

本项目使用Gap Buffer数据结构来存储和管理文本，主要包括以下组件：

1. **GapBuffer**: 基于Gap Buffer的文本缓冲区，在文本中维护一个"间隙"，使得在当前编辑位置附近的插入和删除操作可以在常数时间内完成
2. **TextBuffer**: 主要的文本缓冲区接口，封装了GapBuffer并提供撤销/重做功能
3. **Position**: 表示文本中的位置（行和列）
4. **Range**: 表示文本中的范围（起始位置和结束位置）
5. **UndoStack**: 撤销/重做栈，用于管理文本操作的历史记录

## 使用方法

```go
import "github.com/example/gotextbuffer/textbuffer"

// 创建一个新的文本缓冲区
buffer := textbuffer.NewTextBuffer()

// 插入文本
buffer.Insert(textbuffer.Position{Line: 0, Column: 0}, "Hello, World!")

// 获取文本
text := buffer.GetText()

// 删除文本
buffer.Delete(textbuffer.Range{
    Start: textbuffer.Position{Line: 0, Column: 0},
    End:   textbuffer.Position{Line: 0, Column: 5},
})

// 撤销操作
buffer.Undo()

// 重做操作
buffer.Redo()
```

## 安装

```
go get github.com/example/gotextbuffer
``` 