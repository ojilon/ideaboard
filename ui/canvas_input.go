package ui

import (
	"ideaboard/board"

	"github.com/egoist/mygo/ui"
)

const (
	noteMinW float32 = 140
	noteMinH float32 = 100
)

// input turns pointer and key events into pan, zoom and note interaction.
func (c *Canvas) input(ctx *ui.Context, ev ui.InputEvent) bool {
	// While a note is being edited, the overlay TextArea takes input.
	if c.editID != "" {
		return false
	}

	switch ev.Kind {
	case ui.InputKeyDown:
		if ev.Key == ui.KeySpace {
			c.spaceHeld = true
			return true
		}
		if ev.Key == ui.KeyEscape {
			if c.editID != "" {
				c.commitEdit()
				return true
			}
			if c.colorPickID != "" {
				c.colorPickID = ""
				return true
			}
			if c.onQuit != nil {
				c.onQuit()
				return true
			}
		}
		if ev.Key == ui.KeyDelete || ev.Key == ui.KeyBackspace {
			if c.selected != "" {
				c.board.RemoveItem(c.selected)
				c.selected = ""
				c.changed()
				return true
			}
		}
	case ui.InputKeyUp:
		if ev.Key == ui.KeySpace {
			c.spaceHeld = false
			return true
		}
	case ui.InputPointerDown:
		bx, by := c.windowToBoard(ev.X, ev.Y)
		switch ev.Button {
		case 2:
			// middle button: pan
			c.startPan(ev.X, ev.Y)
			return true
		case 0:
			note := c.hitNote(bx, by)
			if ev.Clicks >= 2 {
				// double-click
				if note != nil {
					c.beginEdit(note)
				} else {
					c.createNote(bx, by)
				}
				c.colorPickID = ""
				return true
			}
			if note != nil {
				c.selected = note.ID
				c.colorPickID = ""
				if c.nearCorner(note, bx, by) {
					c.resizeID = note.ID
					c.resizeStartW, c.resizeStartH = note.W, note.H
					c.resizeStartBX, c.resizeStartBY = bx, by
				} else {
					c.dragID = note.ID
					c.dragOffX, c.dragOffY = bx-note.X, by-note.Y
				}
				return true
			}
			c.selected = ""
			c.colorPickID = ""
			if c.spaceHeld {
				c.startPan(ev.X, ev.Y)
			}
			return true
		case 1:
			// right-click: open the color picker on the note under the pointer.
			note := c.hitNote(bx, by)
			if note != nil {
				c.colorPickID = note.ID
				c.colorPickX, c.colorPickY = ev.X, ev.Y
				c.selected = note.ID
				return true
			}
			c.colorPickID = ""
		}
	case ui.InputPointerMove:
		switch {
		case c.resizeID != "":
			bx, by := c.windowToBoard(ev.X, ev.Y)
			it := c.boardItem(c.resizeID)
			nw := it.W + (bx - c.resizeStartBX)
			nh := it.H + (by - c.resizeStartBY)
			if nw < noteMinW {
				nw = noteMinW
			}
			if nh < noteMinH {
				nh = noteMinH
			}
			it.W, it.H = nw, nh
			ctx.Invalidate()
			c.changed()
			return true
		case c.dragID != "":
			bx, by := c.windowToBoard(ev.X, ev.Y)
			it := c.boardItem(c.dragID)
			c.board.MoveItem(c.dragID, bx-c.dragOffX-it.X, by-c.dragOffY-it.Y)
			ctx.Invalidate()
			c.changed()
			return true
		case c.draggingP:
			c.panTo(ev.X, ev.Y)
			ctx.Invalidate()
			return true
		}
	case ui.InputPointerUp:
		switch ev.Button {
		case 0:
			if c.resizeID != "" {
				c.resizeID = ""
				return true
			}
			if c.dragID != "" {
				c.dragID = ""
				return true
			}
			if c.draggingP {
				c.draggingP = false
				return true
			}
		case 2:
			if c.draggingP {
				c.draggingP = false
				return true
			}
		}
	case ui.InputScroll:
		if ev.Mods&ui.Ctrl != 0 {
			c.zoomAt(ev.X, ev.Y, ev.DY)
			ctx.Invalidate()
			return true
		}
	}
	return false
}

func (c *Canvas) windowToBoard(wx, wy float32) (float32, float32) {
	return wx/c.zoom + c.panX, wy/c.zoom + c.panY
}

func (c *Canvas) boardItem(id string) *board.Item {
	for _, l := range c.board.Layers {
		for _, it := range l.Items {
			if it.ID == id {
				return it
			}
		}
	}
	return nil
}

func (c *Canvas) hitNote(bx, by float32) *board.Item {
	// front-most note wins: scan back to front.
	items := c.board.Items()
	for i := len(items) - 1; i >= 0; i-- {
		it := items[i]
		if it.Kind != board.KindNote {
			continue
		}
		if bx >= it.X && bx <= it.X+it.W && by >= it.Y && by <= it.Y+it.H {
			return it
		}
	}
	return nil
}

func (c *Canvas) startPan(x, y float32) {
	c.draggingP = true
	c.panStartX, c.panStartY = x, y
}

func (c *Canvas) panTo(x, y float32) {
	dx := (c.panStartX - x) / c.zoom
	dy := (c.panStartY - y) / c.zoom
	c.panX += dx
	c.panY += dy
	c.panStartX, c.panStartY = x, y
	c.changed()
}

// zoomAt zooms by a factor derived from the wheel, keeping the point under
// the pointer fixed.
func (c *Canvas) zoomAt(wx, wy, dy float32) {
	const step = 0.0015
	factor := 1 + dy*step
	if factor < 0.3 {
		factor = 0.3
	}
	if factor > 3 {
		factor = 3
	}
	newZoom := c.zoom * factor
	if newZoom < 0.2 {
		newZoom = 0.2
	}
	if newZoom > 5 {
		newZoom = 5
	}
	if newZoom == c.zoom {
		return
	}
	invOld := 1 / c.zoom
	invNew := 1 / newZoom
	c.panX += wx * (invOld - invNew)
	c.panY += wy * (invOld - invNew)
	c.zoom = newZoom
	c.changed()
}

// createNote adds a soft-yellow note at board position (bx, by).
func (c *Canvas) createNote(bx, by float32) {
	w, h := noteMinW, noteMinH
	c.board.AddItem(&board.Item{
		Kind:     board.KindNote,
		X:        bx - w/2,
		Y:        by - h/2,
		W:        w,
		H:        h,
		Color:    "#fff59d",
		Text:     "",
		FontSize: 16,
	})
	c.changed()
}

// beginEdit starts in-place editing of a note.
func (c *Canvas) beginEdit(it *board.Item) {
	c.editID = it.ID
	c.editText = it.Text
}

// commitEdit applies the edited text to its note and leaves edit mode.
func (c *Canvas) commitEdit() {
	if c.editID == "" {
		return
	}
	if it := c.boardNote(c.editID); it != nil {
		it.Text = c.editText
		c.changed()
	}
	c.editID = ""
	c.editText = ""
}
