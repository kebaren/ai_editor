# GoTextBuffer Editor

一个基于GTK4的文本编辑器，使用gotk4框架和自定义的textbuffer实现。

## 功能特点

- 打开和保存文本文件
- 撤销/重做功能
- 搜索和替换（支持区分大小写、全字匹配、正则表达式）
- 大文件支持
- 状态栏显示光标位置和文件状态

## 系统要求

- Go 1.21 或更高版本
- GTK 4.0 或更高版本
- gotk4 依赖项

## 安装指南

### 安装 GTK4

#### Windows

1. 从 https://www.msys2.org/ 安装 MSYS2
2. 打开 MSYS2 MINGW64 shell 并运行：
   ```
   pacman -S mingw-w64-x86_64-gtk4 mingw-w64-x86_64-pkgconf mingw-w64-x86_64-gcc
   ```

#### Linux

```bash
# Ubuntu/Debian
sudo apt install libgtk-4-dev pkg-config gcc

# Fedora
sudo dnf install gtk4-devel pkgconf gcc
```

#### macOS

```bash
brew install gtk4 pkg-config gcc
```

### 安装 Go 依赖项

```bash
cd gui
go mod tidy
```

## 构建和运行

```bash
cd gui
go build -o editor.exe
./editor.exe
```

## 使用说明

### 文件操作

- **新建文件**：创建一个新的空文件
- **打开文件**：打开一个现有文件
- **保存**：保存当前文件
- **另存为**：使用新名称保存当前文件

### 编辑

- **撤销**：撤销上一次操作
- **重做**：重做上一次撤销的操作
- **剪切/复制/粘贴**：标准剪贴板操作

### 搜索和替换

- **查找**：在文档中搜索文本
- **替换**：在文档中替换文本

## 实现细节

该应用程序使用自定义的textbuffer实现，提供高效的文本处理，特别是对大文件的支持。GUI使用GTK4通过gotk4框架构建，该框架为GTK提供Go绑定。

主要组件包括：

- **TextBuffer**：核心文本处理组件
- **EditorWindow**：主应用程序窗口
- **BufferAdapter**：处理GTK文本缓冲区和自定义textbuffer之间的集成

## 界面风格

编辑器界面采用Windows风格设计，包括：

- 菜单栏和工具栏采用浅灰色背景
- 使用Consolas等等宽字体
- 状态栏显示详细信息
- 按钮和输入框采用Windows风格的边框和悬停效果

## 开发者信息

本项目是一个开源项目，欢迎贡献代码和提出建议。 