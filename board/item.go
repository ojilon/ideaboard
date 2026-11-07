// Package board holds the in-memory model of a single idea board.
// A board has layers; each layer holds items (notes, regions, connectors,
// strokes, images). Items know their own kind, position, size, color and
// z-order. The board exposes small add/move/remove helpers and hit-testing.
package board

import (
	"math/rand"
	"time"
)

// Kind is the kind of an item.
type Kind string

const (
	KindNote      Kind = "note"
	KindRegion    Kind = "region"
	KindConnector Kind = "connector"
	KindStroke    Kind = "stroke"
	KindImage     Kind = "image"
)

// Item is one drawable thing on a board.
type Item struct {
	ID   string
	Kind Kind
	// X, Y top-left; W, H size, in board coordinates (DIPs).
	X, Y, W, H float32
	// Color is "#rrggbb" or "#rrggbbaa". Empty means the theme default.
	Color string
	// Z is draw order; larger draws on top.
	Z int

	// Kind-specific fields (zero when not applicable).
	Text     string       // note text
	FontSize float32      // note text size
	Label    string       // region label
	From, To string       // connector endpoint note ids
	Points   [][2]float32 // stroke points
	Width    float32      // stroke width
	Asset    string       // image asset file name
	Fit      string       // image fit: "contain" | "cover"
}

// newID makes a collision-resistant id.
func newID() string {
	b := make([]byte, 8)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(b)
	const chars = "0123456789abcdef"
	out := make([]byte, 16)
	for i := range b {
		out[i*2] = chars[b[i]>>4]
		out[i*2+1] = chars[b[i]&0xf]
	}
	return string(out)
}

// Layer is a z-ordered group of items on a board.
type Layer struct {
	Items []*Item
}

// View holds the pan/zoom state of the canvas viewport.
type View struct {
	PanX, PanY float32
	Zoom       float32
}

// Board is the root model of a board.
type Board struct {
	Name   string
	View   View
	Layers []*Layer
}

// NewBoard makes an empty board with one layer.
func NewBoard(name string) *Board {
	return &Board{
		Name:   name,
		View:   View{Zoom: 1},
		Layers: []*Layer{{}},
	}
}

// Items returns every item on every layer, in draw order (layer then index).
func (b *Board) Items() []*Item {
	var out []*Item
	for _, l := range b.Layers {
		out = append(out, l.Items...)
	}
	return out
}

// Item returns the item with the given id, or nil.
func (b *Board) Item(id string) *Item {
	for _, l := range b.Layers {
		for _, it := range l.Items {
			if it.ID == id {
				return it
			}
		}
	}
	return nil
}

// AddItem appends an item to the last layer with a fresh id and top z.
func (b *Board) AddItem(it *Item) {
	if it.ID == "" {
		it.ID = newID()
	}
	it.Z = b.topZ() + 1
	b.Restore(it)
}

// Restore appends an item to the last layer keeping its id and z. It is
// used to undo a removal.
func (b *Board) Restore(it *Item) {
	last := b.Layers[len(b.Layers)-1]
	last.Items = append(last.Items, it)
}

// RemoveItem removes the item with the given id.
func (b *Board) RemoveItem(id string) {
	for _, l := range b.Layers {
		kept := l.Items[:0]
		for _, it := range l.Items {
			if it.ID != id {
				kept = append(kept, it)
			}
		}
		l.Items = kept
	}
}

// topZ returns the largest z used, or -1.
func (b *Board) topZ() int {
	top := -1
	for _, it := range b.Items() {
		if it.Z > top {
			top = it.Z
		}
	}
	return top
}

// ToFront brings the item with the given id to the top z.
func (b *Board) ToFront(id string) {
	if it := b.Item(id); it != nil {
		it.Z = b.topZ() + 1
	}
}
