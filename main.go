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
	"ideaboard/settings"
	"ideaboard/storage"
	"ideaboard/ui"
)

type app struct {
	boardStore  *storage.Store
	board       *board.Board
	canvas      *ui.Canvas
	cfg         settings.Settings
	showSettings bool
	win         *mygo.Window
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
	a.boardStore = storage.Open(boardsDir())
	a.canvas = ui.NewCanvas(a.board)
	a.canvas.OnChange(func() { a.boardStore.Note(a.board) })
	a.canvas.OnQuit(func() { mygo.App.Quit() })
	a.canvas.OnAttachImage(func() { a.attachImage(a.win, a.board) })
	a.canvas.OnSettingsToggle(func() { a.showSettings = !a.showSettings })

	a.cfg = settings.Load()
	a.applySettings(a.cfg)

	boards, _ := a.boardStore.ListBoards()
	name := "Default"
	if len(boards) > 0 {
		name = boards[len(boards)-1]
	}
	b, err := a.boardStore.LoadBoard(name)
	if err != nil {
		b = board.NewBoard(name)
	}
	a.board = b
	// Rebind the canvas to the loaded board (it was created with a placeholder).
	a.canvas = ui.NewCanvas(b)
	a.canvas.OnChange(func() { a.boardStore.Note(b) })
	a.canvas.OnQuit(func() { mygo.App.Quit() })
	a.canvas.OnAttachImage(func() { a.attachImage(a.win, b) })
	a.canvas.OnSettingsToggle(func() { a.showSettings = !a.showSettings })
	if a.cfg.Background != "" {
		a.loadBackground(a.cfg.Background, a.cfg.BackgroundDim)
	}

	win := mygo.NewWindow(mygo.WindowOptions{
		Title:       "IdeaBoard - " + b.Name,
		Width:       a.cfg.Window.W, Height: a.cfg.Window.H,
		Transparent: true,
		Opacity:     0.92,
		Frameless:   true,
		CornerRadius: 8,
		Content:     mygo_ui.View(a.view),
	})
	a.win = win
}

func (a *app) applySettings(s settings.Settings) {
	mygo.Theme.SetSource(mygo.ThemeSource(s.Theme))
	if s.FontFile != "" {
		data, err := os.ReadFile(s.FontFile)
		if err == nil {
			if err := mygo_ui.RegisterFont(data, s.FontFamily); err != nil {
				log.Printf("ideaboard: register font %q: %v", s.FontFile, err)
			}
		} else {
			log.Printf("ideaboard: read font %q: %v", s.FontFile, err)
		}
	}
}

func (a *app) loadBackground(path string, dim float32) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ideaboard: read background %q: %v", path, err)
		return
	}
	src, err := mygo_ui.DecodeImage(data)
	if err != nil {
		log.Printf("ideaboard: decode background %q: %v", path, err)
		return
	}
	a.canvas.SetBackground(src, dim)
}

func (a *app) view(c *mygo_ui.Context) {
	a.canvas.View(c)
	if a.showSettings {
		settings.Panel(c, a.cfg, settings.Callbacks{
			OnTheme:      func(t settings.Theme) { a.cfg.Theme = t; a.applySettings(a.cfg); settings.Save(a.cfg) },
			OnFontFamily: func(f string) { a.cfg.FontFamily = f; settings.Save(a.cfg) },
			OnFontFile:   func(f string) { a.pickFontFile() },
			OnBackground: func(path string, dim float32) {
				a.cfg.Background = path
				a.cfg.BackgroundDim = dim
				a.loadBackground(path, dim)
				settings.Save(a.cfg)
			},
			OnClose: func() { a.showSettings = false },
		})
	}
}

func (a *app) pickFontFile() {
	if a.win == nil {
		return
	}
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: a.win,
		Title:  "Choose font",
		Filters: []mygo.FileFilter{
			{Name: "Fonts", Extensions: []string{"ttf", "otf"}, MIMETypes: []string{"font/*"}},
		},
	})
	if err != nil || len(paths) == 0 {
		return
	}
	a.cfg.FontFile = paths[0]
	a.cfg.FontFamily = filepath.Base(paths[0])
	a.applySettings(a.cfg)
	settings.Save(a.cfg)
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
		assetsDir, err := a.boardStore.AssetsDir(b.Name)
		if err != nil {
			log.Printf("ideaboard: assets dir %q: %v", b.Name, err)
			continue
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			ext = ".png"
		}
		assetID := "img_" + strings.TrimSuffix(filepath.Base(path), ext) + ext
		if _, err := os.Stat(filepath.Join(assetsDir, assetID)); err == nil {
			assetID = "img_" + strings.TrimSuffix(filepath.Base(path), ext) + "_" + fmt.Sprint(len(b.Items())) + ext
		}
		if err := os.WriteFile(filepath.Join(assetsDir, assetID), data, 0o644); err != nil {
			log.Printf("ideaboard: write image asset %q: %v", assetID, err)
			continue
		}
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

func (a *app) saveNow() {
	if a.boardStore != nil && a.board != nil {
		if err := a.boardStore.SaveNow(a.board); err != nil {
			log.Printf("ideaboard: board save failed: %v", err)
		}
	}
	settings.Save(a.cfg)
}

func boardsDir() string {
	root, err := os.UserConfigDir()
	if err != nil {
		root = "."
	}
	return filepath.Join(root, "IdeaBoard", "boards")
}
