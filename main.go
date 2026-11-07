// IdeaBoard is a quiet, minimal note-board for capturing scattered ideas
// quickly while reading. Open the app, double-click, type, done.
//
// This file is the entry point: it creates the MyGo native-UI window and
// wires the app lifecycle together. The board model, canvas, storage and
// settings live in the other packages as they are added.
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type app struct{}

// view builds the window's interface from the app state, each frame.
func (a *app) view(c *ui.Context) {
	ui.Column(c).Fill().Gap(24).Children(func() {
		ui.Text(c, "IdeaBoard").FontSize(34).Bold()
		ui.Text(c, "Double-click empty space to create a note.").
			FontSize(14).TextColor(c.Theme().TextMuted)
	})
}

func main() {
	a := &app{}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:       "IdeaBoard",
			Width:       900,
			Height:      600,
			Transparent: true,
			Opacity:     0.92,
			Frameless:   true,
			CornerRadius: 8,
			Content:     ui.View(a.view),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
