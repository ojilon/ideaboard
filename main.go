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
	mygo_ui "github.com/egoist/mygo/ui"

	"ideaboard/board"
	"ideaboard/ui"
)

type app struct {
	canvas *ui.Canvas
}

// view builds the window's interface from the app state, each frame.
func (a *app) view(c *mygo_ui.Context) {
	a.canvas.View(c)
}

func main() {
	b := board.NewBoard("Default")
	a := &app{canvas: ui.NewCanvas(b)}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:       "IdeaBoard",
			Width:       900,
			Height:      600,
			Transparent: true,
			Opacity:     0.92,
			Frameless:   true,
			CornerRadius: 8,
			Content:     mygo_ui.View(a.view),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
