// IdeaBoard is a quiet, minimal note-board for capturing scattered ideas
// quickly while reading. Open the app, double-click, type, done.
//
// This file is the entry point: it creates the MyGo native-UI window, loads
// and saves boards through the storage package, and runs the app lifecycle.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	mygo_ui "github.com/egoist/mygo/ui"

	"ideaboard/board"
	"ideaboard/storage"
	"ideaboard/ui"
)

type app struct {
	store  *storage.Store
	board  *board.Board
	canvas *ui.Canvas
}

func main() {
	a := &app{}
	mygo.App.WhenReady(func() {
		a.start()
	})
	mygo.App.OnWillQuit(func(e *mygo.QuitEvent) {
		a.saveNow()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *app) start() {
	a.store = storage.Open(boardsDir())
	boards, _ := a.store.ListBoards()
	name := "Default"
	if len(boards) > 0 {
		name = boards[len(boards)-1] // open the most recently created board
	}
	b, err := a.store.LoadBoard(name)
	if err != nil {
		b = board.NewBoard(name)
	}
	a.board = b
	a.canvas = ui.NewCanvas(b)
	a.canvas.OnChange(func() { a.store.Note(b) })
	a.canvas.OnQuit(func() { mygo.App.Quit() })

	mygo.NewWindow(mygo.WindowOptions{
		Title:       "IdeaBoard - " + b.Name,
		Width:       900,
		Height:      600,
		Transparent: true,
		Opacity:     0.92,
		Frameless:   true,
		CornerRadius: 8,
		Content:     mygo_ui.View(a.view),
	})
}

func (a *app) view(c *mygo_ui.Context) {
	a.canvas.View(c)
}

func (a *app) saveNow() {
	if a.store != nil && a.board != nil {
		if err := a.store.SaveNow(a.board); err != nil {
			log.Printf("ideaboard: save failed: %v", err)
		}
	}
}

func boardsDir() string {
	root, err := os.UserConfigDir()
	if err != nil {
		root = "."
	}
	return filepath.Join(root, "IdeaBoard", "boards")
}
