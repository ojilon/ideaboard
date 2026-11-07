// IdeaBoard is a quiet, minimal note-board for capturing scattered ideas
// quickly while reading. Open the app, double-click, type, done.
//
// This file is the entry point: it creates the MyGo native-UI window, loads
// and saves boards through the storage package, and runs the app lifecycle.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	win := mygo.NewWindow(mygo.WindowOptions{
		Title:       "IdeaBoard - " + b.Name,
		Width:       900,
		Height:      600,
		Transparent: true,
		Opacity:     0.92,
		Frameless:   true,
		CornerRadius: 8,
		Content:     mygo_ui.View(a.view),
	})

	a.canvas.OnAttachImage(func() {
		a.attachImage(win, b)
	})
}

func (a *app) attachImage(win *mygo.Window, b *board.Board) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: win,
		Title:  "Attach image",
		Filters: []mygo.FileFilter{
			{Name: "Images", Extensions: []string{"png", "jpg", "jpeg", "webp", "gif"}, MIMETypes: []string{"image/*"}},
		},
	})
	if err != nil || len(paths) == 0 {
		return
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("ideaboard: read image %q: %v", path, err)
			continue
		}
		src, err := mygo_ui.DecodeImage(data)
		if err != nil {
			log.Printf("ideaboard: decode image %q: %v", path, err)
			continue
		}
		assetsDir, err := a.store.AssetsDir(b.Name)
		if err != nil {
			log.Printf("ideaboard: assets dir %q: %v", b.Name, err)
			continue
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			ext = ".png"
		}
		assetID := "img_" + strings.TrimSuffix(filepath.Base(path), ext) + ext
		// avoid collisions
		if _, err := os.Stat(filepath.Join(assetsDir, assetID)); err == nil {
			assetID = "img_" + strings.TrimSuffix(filepath.Base(path), ext) + "_" + fmt.Sprint(len(b.Items())) + ext
		}
		if err := os.WriteFile(filepath.Join(assetsDir, assetID), data, 0o644); err != nil {
			log.Printf("ideaboard: write image asset %q: %v", assetID, err)
			continue
		}
		// Place at board center, sized to a max of 320 DIPs wide.
		w, h := mygo_ui.ImageSize(src)
		if w <= 0 || h <= 0 {
			w, h = 240, 180
		}
		maxW := float32(320)
		scale := float32(1)
		if w > maxW {
			scale = maxW / w
		}
		dw, dh := float32(w)*scale, float32(h)*scale
		bx := a.canvas.PanX() + 100
		by := a.canvas.PanY() + 60
		a.canvas.AddImage(assetID, &board.Item{
			Kind:   board.KindImage,
			X:      bx, Y: by,
			W:      dw, H: dh,
			Color:  "", Asset: assetID, Fit: "contain",
		}, src)
	}
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
