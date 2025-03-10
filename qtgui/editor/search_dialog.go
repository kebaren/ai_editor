package editor

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// SearchDialog 表示搜索对话框
type SearchDialog struct {
	dialog         *widgets.QDialog
	searchLineEdit *widgets.QLineEdit
	caseSensitive  *widgets.QCheckBox
	wholeWord      *widgets.QCheckBox
	useRegex       *widgets.QCheckBox
	findNextButton *widgets.QPushButton
	findPrevButton *widgets.QPushButton
	closeButton    *widgets.QPushButton
	onFindNext     func(text string, caseSensitive, wholeWord, useRegex bool)
	onFindPrevious func(text string, caseSensitive, wholeWord, useRegex bool)
}

// NewSearchDialog 创建一个新的搜索对话框
func NewSearchDialog(parent widgets.QWidget_ITF) *SearchDialog {
	dialog := widgets.NewQDialog(parent, core.Qt__Dialog)
	dialog.SetWindowTitle("搜索")
	dialog.SetMinimumWidth(400)
	dialog.SetModal(false)

	sd := &SearchDialog{
		dialog: dialog,
	}
	sd.setupUI()
	return sd
}

// setupUI 设置对话框UI
func (sd *SearchDialog) setupUI() {
	// 主布局
	layout := widgets.NewQVBoxLayout2(sd.dialog)

	// 搜索输入区域
	formLayout := widgets.NewQFormLayout(nil)
	sd.searchLineEdit = widgets.NewQLineEdit(nil)
	formLayout.AddRow3("查找内容:", sd.searchLineEdit)
	layout.AddLayout(formLayout, 0)

	// 选项区域
	optionsGroup := widgets.NewQGroupBox2("选项", nil)
	optionsLayout := widgets.NewQVBoxLayout2(optionsGroup)

	sd.caseSensitive = widgets.NewQCheckBox2("区分大小写", nil)
	sd.wholeWord = widgets.NewQCheckBox2("全字匹配", nil)
	sd.useRegex = widgets.NewQCheckBox2("使用正则表达式", nil)

	optionsLayout.AddWidget(sd.caseSensitive, 0, 0)
	optionsLayout.AddWidget(sd.wholeWord, 0, 0)
	optionsLayout.AddWidget(sd.useRegex, 0, 0)

	layout.AddWidget(optionsGroup, 0, 0)

	// 按钮区域
	buttonLayout := widgets.NewQHBoxLayout()

	sd.findNextButton = widgets.NewQPushButton2("查找下一个", nil)
	sd.findPrevButton = widgets.NewQPushButton2("查找上一个", nil)
	sd.closeButton = widgets.NewQPushButton2("关闭", nil)

	buttonLayout.AddWidget(sd.findNextButton, 0, 0)
	buttonLayout.AddWidget(sd.findPrevButton, 0, 0)
	buttonLayout.AddStretch(1)
	buttonLayout.AddWidget(sd.closeButton, 0, 0)

	layout.AddLayout(buttonLayout, 0)

	// 连接信号
	sd.connectSignals()
}

// connectSignals 连接对话框信号
func (sd *SearchDialog) connectSignals() {
	sd.findNextButton.ConnectClicked(func(checked bool) {
		if sd.onFindNext != nil {
			text := sd.searchLineEdit.Text()
			sd.onFindNext(
				text,
				sd.caseSensitive.IsChecked(),
				sd.wholeWord.IsChecked(),
				sd.useRegex.IsChecked(),
			)
		}
	})

	sd.findPrevButton.ConnectClicked(func(checked bool) {
		if sd.onFindPrevious != nil {
			text := sd.searchLineEdit.Text()
			sd.onFindPrevious(
				text,
				sd.caseSensitive.IsChecked(),
				sd.wholeWord.IsChecked(),
				sd.useRegex.IsChecked(),
			)
		}
	})

	sd.closeButton.ConnectClicked(func(checked bool) {
		sd.dialog.Hide()
	})

	// 按下回车键时查找下一个
	sd.searchLineEdit.ConnectReturnPressed(func() {
		sd.findNextButton.Click()
	})
}

// SetOnFindNext 设置查找下一个的回调函数
func (sd *SearchDialog) SetOnFindNext(callback func(text string, caseSensitive, wholeWord, useRegex bool)) {
	sd.onFindNext = callback
}

// SetOnFindPrevious 设置查找上一个的回调函数
func (sd *SearchDialog) SetOnFindPrevious(callback func(text string, caseSensitive, wholeWord, useRegex bool)) {
	sd.onFindPrevious = callback
}

// Show 显示对话框
func (sd *SearchDialog) Show() {
	sd.dialog.Show()
	sd.searchLineEdit.SetFocus2()

	// 尝试获取文本编辑器中的选中文本
	activeWindow := widgets.QApplication_ActiveWindow()
	if activeWindow != nil {
		// 使用类型断言前先转换为interface{}
		if mainWindow, ok := interface{}(activeWindow).(*widgets.QMainWindow); ok {
			if centralWidget := mainWindow.CentralWidget(); centralWidget != nil {
				// 简化实现，不再尝试查找文本编辑器
				// 在实际应用中，可以通过其他方式获取文本编辑器的引用
			}
		}
	}
}

// Hide 隐藏对话框
func (sd *SearchDialog) Hide() {
	sd.dialog.Hide()
}

// GetSearchLineEdit 获取搜索输入框
func (sd *SearchDialog) GetSearchLineEdit() *widgets.QLineEdit {
	return sd.searchLineEdit
}

// GetCaseSensitiveCheckBox 获取区分大小写复选框
func (sd *SearchDialog) GetCaseSensitiveCheckBox() *widgets.QCheckBox {
	return sd.caseSensitive
}

// GetWholeWordCheckBox 获取全字匹配复选框
func (sd *SearchDialog) GetWholeWordCheckBox() *widgets.QCheckBox {
	return sd.wholeWord
}

// GetUseRegexCheckBox 获取使用正则表达式复选框
func (sd *SearchDialog) GetUseRegexCheckBox() *widgets.QCheckBox {
	return sd.useRegex
}
