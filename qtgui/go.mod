module github.com/example/orange/qtgui

go 1.23.0

toolchain go1.23.4

require (
	github.com/example/gotextbuffer/textbuffer v0.0.0-00010101000000-000000000000
	//github.com/example/gotextbuffer v0.0.0
	github.com/therecipe/qt v0.0.0-20200904063919-c0c124a5770d
)

require (
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/therecipe/env_darwin_amd64_513 v0.0.0-20190626001412-d8e92e8db4d0 // indirect
	github.com/therecipe/env_linux_amd64_513 v0.0.0-20190626000307-e137a3934da6 // indirect
	github.com/therecipe/env_windows_amd64_513 v0.0.0-20190626000028-79ec8bd06fb2 // indirect
	github.com/therecipe/env_windows_amd64_513/Tools v0.0.0-20190626000028-79ec8bd06fb2 // indirect
	github.com/therecipe/qt/internal/binding/files/docs/5.12.0 v0.0.0-20200904063919-c0c124a5770d // indirect
	github.com/therecipe/qt/internal/binding/files/docs/5.13.0 v0.0.0-20200904063919-c0c124a5770d // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/tools v0.31.0 // indirect
)

replace github.com/example/gotextbuffer/textbuffer => ../textbuffer
