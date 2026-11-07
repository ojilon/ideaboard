// IdeaBoard is a quiet, minimal note-board for capturing scattered ideas
// quickly while reading. Open the app, double-click, type, done.
//
// This file is the entry point. It owns the app shell: the board list, the
// open board, the settings, and the MyGo window and its lifecycle.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	mygo_ui "github.com/egoist/mygo/ui"

	"ideaboard/board"
	"ideaboard/settings"
	"ideaboard/storage"
	"ideaboard/ui"
)

type app struct {
	store        *storage.Store
	cfg          settings.Settings
	bg           mygo_ui.ImageSource // decoded background, shared by boards
	board        *board.Board        // open board; nil on the board list
	canvas       *ui.Canvas
	list         *ui.BoardList
	showSettings bool
	win          *mygo.Window
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
	a.cfg = settings.Load()
	a.applySettings(a.cfg)
	if a.cfg.Background != "" {
		a.loadBackground(a.cfg.Background)
	}
	a.list = &ui.BoardList{
		OnOpen:     a.openBoard,
		OnNew:      a.newBoard,
		OnSettings: a.toggleSettings,
	}

	// Open the last board when it still exists; otherwise show the list.
	if a.cfg.LastBoard != "" && a.store.Exists(a.cfg.LastBoard) {
		a.openBoard(a.cfg.LastBoard)
	}

	win := mygo.NewWindow(mygo.WindowOptions{
		Title:        "IdeaBoard",
		Width:        a.cfg.Window.W,
		Height:       a.cfg.Window.H,
		Transparent:  true,
		Opacity:      float64(a.cfg.WindowOpacity),
		Frameless:    true,
		CornerRadius: 8,
		Content:      mygo_ui.View(a.view),
	})
	a.win = win
}

// view builds one frame: the open board, or the board list.
func (a *app) view(c *mygo_ui.Context) {
	if a.canvas == nil {
		a.list.View(c)
		return
	}
	a.canvas.View(c)
	if a.showSettings {
		settings.Panel(c, a.cfg, settings.Callbacks{
			OnTheme:          func(t settings.Theme) { a.cfg.Theme = t; a.applySettings(a.cfg); a.save() },
			OnFontFamily:     func(f string) { a.cfg.FontFamily = f; a.canvas.SetNoteFont(f); a.save() },
			OnFontFile:       a.pickFontFile,
			OnPickBackground: a.pickBackground,
			OnBackground: func(path string, dim float32) {
				a.cfg.BackgroundDim = dim
				if path != a.cfg.Background {
					a.cfg.Background = path
					a.bg = nil
					if path != "" {
						a.loadBackground(path)
					}
				}
				a.applyBackground()
				a.save()
			},
			OnOpacity: func(o float32) {
				a.cfg.WindowOpacity = o
				if a.win != nil {
					a.win.SetOpacity(float64(o))
				}
				a.save()
			},
			OnClose: a.toggleSettings,
		})
	}
}

// toggleSettings opens or closes the settings panel.
func (a *app) toggleSettings() {
	a.showSettings = !a.showSettings
	if a.canvas != nil {
		a.canvas.SetShowSettings(a.showSettings)
	}
}

// openBoard saves the open board, then loads and shows the named board.
func (a *app) openBoard(name string) {
	a.saveNow()
	b, err := a.store.LoadBoard(name)
	if err != nil {
		log.Printf("ideaboard: open board %q: %v", name, err)
		return
	}
	a.board = b
	c := ui.NewCanvas(b)
	c.OnChange(func() { a.store.Note(b) })
	c.OnAttachImage(a.attachImage)
	c.OnPaste(a.paste)
	c.OnSettingsToggle(a.toggleSettings)
	c.OnBoards(a.closeBoard)
	c.OnQuit(func() { mygo.App.Quit() })
	c.SetNoteFont(a.cfg.FontFamily)
	c.SetAssetLoader(a.assetLoader(name))
	c.SetBackground(a.bg, a.cfg.BackgroundDim)
	c.SetShowSettings(a.showSettings)
	a.canvas = c
	a.cfg.LastBoard = name
	a.save()
}

// closeBoard saves the open board and returns to the board list.
func (a *app) closeBoard() {
	a.saveNow()
	a.board = nil
	a.canvas = nil
	a.refreshList()
}

// newBoard creates an empty board with a free name and opens it.
func (a *app) newBoard() {
	name := "Untitled board"
	for i := 2; a.store.Exists(name); i++ {
		name = fmt.Sprintf("Untitled board %d", i)
	}
	if err := a.store.SaveBoard(board.NewBoard(name)); err != nil {
		log.Printf("ideaboard: create board %q: %v", name, err)
		return
	}
	a.openBoard(name)
}

// refreshList reloads the board list from disk.
func (a *app) refreshList() {
	infos, err := a.store.ListBoardInfo()
	if err != nil {
		log.Printf("ideaboard: list boards: %v", err)
	}
	a.list.Infos = infos
}

// applySettings applies the theme and the note font to the running app.
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

// loadBackground decodes the background image. Only PNG and JPEG decode.
func (a *app) loadBackground(path string) {
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
	a.bg = src
}

// applyBackground hands the current background to the open board.
func (a *app) applyBackground() {
	if a.canvas != nil {
		a.canvas.SetBackground(a.bg, a.cfg.BackgroundDim)
	}
}

