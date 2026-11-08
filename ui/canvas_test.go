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

func keyDownWithMod(mods ui.Modifiers, k ui.Key) ui.InputEvent {
	return ui.InputEvent{Kind: ui.InputKeyDown, Mods: mods, Key: k}
}

// newNote double-clicks empty space, types text and commits it, as a user
// would. It returns the note.
func newNote(t *testing.T, c *Canvas, b *board.Board, x, y float32, text string) *board.Item {
	t.Helper()
	before := len(b.Items())
	if !c.input(nil, pointerDown(0, x, y, 2)) {
		t.Fatal("double-click not taken")
	}
	if len(b.Items()) != before+1 {
		t.Fatalf("expected a new note, have %d items", len(b.Items()))
	}
	it := b.Items()[len(b.Items())-1]
	if c.editID != it.ID {
		t.Fatal("new note should open for typing")
	}
	c.editText = text
	c.commitEdit()
	return it
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

	it := newNote(t, c, b, 100, 100, "Buy milk")
	if it.Kind != board.KindNote || it.Text != "Buy milk" {
		t.Fatalf("bad note: %+v", it)
	}

	// Select the note with a click on its center.
	if !c.input(nil, pointerDown(0, it.X+it.W/2, it.Y+it.H/2, 1)) {
		t.Fatal("select note not taken")
	}
	if c.selected != it.ID {
		t.Fatalf("selected = %q, want %q", c.selected, it.ID)
	}
	c.input(nil, pointerUp(0, it.X+it.W/2, it.Y+it.H/2))

	if !c.input(nil, keyDown(ui.KeyDelete)) {
		t.Fatal("delete not taken")
	}
	if len(b.Items()) != 0 {
		t.Fatalf("expected 0 items after delete, got %d", len(b.Items()))
	}

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

func TestEmptyNoteIsRemovedOnCommit(t *testing.T) {
	b := board.NewBoard("T")
	c := NewCanvas(b)
	c.input(nil, pointerDown(0, 100, 100, 2))
	if len(b.Items()) != 1 {
		t.Fatalf("expected 1 note, got %d", len(b.Items()))
	}
	c.commitEdit() // nothing typed
	if len(b.Items()) != 0 {
		t.Fatalf("empty note should be removed, have %d", len(b.Items()))
	}
}

func TestEscapeNeverQuits(t *testing.T) {
	b := board.NewBoard("T")
	c := NewCanvas(b)
	quit := false
	c.OnQuit(func() { quit = true })
	c.input(nil, keyDown(ui.KeyEscape))
	if quit {
		t.Fatal("Escape must not quit the app")
	}
}

func TestRegionAndConnectorWithTools(t *testing.T) {
	b := board.NewBoard("T")
	c := NewCanvas(b)
	n1 := newNote(t, c, b, 100, 100, "one")
	n2 := newNote(t, c, b, 400, 100, "two")

	// Region: a plain click makes a default-size, labeled region.
	if !c.input(nil, keyDown(ui.KeyR)) {
		t.Fatal("region tool key not taken")
	}
	c.input(nil, pointerDown(0, 700, 300, 1))
	c.input(nil, pointerUp(0, 700, 300))
	reg := b.Items()[len(b.Items())-1]
	if reg.Kind != board.KindRegion || reg.Label != "Region" || reg.W != 320 {
		t.Fatalf("bad region: %+v", reg)
	}
	if c.tool != toolSelect {
		t.Fatal("tool should return to select after a region")
	}

	// Connector: click one note, then the other.
	c.input(nil, keyDown(ui.KeyC))
	c.input(nil, pointerDown(0, n1.X+n1.W/2, n1.Y+n1.H/2, 1))
	c.input(nil, pointerUp(0, n1.X+n1.W/2, n1.Y+n1.H/2))
	c.input(nil, pointerDown(0, n2.X+n2.W/2, n2.Y+n2.H/2, 1))
	c.input(nil, pointerUp(0, n2.X+n2.W/2, n2.Y+n2.H/2))
	con := b.Items()[len(b.Items())-1]
	if con.Kind != board.KindConnector || con.From != n1.ID || con.To != n2.ID {
		t.Fatalf("bad connector: %+v", con)
	}
	// Its ends sit on the notes' edges, on the line between their centers.
	x0, _, x1, _, ok := c.connectorEnds(con)
	if !ok {
		t.Fatal("connector has no ends")
	}
	if x0 < n1.X+n1.W-1 || x0 > n1.X+n1.W+connGap+1 || x1 > n2.X+1 || x1 < n2.X-connGap-1 {
		t.Errorf("connector ends not on note edges: x0=%v x1=%v", x0, x1)
	}
}

func TestResizeIsNotCumulative(t *testing.T) {
	b := board.NewBoard("T")
	c := NewCanvas(b)
	n := newNote(t, c, b, 300, 300, "x")
	c.setTool(toolSelect)
	startW := n.W

	// Press on the bottom-right handle and drag 30 DIP right. The same pointer
	// position reported twice must not grow the note twice.
	bx, by := n.X+n.W-2, n.Y+n.H-2
	c.input(nil, pointerDown(0, bx, by, 1))
	c.input(nil, pointerMove(bx+30, by))
	c.input(nil, pointerMove(bx+30, by))
	c.input(nil, pointerUp(0, bx+30, by))
	if n.W != startW+30 {
		t.Fatalf("W = %v, want %v", n.W, startW+30)
	}
}

func TestCanvasPenUndoRedo(t *testing.T) {
	b := board.NewBoard("T")
	c := NewCanvas(b)
	c.setTool(toolPen)

	// Draw a stroke: down, move, up.
	if !c.input(nil, pointerDown(0, 100, 100, 1)) {
		t.Fatal("pen down not taken")
	}
	if !c.penActive {
		t.Fatal("expected pen to be active")
	}
	c.input(nil, pointerMove(140, 140))
	c.input(nil, pointerUp(0, 140, 140))
	if c.penActive {
		t.Fatal("expected pen to end on up")
	}
	if strokeCount(b) != 1 {
		t.Fatalf("expected 1 stroke, got %d", strokeCount(b))
	}

	// Undo removes it, redo restores it.
	c.input(nil, keyDownWithMod(ui.Ctrl, ui.KeyZ))
	if strokeCount(b) != 0 {
		t.Fatalf("expected 0 strokes after undo, got %d", strokeCount(b))
	}
	if len(c.redo) != 1 {
		t.Fatalf("expected 1 redo entry, got %d", len(c.redo))
	}
	c.input(nil, keyDownWithMod(ui.Ctrl, ui.KeyY))
	if strokeCount(b) != 1 {
		t.Fatalf("expected 1 stroke after redo, got %d", strokeCount(b))
	}
	if len(c.redo) != 0 {
		t.Fatalf("expected empty redo stack, got %d", len(c.redo))
	}
}

func TestEraserRemovesStrokeAndUndoRestoresIt(t *testing.T) {
	b := board.NewBoard("T")
	c := NewCanvas(b)
	c.addStroke([][2]float32{{100, 100}, {150, 100}})
	c.setTool(toolEraser)
	c.input(nil, pointerDown(0, 120, 102, 1))
	if strokeCount(b) != 0 {
		t.Fatalf("eraser should remove the stroke, have %d", strokeCount(b))
	}
	c.input(nil, keyDownWithMod(ui.Ctrl, ui.KeyZ))
	if strokeCount(b) != 1 {
		t.Fatalf("undo should bring the erased stroke back, have %d", strokeCount(b))
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

func TestCanvasAddImage(t *testing.T) {
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
	it := &board.Item{X: 10, Y: 20, W: 100, H: 75, Fit: "contain"}
	c.AddImage("img_a.png", it, src)
	if len(b.Items()) != 1 {
		t.Fatalf("expected 1 image item, got %d", len(b.Items()))
	}
	if c.imageCache["img_a.png"] == nil {
		t.Fatal("expected image in cache")
	}
	if b.Items()[0].Asset != "img_a.png" || b.Items()[0].Kind != board.KindImage {
		t.Fatal("asset id or kind not stored")
	}

	// nil src is ignored.
	c.AddImage("img_b.png", &board.Item{Kind: board.KindImage, Asset: "img_b.png"}, nil)
	if len(b.Items()) != 1 {
		t.Fatalf("nil src should not add an item, got %d", len(b.Items()))
	}
}

func TestImageLoadsFromAssetOnce(t *testing.T) {
	b := board.NewBoard("T")
	c := NewCanvas(b)
	calls := 0
	c.SetAssetLoader(func(asset string) ui.ImageSource {
		calls++
		return nil // missing file
	})
	c.image("gone.png")
	c.image("gone.png")
	if calls != 1 {
		t.Fatalf("failed load retried: %d calls", calls)
	}
}

func TestNoteColorsUsePalette(t *testing.T) {
	for _, s := range noteStyles {
		if noteFill(s.fill) != mustHex(s.fill) {
			t.Errorf("palette color %s not used", s.fill)
		}
	}
	// Strokes and old files with other hex colors keep their own color.
	if got, ok := parseColor("#c8e6c9"); !ok || got != mustHex("#c8e6c9") {
		t.Error("valid hex color not parsed")
	}
	if noteFill("bogus") != mustHex(noteFillDefault) {
		t.Error("unknown color should fall back to the default fill")
	}
}
