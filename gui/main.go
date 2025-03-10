package main

import (
	"log"
	"os"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/example/gotextbuffer/gui/editor"
)

func main() {
	// 创建应用程序
	app := gtk.NewApplication("com.example.gotextbuffer", 0)

	// 当应用程序被激活时创建窗口
	app.ConnectActivate(func() {
		// 设置应用程序级别的功能
		editor.SetupApplication(app)

		// 创建一个新的编辑器窗口
		win := editor.NewEditorWindow(app)

		// 显示窗口
		win.Show()
	})

	// 运行应用程序
	if code := app.Run(os.Args); code > 0 {
		log.Fatalf("Application exited with code %d", code)
	}
}