// pickBackground asks for a background image and uses it.
func (a *app) pickBackground() {
	path := a.pickFile("Choose background", "Images", []string{"png", "jpg", "jpeg"}, "image/*")
	if path == "" {
		return
	}
	a.cfg.Background = path
	a.bg = nil
	a.loadBackground(path)
	a.applyBackground()
	a.save()
}

// pickFontFile asks for a font file and uses it for note text.
func (a *app) pickFontFile() {
	path := a.pickFile("Choose font", "Fonts", []string{"ttf", "otf"}, "font/*")
	if path == "" {
		return
	}
	a.cfg.FontFile = path
	a.cfg.FontFamily = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	a.applySettings(a.cfg)
	if a.canvas != nil {
		a.canvas.SetNoteFont(a.cfg.FontFamily)
	}
	a.save()
}

// pickFile shows an open dialog and returns the first path, or "".
func (a *app) pickFile(title, filterName string, exts []string, mime string) string {
	if a.win == nil {
		return ""
	}
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent:  a.win,
		Title:   title,
		Filters: []mygo.FileFilter{{Name: filterName, Extensions: exts, MIMETypes: []string{mime}}},
	})
	if err != nil || len(paths) == 0 {
		return ""
	}
	return paths[0]
}

// attachImage asks for image files and adds each one to the open board.
func (a *app) attachImage() {
	if a.canvas == nil || a.win == nil {
		return
	}
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: a.win,
		Title:  "Attach image",
		Filters: []mygo.FileFilter{
			{Name: "Images", Extensions: []string{"png", "jpg", "jpeg"}, MIMETypes: []string{"image/*"}},
		},
	})
	if err != nil {
		return
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			log.Printf("ideaboard: read image %q: %v", p, err)
			continue
		}
		a.addImage(data, filepath.Ext(p))
	}
}

// paste adds an image from the clipboard: bitmap data first, then copied files.
func (a *app) paste() {
	if a.canvas == nil {
		return
	}
	if data, err := mygo.Clipboard.ReadFormat(transfer.PNG); err == nil && len(data) > 0 {
		a.addImage(data, ".png")
		return
	}
	files, err := mygo.Clipboard.ReadFiles()
	if err != nil {
		return
	}
	for _, p := range files {
		ext := strings.ToLower(filepath.Ext(p))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			continue
		}
		if data, err := os.ReadFile(p); err == nil {
			a.addImage(data, ext)
		}
	}
}

// addImage stores image bytes as a new asset of the open board and places
// the image in the middle of the view, scaled down to at most 320 wide.
func (a *app) addImage(data []byte, ext string) {
	if a.canvas == nil || a.board == nil {
		return
	}
	src, err := mygo_ui.DecodeImage(data)
	if err != nil {
		log.Printf("ideaboard: decode image: %v", err)
		return
	}
	dir, err := a.store.AssetsDir(a.board.Name)
	if err != nil {
		log.Printf("ideaboard: assets dir: %v", err)
		return
	}
	ext = strings.ToLower(ext)
	if ext == "" {
		ext = ".png"
	}
	assetID := fmt.Sprintf("img_%d%s", time.Now().UnixNano(), ext)
	if err := os.WriteFile(filepath.Join(dir, assetID), data, 0o644); err != nil {
		log.Printf("ideaboard: write image asset %q: %v", assetID, err)
		return
	}
	w, h := mygo_ui.ImageSize(src)
	if w <= 0 || h <= 0 {
		w, h = 240, 180
	}
	if w > 320 {
		h *= 320 / w
		w = 320
	}
	cx, cy := a.canvas.ViewCenter()
	a.canvas.AddImage(assetID, &board.Item{
		X: cx - w/2, Y: cy - h/2, W: w, H: h,
		Fit: "contain",
	}, src)
}

// assetLoader reads an image asset of a board from disk.
func (a *app) assetLoader(name string) func(asset string) mygo_ui.ImageSource {
	return func(asset string) mygo_ui.ImageSource {
		dir, err := a.store.AssetsDir(name)
		if err != nil {
			return nil
		}
		data, err := os.ReadFile(filepath.Join(dir, filepath.Base(asset)))
		if err != nil {
			log.Printf("ideaboard: read asset %q: %v", asset, err)
			return nil
		}
		src, err := mygo_ui.DecodeImage(data)
		if err != nil {
			log.Printf("ideaboard: decode asset %q: %v", asset, err)
			return nil
		}
		return src
	}
}

// save writes the settings file.
func (a *app) save() {
	if err := settings.Save(a.cfg); err != nil {
		log.Printf("ideaboard: save settings: %v", err)
	}
}

// saveNow flushes the open board and the settings, including window size.
func (a *app) saveNow() {
	if a.win != nil {
		w, h := a.win.Size()
		x, y := a.win.Position()
		if w > 0 && h > 0 {
			a.cfg.Window.X, a.cfg.Window.Y = x, y
			a.cfg.Window.W, a.cfg.Window.H = w, h
		}
	}
	if a.store != nil && a.board != nil {
		if err := a.store.SaveNow(a.board); err != nil {
			log.Printf("ideaboard: board save failed: %v", err)
		}
	}
	a.save()
}

func boardsDir() string {
	root, err := os.UserConfigDir()
	if err != nil {
		root = "."
	}
	return filepath.Join(root, "IdeaBoard", "boards")
}
