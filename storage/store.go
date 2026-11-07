// Package storage persists a board to disk: one folder per board
// (<boards>/<name>/board.json + <boards>/<name>/assets/).
//
// Writes are safe: serialize to board.json.tmp in the same folder, then
// rename over board.json, so a crash never leaves a half-written board.json.
// Changes are autosaved debounced, plus a final save on close/quit.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ideaboard/board"
)

// FormatVersion is the board.json schema version.
const FormatVersion = 1

// DefaultSaveDebounce is how long to wait after the last change before
// autosaving.
const DefaultSaveDebounce = time.Second

// fileItem is the on-disk form of a board.Item.
type fileItem struct {
	ID     string   `json:"id"`
	Kind   string   `json:"kind"`
	X      float32  `json:"x"`
	Y      float32  `json:"y"`
	W      float32  `json:"w"`
	H      float32  `json:"h"`
	Color  string   `json:"color"`
	Z      int      `json:"z"`
	Text   string   `json:"text,omitempty"`
	FontSize float32 `json:"fontSize,omitempty"`
	Label  string   `json:"label,omitempty"`
	From   string   `json:"from,omitempty"`
	To     string   `json:"to,omitempty"`
	Points [][2]float32 `json:"points,omitempty"`
	Width  float32 `json:"width,omitempty"`
	Asset  string   `json:"asset,omitempty"`
	Fit    string   `json:"fit,omitempty"`
}

type fileView struct {
	PanX float32 `json:"panX"`
	PanY float32 `json:"panY"`
	Zoom float32 `json:"zoom"`
}

type boardFile struct {
	Version int        `json:"version"`
	Name    string     `json:"name"`
	View    fileView   `json:"view"`
	Items   []fileItem `json:"items"`
}

// EncodeBoard converts a board into its on-disk file form.
func EncodeBoard(b *board.Board) boardFile {
	f := boardFile{Version: FormatVersion, Name: b.Name}
	f.View = fileView{PanX: b.View.PanX, PanY: b.View.PanY, Zoom: b.View.Zoom}
	items := b.Items()
	f.Items = make([]fileItem, 0, len(items))
	for _, it := range items {
		fi := fileItem{
			ID:      it.ID,
			Kind:    string(it.Kind),
			X:       it.X, Y: it.Y, W: it.W, H: it.H,
			Color:   it.Color, Z: it.Z,
			Text:    it.Text, FontSize: it.FontSize,
			Label:   it.Label, From: it.From, To: it.To,
			Width:   it.Width, Asset: it.Asset, Fit: it.Fit,
		}
		if len(it.Points) > 0 {
			fi.Points = make([][2]float32, len(it.Points))
			copy(fi.Points, it.Points)
		}
		f.Items = append(f.Items, fi)
	}
	return f
}

// DecodeBoard converts a file form back into a board. Unknown kinds are kept
// as-is so the board round-trips even if this build doesn't render them.
func DecodeBoard(f boardFile) *board.Board {
	b := board.NewBoard(f.Name)
	b.View.PanX, b.View.PanY, b.View.Zoom = f.View.PanX, f.View.PanY, f.View.Zoom
	if b.View.Zoom == 0 {
		b.View.Zoom = 1
	}
	for _, fi := range f.Items {
		it := &board.Item{
			ID:      fi.ID,
			Kind:    board.Kind(fi.Kind),
			X:       fi.X, Y: fi.Y, W: fi.W, H: fi.H,
			Color:   fi.Color, Z: fi.Z,
			Text:    fi.Text, FontSize: fi.FontSize,
			Label:   fi.Label, From: fi.From, To: fi.To,
			Width:   fi.Width, Asset: fi.Asset, Fit: fi.Fit,
		}
		if len(fi.Points) > 0 {
			it.Points = make([][2]float32, len(fi.Points))
			copy(it.Points, fi.Points)
		}
		if it.ID == "" {
			it.ID = fi.Kind + "-" + fmt.Sprint(len(b.Layers[0].Items))
		}
		b.Layers[0].Items = append(b.Layers[0].Items, it)
	}
	return b
}

// Store persists boards to a boards directory.
type Store struct {
	dir     string
	debounce time.Duration
	timer    *time.Timer
	pending  *board.Board
	dirty    bool
	onSave   func(err error)
}

// Open opens (creating if needed) the store at dir.
func Open(dir string) *Store {
	return &Store{dir: dir, debounce: DefaultSaveDebounce}
}

// Debounce sets the autosave debounce.
func (s *Store) Debounce(d time.Duration) { s.debounce = d }

// OnSave sets a callback invoked after each save attempt.
func (s *Store) OnSave(fn func(err error)) { s.onSave = fn }

// dirFor returns the folder for a board.
func (s *Store) dirFor(name string) string { return filepath.Join(s.dir, name) }

// ListBoards returns the names of all boards (subdirectories containing a
// board.json), sorted.
func (s *Store) ListBoards() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(s.dir, e.Name(), "board.json")); err == nil {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// LoadBoard reads board.json for the named board.
func (s *Store) LoadBoard(name string) (*board.Board, error) {
	data, err := os.ReadFile(filepath.Join(s.dirFor(name), "board.json"))
	if err != nil {
		return nil, err
	}
	var f boardFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse board.json: %w", err)
	}
	return DecodeBoard(f), nil
}

// SaveBoard writes the board safely (temp file + rename).
func (s *Store) SaveBoard(b *board.Board) error {
	dir := s.dirFor(b.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		return err
	}
	f := EncodeBoard(b)
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, "board.json.tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, filepath.Join(dir, "board.json")); err != nil {
		return err
	}
	s.dirty = false
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	s.pending = nil
	if s.onSave != nil {
		s.onSave(nil)
	}
	return nil
}

// Note marks the board dirty and schedules a debounced autosave.
func (s *Store) Note(b *board.Board) {
	s.dirty = true
	s.pending = b
	if s.timer != nil {
		s.timer.Reset(s.debounce)
		return
	}
	s.timer = time.AfterFunc(s.debounce, func() {
		s.timer = nil
		b := s.pending
		s.pending = nil
		if b == nil {
			return
		}
		if err := s.SaveBoard(b); err != nil && s.onSave != nil {
			s.onSave(err)
		}
	})
}

// SaveNow flushes any pending autosave immediately.
func (s *Store) SaveNow(b *board.Board) error {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	s.pending = nil
	return s.SaveBoard(b)
}

