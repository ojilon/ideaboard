package ui

import (
	"testing"

	"ideaboard/board"
	"ideaboard/storage"

	"github.com/egoist/mygo/ui"
)

func pointerDown(button, x, y float32, clicks int) ui.InputEvent {
	return ui.InputEvent{Kind: ui.InputPointerDown, Button: int(button), X: x, Y: y, Clicks: clicks}
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
