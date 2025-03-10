module github.com/example/gotextbuffer

go 1.21

require github.com/example/gotextbuffer/gui v0.0.0-00010101000000-000000000000

require (
	github.com/example/gotextbuffer/textbuffer v0.0.0-00010101000000-000000000000 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20231121144256-b99613f794b6 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
)

replace github.com/example/gotextbuffer/textbuffer => ./textbuffer

replace github.com/example/gotextbuffer/gui => ./gui
