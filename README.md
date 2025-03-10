# GoTextBuffer 文本编辑器

一个基于自定义文本缓冲区实现的高性能文本编辑器，支持大文件处理、搜索替换、撤销重做等功能。

## 功能特点

- 高性能文本处理，支持大文件（GB级别）
- 基于Gap Buffer的高效文本编辑
- 撤销/重做功能
- 搜索和替换（支持正则表达式）
- 跨平台GUI界面（基于GTK4）
- 内存使用监控

## 项目结构

- `textbuffer/`: 核心文本缓冲区实现
  - `gap_buffer.go`: Gap Buffer实现
  - `text_buffer.go`: 文本缓冲区API
  - `position.go`: 位置和范围定义
  - `undo_stack.go`: 撤销/重做栈
  - `memory_monitor.go`: 内存监控
  - `memory_stats.go`: 内存统计
  - `memory_pool.go`: 内存池
  - `lua_plugin.go`: Lua插件系统
  - `lsp.go`: 语言服务器协议支持
  - `profiler.go`: 性能分析

- `gui/`: GUI界面实现
  - `main.go`: GUI应用程序入口
  - `editor/`: 编辑器组件
    - `window.go`: 主窗口实现
    - `buffer_adapter.go`: 缓冲区适配器
    - `shortcuts.go`: 键盘快捷键
    - `application.go`: 应用程序级功能

## 构建和运行

### 命令行版本

```bash
go build -o gotextbuffer.exe
./gotextbuffer.exe
```

### GUI版本

#### 安装依赖

首先，您需要安装GTK4：

**Windows**:
1. 安装MSYS2: https://www.msys2.org/
2. 打开MSYS2 MINGW64 shell并运行:
   ```
   pacman -S mingw-w64-x86_64-gtk4 mingw-w64-x86_64-pkgconf mingw-w64-x86_64-gcc
   ```

**Linux**:
```bash
# Ubuntu/Debian
sudo apt install libgtk-4-dev pkg-config gcc

# Fedora
sudo dnf install gtk4-devel pkgconf gcc
```

**macOS**:
```bash
brew install gtk4 pkg-config gcc
```

#### 构建和运行GUI版本

```bash
go build -o gui.exe
./gui.exe
```

## 使用方法

### GUI界面

1. 文件操作
   - 新建文件: Ctrl+N
   - 打开文件: Ctrl+O
   - 保存文件: Ctrl+S
   - 另存为: Ctrl+Shift+S

2. 编辑操作
   - 撤销: Ctrl+Z
   - 重做: Ctrl+Y
   - 剪切: Ctrl+X
   - 复制: Ctrl+C
   - 粘贴: Ctrl+V

3. 搜索和替换
   - 查找: Ctrl+F
   - 替换: Ctrl+H
   - 查找下一个: F3
   - 查找上一个: Shift+F3

## 性能测试

项目包含了性能测试，用于测试大文件处理能力：

```bash
cd textbuffer
go test -v -run TestLargeFilePerformance
```

这将测试以下操作的性能：
- 生成1GB测试数据
- 加载1GB文本
- 读取全部文本
- 随机位置插入1KB文本
- 随机位置删除1KB文本
- 搜索文本
- 替换所有匹配

## 许可证

MIT 