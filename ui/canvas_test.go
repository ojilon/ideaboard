package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"ideaboard/board"
	"ideaboard/storage"

	"github.com/egoist/mygo/ui"
)

func pointerDown(button, x, y float32, clicks int) ui.InputEvent {
	return ui.InputEvent{Kind: ui.InputPointerDown, Button: int(button), X: x, Y: y, Clicks: clicks}
}

func pointerMove(x, y float32) ui.InputEvent {
	return ui.InputEvent{Kind: ui.InputPointerMove, Button: 0, X: x, Y: y}
}

func pointerUp(button float32, x, y float32) ui.InputEvent {
	return ui.InputEvent{Kind: ui.InputPointerUp, Button: int(button), X: x, Y: y}
}

func keyDown(k ui.Key) ui.InputEvent {
	return ui.InputEvent{Kind: ui.InputKeyDown, Key: k}
}

func TestCanvasCreateNoteTriggersAutosave(t *testing.T) {
	dir := t.TempDir()
	store := storage.Open(dir)
	b := board.NewBoard("T")
	c := NewCanvas(b)
	saved := false
	c.OnChange(func() { store.Note(b) })
	store.OnSave(func(err error) {
		if err == nil {
			saved = true
		}
	})

	// double-click empty space at (100, 100) -> create note
	if !c.input(nil, pointerDown(0, 100, 100, 2)) {
		t.Fatal("create note not taken")
	}
	if len(b.Items()) != 1 {
		t.Fatalf("expected 1 note, got %d", len(b.Items()))
	}
	it := b.Items()[0]
	if it.Kind != board.KindNote {
		t.Fatalf("expected note, got %v", it.Kind)
	}

	// single-click the note to select it
	if !c.input(nil, pointerDown(0, it.X+it.W/2, it.Y+it.H/2, 1)) {
		t.Fatal("select note not taken")
	}
	if c.selected != it.ID {
		t.Fatalf("selected = %q, want %q", c.selected, it.ID)
	}

	// delete it
	if !c.input(nil, keyDown(ui.KeyDelete)) {
		t.Fatal("delete not taken")
	}
	if len(b.Items()) != 0 {
		t.Fatalf("expected 0 notes after delete, got %d", len(b.Items()))
	}

	// flush debounce and confirm store saved + persisted empty board
	store.SaveNow(b)
	if !saved {
		t.Fatal("expected store to have saved")
	}
	loaded, err := store.LoadBoard("T")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Items()) != 0 {
		t.Fatalf("expected empty board on disk, got %d", len(loaded.Items()))
	}
}

func TestCanvasRegionAndConnector(t *testing.T) {
	b := board.NewBoard("T")
	c := NewCanvas(b)

	// two notes
	if !c.input(nil, pointerDown(0, 100, 100, 2)) {
		t.Fatal("create note 1")
	}
	if !c.input(nil, pointerDown(0, 400, 100, 2)) {
		t.Fatal("create note 2")
	}
	if len(b.Items()) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(b.Items()))
	}
	n1, n2 := b.Items()[0], b.Items()[1]

	// region via 'r'
	if !c.input(nil, keyDown(ui.KeyR)) {
		t.Fatal("add region not taken")
	}
	if len(b.Items()) != 3 {
		t.Fatalf("expected 3 items after region, got %d", len(b.Items()))
	}
	reg := b.Items()[2]
	if reg.Kind != board.KindRegion || reg.Label != "Mon" {
		t.Fatalf("bad region: %+v", reg)
	}

	// connector via 'c' between the two notes
	if !c.input(nil, keyDown(ui.KeyC)) {
		t.Fatal("add connector not taken")
	}
	if len(b.Items()) != 4 {
		t.Fatalf("expected 4 items after connector, got %d", len(b.Items()))
	}
	con := b.Items()[3]
	if con.Kind != board.KindConnector || con.From != n1.ID || con.To != n2.ID {
		t.Fatalf("bad connector: %+v", con)
	}
}

func strokeCount(b *board.Board) int {
	n := 0
	for _, it := range b.Items() {
		if it.Kind == board.KindStroke {
			n++
		}
	}
	return n
}

func TestCanvasPenUndoRedo(t *testing.T) {
	b := board.NewBoard("T")
	c := NewCanvas(b)

	// Draw a stroke: down, move, up.
	if !c.input(nil, pointerDown(0, 100, 100, 1)) {
		t.Fatal("pen down not taken")
	}
	if !c.penActive {
		t.Fatal("expected pen to be active")
	}
	if !c.input(nil, pointerMove(140, 140)) {
		t.Fatal("pen move not taken")
	}
	if !c.input(nil, pointerUp(0, 140, 140)) {
		t.Fatal("pen up not taken")
	}
	if c.penActive {
		t.Fatal("expected pen to end on up")
	}
	if strokeCount(b) != 1 {
		t.Fatalf("expected 1 stroke, got %d", strokeCount(b))
	}

	// Undo removes it.
	if !c.input(nil, keyDownWithMod(ui.Ctrl, ui.KeyZ)) {
		t.Fatal("undo not taken")
	}
	if strokeCount(b) != 0 {
		t.Fatalf("expected 0 strokes after undo, got %d", strokeCount(b))
	}
	if len(c.undone) != 1 {
		t.Fatalf("expected 1 undone stroke, got %d", len(c.undone))
	}

	// Redo restores it.
	if !c.input(nil, keyDownWithMod(ui.Ctrl, ui.KeyY)) {
		t.Fatal("redo not taken")
	}
	if strokeCount(b) != 1 {
		t.Fatalf("expected 1 stroke after redo, got %d", strokeCount(b))
	}
	if len(c.undone) != 0 {
		t.Fatalf("expected 0 undone after redo, got %d", len(c.undone))
	}
}

func keyDownWithMod(mods ui.Modifiers, k ui.Key) ui.InputEvent {
	return ui.InputEvent{Kind: ui.InputKeyDown, Mods: mods, Key: k}
}

func TestCanvasAddImage(t *testing.T) {
	// Build a tiny 2x2 PNG in memory and decode it via MyGo.
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetRGBA(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode: %v", err)
	}
	src, err := ui.DecodeImage(buf.Bytes())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if w, h := ui.ImageSize(src); w != 2 || h != 2 {
		t.Fatalf("image size = %vx%v, want 2x2", w, h)
	}

	b := board.NewBoard("T")
	c := NewCanvas(b)
	it := &board.Item{Kind: board.KindImage, X: 10, Y: 20, W: 100, H: 75, Asset: "img_a.png", Fit: "contain"}
	c.AddImage("img_a.png", it, src)
	if len(b.Items()) != 1 {
		t.Fatalf("expected 1 image item, got %d", len(b.Items()))
	}
	if c.imageCache["img_a.png"] == nil {
		t.Fatal("expected image in cache")
	}
	if b.Items()[0].Asset != "img_a.png" {
		t.Fatal("asset id not stored")
	}

	// nil src is ignored.
	c.AddImage("img_b.png", &board.Item{Kind: board.KindImage, Asset: "img_b.png"}, nil)
	if len(b.Items()) != 1 {
		t.Fatalf("nil src should not add an item, got %d", len(b.Items()))
	}
}
