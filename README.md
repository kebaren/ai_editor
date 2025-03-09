# GoTextBuffer

一个Go语言实现的文本缓冲区，参考VSCode的TextBuffer实现逻辑。

## 功能特性

- 高效的文本存储和操作
- 支持插入、删除、替换等基本文本操作
- 支持行和列的定位
- 支持撤销和重做操作
- 支持文本分块存储，优化大文件处理

## 使用方法

```go
import "github.com/example/gotextbuffer/textbuffer"

// 创建一个新的文本缓冲区
buffer := textbuffer.NewTextBuffer()

// 插入文本
buffer.Insert(0, 0, "Hello, World!")

// 获取文本
text := buffer.GetText()

// 删除文本
buffer.Delete(0, 0, 5)
```

## 安装

```
go get github.com/example/gotextbuffer
``` 