package ui

import (
	"ideaboard/board"

	"github.com/egoist/mygo/ui"
)

const (
	noteMinW float32 = 180
	noteMinH float32 = 112
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
		// Shortcuts.
		if ev.Key == ui.KeyR {
			c.addRegion()
			return true
		}
		if ev.Key == ui.KeyC {
			c.addConnector()
			return true
		}
		if ev.Mods&ui.Ctrl != 0 && ev.Key == ui.KeyZ {
			c.undo()
			return true
		}
		if ev.Mods&ui.Ctrl != 0 && ev.Key == ui.KeyY {
			c.redo()
			return true
		}
		if ev.Mods&ui.Ctrl != 0 && ev.Key == ui.KeyComma {
			if c.onSettingsToggle != nil {
				c.onSettingsToggle()
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
			} else {
				c.startPen(bx, by)
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
		case c.penActive:
			bx, by := c.windowToBoard(ev.X, ev.Y)
			c.addPenPoint(bx, by)
			ctx.Invalidate()
			return true
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
			if c.penActive {
				c.endPen()
				return true
			}
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
		Color:    "#FFF4C7", // design system note-yellow
		Text:     "",
		FontSize: 16,
	})
	c.changed()
}

// addRegion places a labeled dotted region at the board center of the view.
func (c *Canvas) addRegion() {
	c.board.AddItem(&board.Item{
		Kind:   board.KindRegion,
		X:      c.panX + 100,
		Y:      c.panY + 60,
		W:      320,
		H:      200,
		Color:  "", // design-system region outline color
		Label:  "Mon",
		FontSize: 0,
	})
	c.changed()
}

// addConnector draws a dotted line between the first two notes on the board.
func (c *Canvas) addConnector() {
	notes := c.notes()
	if len(notes) < 2 {
		return
	}
	c.board.AddItem(&board.Item{
		Kind:   board.KindConnector,
		X:      0, Y: 0, W: 0, H: 0,
		Color:  "",
		From:   notes[0].ID,
		To:     notes[1].ID,
		FontSize: 0,
	})
	c.changed()
}

func (c *Canvas) notes() []*board.Item {
	var out []*board.Item
	for _, l := range c.board.Layers {
		for _, it := range l.Items {
			if it.Kind == board.KindNote {
				out = append(out, it)
			}
		}
	}
	return out
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

// startPen begins a freehand stroke at board point (bx, by).
func (c *Canvas) startPen(bx, by float32) {
	c.penActive = true
	c.penPoints = [][2]float32{{bx, by}}
}

// addPenPoint appends a point to the in-progress stroke.
func (c *Canvas) addPenPoint(bx, by float32) {
	last := c.penPoints[len(c.penPoints)-1]
	dx, dy := bx-last[0], by-last[1]
	// Only add a point when it moves far enough to keep strokes light.
	if dx*dx+dy*dy < 4 {
		return
	}
	c.penPoints = append(c.penPoints, [2]float32{bx, by})
}

// endPen finalizes the in-progress stroke as a board item.
func (c *Canvas) endPen() {
	if !c.penActive || len(c.penPoints) < 2 {
		c.penActive = false
		c.penPoints = nil
		return
	}
	// Bounds of the stroke.
	minX, minY := c.penPoints[0][0], c.penPoints[0][1]
	maxX, maxY := minX, minY
	for _, pt := range c.penPoints {
		if pt[0] < minX {
			minX = pt[0]
		}
		if pt[0] > maxX {
			maxX = pt[0]
		}
		if pt[1] < minY {
			minY = pt[1]
		}
		if pt[1] > maxY {
			maxY = pt[1]
		}
	}
	it := &board.Item{
		Kind:   board.KindStroke,
		X:      minX, Y: minY,
		W:      maxX - minX, H: maxY - minY,
		Color:  "#000000", // pen color
		Width:  2,
		Points: c.penPoints,
	}
	c.board.AddItem(it)
	c.penActive = false
	c.penPoints = nil
	c.changed()
}

// undo removes the most recently added stroke so it can be redone.
func (c *Canvas) undo() {
	var last *board.Item
	items := c.board.Items()
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].Kind == board.KindStroke {
			last = items[i]
			break
		}
	}
	if last == nil {
		return
	}
	c.board.RemoveItem(last.ID)
	c.undone = append(c.undone, last)
	c.changed()
}

// redo restores the most recently undone stroke.
func (c *Canvas) redo() {
	if len(c.undone) == 0 {
		return
	}
	it := c.undone[len(c.undone)-1]
	c.undone = c.undone[:len(c.undone)-1]
	c.board.AddItem(it)
	c.changed()
}
