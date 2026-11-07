package storage

import (
	"os"
	"path/filepath"
	"testing"

	"ideaboard/board"
)

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := Open(dir)

	b := board.NewBoard("Test")
	b.AddItem(&board.Item{Kind: board.KindNote, X: 10, Y: 20, W: 100, H: 80, Color: "#fff59d", Text: "hello", FontSize: 14})
	b.AddItem(&board.Item{Kind: board.KindNote, X: 130, Y: 20, W: 100, H: 80, Color: "#bbdefb", Text: "world"})

	if err := s.SaveBoard(b); err != nil {
		t.Fatalf("save: %v", err)
	}

	// board.json and assets/ must exist.
	if _, err := os.Stat(filepath.Join(dir, "Test", "board.json")); err != nil {
		t.Errorf("board.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "Test", "assets")); err != nil {
		t.Errorf("assets/ missing: %v", err)
	}

	loaded, err := s.LoadBoard("Test")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Name != "Test" {
		t.Errorf("name = %q, want %q", loaded.Name, "Test")
	}
	if len(loaded.Items()) != 2 {
		t.Errorf("items = %d, want 2", len(loaded.Items()))
	}
	it := loaded.Items()[0]
	if it.Text != "hello" || it.Color != "#fff59d" || it.X != 10 || it.Y != 20 {
		t.Errorf("note mismatch: %+v", it)
	}
}

func TestDecodeIgnoresUnknownKind(t *testing.T) {
	dir := t.TempDir()
	s := Open(dir)

	b := board.NewBoard("X")
	b.AddItem(&board.Item{Kind: board.KindNote, X: 1, Y: 1, W: 2, H: 2, Text: "n"})
	if err := s.SaveBoard(b); err != nil {
		t.Fatal(err)
	}
	loaded, err := s.LoadBoard("X")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Items()[0].Text != "n" {
		t.Errorf("text not preserved")
	}
}

func TestListBoards(t *testing.T) {
	dir := t.TempDir()
	s := Open(dir)
	if _, err := s.ListBoards(); err != nil {
		t.Fatalf("list empty dir: %v", err)
	}
	b := board.NewBoard("A")
	if err := s.SaveBoard(b); err != nil {
		t.Fatal(err)
	}
	names, err := s.ListBoards()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(names) != 1 || names[0] != "A" {
		t.Errorf("boards = %v, want [A]", names)
	}
}
