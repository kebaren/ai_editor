package editor

import (
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// SetupApplication sets up the application-level functionality
func SetupApplication(app *gtk.Application) {
	// Add application actions
	quitAction := gio.NewSimpleAction("quit", nil)
	quitAction.ConnectActivate(func(parameter *glib.Variant) {
		app.Quit()
	})
	app.AddAction(quitAction)

	// Set up keyboard accelerators
	app.SetAccelsForAction("app.quit", []string{"<Ctrl>q"})
	app.SetAccelsForAction("win.new", []string{"<Ctrl>n"})
	app.SetAccelsForAction("win.open", []string{"<Ctrl>o"})
	app.SetAccelsForAction("win.save", []string{"<Ctrl>s"})
	app.SetAccelsForAction("win.save-as", []string{"<Ctrl><Shift>s"})
	app.SetAccelsForAction("win.undo", []string{"<Ctrl>z"})
	app.SetAccelsForAction("win.redo", []string{"<Ctrl>y"})
	app.SetAccelsForAction("win.cut", []string{"<Ctrl>x"})
	app.SetAccelsForAction("win.copy", []string{"<Ctrl>c"})
	app.SetAccelsForAction("win.paste", []string{"<Ctrl>v"})
	app.SetAccelsForAction("win.find", []string{"<Ctrl>f"})
	app.SetAccelsForAction("win.replace", []string{"<Ctrl>h"})
}
