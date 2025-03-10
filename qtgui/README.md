# Orange Editor - Qt GUI

这是Orange编辑器的Qt GUI实现版本，使用GoQt框架构建。

## 功能特性

- 基于Qt的现代化GUI界面
- 使用自定义的gap buffer文本缓冲区实现
- 文件操作：新建、打开、保存、另存为
- 编辑功能：撤销、重做、剪切、复制、粘贴
- 搜索功能：支持区分大小写、全字匹配、正则表达式
- 状态栏显示当前光标位置

## 依赖项

- Go 1.16+
- [GoQt](https://github.com/therecipe/qt)（Qt绑定for Go）
- Qt 5.12+

## 安装

1. 安装Go（1.16或更高版本）
2. 安装Qt（5.12或更高版本）
3. 安装GoQt：

```bash
go get -u github.com/therecipe/qt/cmd/...
```

4. 安装项目依赖：

```bash
go mod download
```

## 构建

在项目根目录下运行：

```bash
# 生成Qt绑定
$(go env GOPATH)/bin/qtdeploy build desktop .

# 或者直接运行
go run qtgui/main.go
```

## 使用方法

### 基本操作

- **新建文件**：Ctrl+N 或 文件菜单 -> 新建
- **打开文件**：Ctrl+O 或 文件菜单 -> 打开
- **保存文件**：Ctrl+S 或 文件菜单 -> 保存
- **另存为**：Ctrl+Shift+S 或 文件菜单 -> 另存为
- **退出**：Ctrl+Q 或 文件菜单 -> 退出

### 编辑操作

- **撤销**：Ctrl+Z 或 编辑菜单 -> 撤销
- **重做**：Ctrl+Y 或 编辑菜单 -> 重做
- **剪切**：Ctrl+X 或 编辑菜单 -> 剪切
- **复制**：Ctrl+C 或 编辑菜单 -> 复制
- **粘贴**：Ctrl+V 或 编辑菜单 -> 粘贴

### 搜索操作

- **查找**：Ctrl+F 或 搜索菜单 -> 查找
- 在搜索对话框中，可以设置是否区分大小写、全字匹配、使用正则表达式
- 使用"查找下一个"和"查找上一个"按钮在文档中导航

## 项目结构

- `main.go` - 应用程序入口
- `editor/` - 编辑器相关组件
  - `window.go` - 主窗口实现
  - `buffer_adapter.go` - 缓冲区适配器，连接Qt文本编辑器和TextBuffer
  - `search_dialog.go` - 搜索对话框实现

## 许可证

MIT 