package main

import (
	"os"
	"path/filepath"

	"github.com/example/gotextbuffer/textbuffer"
	"github.com/example/orange/qtgui/editor"
	"github.com/therecipe/qt/widgets"
)

func main() {
	// 创建Qt应用程序
	app := widgets.NewQApplication(len(os.Args), os.Args)
	app.SetApplicationName("Orange Editor")
	app.SetApplicationDisplayName("Orange Editor")
	app.SetApplicationVersion("1.0.0")

	// 创建文本缓冲区
	buffer := textbuffer.NewTextBuffer()

	// 创建编辑器窗口
	window := editor.NewEditorWindow(nil)
	window.SetTextBuffer(buffer)
	window.Show()

	// 如果命令行参数中有文件路径，则打开该文件
	if len(os.Args) > 1 {
		filePath := os.Args[1]
		absPath, err := filepath.Abs(filePath)
		if err == nil {
			window.OpenFile(absPath)
		}
	}

	// 运行应用程序
	os.Exit(app.Exec())
}
