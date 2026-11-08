package ui

import (
	"ideaboard/board"

	"github.com/egoist/mygo/ui"
)

// input turns pointer and key events into tool actions, pan and zoom.
func (c *Canvas) input(ctx *ui.Context, ev ui.InputEvent) bool {
	// While a note is being edited, the overlay text area has the keyboard.
	// A press outside it finishes the edit, as clicking away does.
	if c.editID != "" {
		if ev.Kind != ui.InputPointerDown {
			return false
		}
		c.commitEdit()
	}
	switch ev.Kind {
	case ui.InputKeyDown:
		return c.keyDown(ev)
	case ui.InputKeyUp:
		if ev.Key == ui.KeySpace {
			c.spaceHeld = false
			return true
		}
	case ui.InputPointerDown:
		return c.pointerDown(ev)
	case ui.InputPointerMove:
		return c.pointerMove(ctx, ev)
	case ui.InputPointerUp:
		return c.pointerUp(ev)
	case ui.InputScroll:
		if ev.Mods&ui.Ctrl != 0 {
			c.zoomAt(ev.X, ev.Y, ev.DY)
			ctx.Invalidate()
			return true
		}
	}
	return false
}

// keyDown handles shortcuts. Plain letters pick tools (the design system's
// keyboard equivalents); Ctrl combinations are edits and app commands.
func (c *Canvas) keyDown(ev ui.InputEvent) bool {
	ctrl := ev.Mods&ui.Ctrl != 0
	if ctrl {
		switch ev.Key {
		case ui.KeyZ:
			if ev.Mods&ui.Shift != 0 {
				c.redoLast()
			} else {
				c.undo()
			}
		case ui.KeyY:
			c.redoLast()
		case ui.KeyV:
			if c.onPaste != nil {
				c.onPaste()
			}
		case ui.KeyComma:
			if c.onSettingsToggle != nil {
				c.onSettingsToggle()
			}
		case ui.KeyQ:
			if c.onQuit != nil {
				c.onQuit()
			}
		default:
			return false
		}
		return true
	}
	switch ev.Key {
	case ui.KeySpace:
		c.spaceHeld = true
	case ui.KeyEscape:
		c.escape()
	case ui.KeyDelete, ui.KeyBackspace:
		return c.deleteSelected()
	case ui.KeyV:
		c.setTool(toolSelect)
	case ui.KeyN:
		x, y := c.ViewCenter()
		c.createNoteAt(x, y)
		c.setTool(toolSelect)
	case ui.KeyR:
		c.setTool(toolRegion)
	case ui.KeyC:
		c.setTool(toolConnector)
	case ui.KeyP:
		c.setTool(toolPen)
	case ui.KeyE:
		c.setTool(toolEraser)
	case ui.KeyI:
		c.setTool(toolImage)
		c.pickImage()
	default:
		return false
	}
	return true
}

// escape cancels the current gesture or tool, then clears the selection.
// It never quits the app: the window is frameless and has no other close
// shortcut, so Ctrl+Q is the keyboard quit.
func (c *Canvas) escape() {
	switch {
	case c.colorPickID != "":
		c.colorPickID = ""
	case c.tool != toolSelect:
		c.setTool(toolSelect)
	case c.selected != "":
		c.selected = ""
	}
}

// pickImage opens the image picker and returns to the select tool.
func (c *Canvas) pickImage() {
	c.setTool(toolSelect)
	if c.onAttachImage != nil {
		c.onAttachImage()
	}
}

func (c *Canvas) pointerDown(ev ui.InputEvent) bool {
	bx, by := c.windowToBoard(ev.X, ev.Y)
	switch ev.Button {
	case 2: // middle button pans
		c.startPan(ev.X, ev.Y)
		return true
	case 1: // right button opens the color picker on a note
		c.colorPickID = ""
		if it := c.hitItem(bx, by); itemKindNote(it) {
			c.colorPickID = it.ID
			c.colorPickX, c.colorPickY = ev.X, ev.Y
			c.selected = it.ID
			return true
		}
		return false
	case 0:
	default:
		return false
	}

	c.colorPickID = ""
	if c.spaceHeld {
		c.startPan(ev.X, ev.Y)
		return true
	}
	hit := c.hitItem(bx, by)
	if ev.Clicks >= 2 {
		c.doubleClick(hit, bx, by)
		return true
	}
	switch c.tool {
	case toolSelect:
		return c.selectDown(hit, bx, by, ev)
	case toolNote:
		c.createNoteAt(bx, by)
		c.setTool(toolSelect)
	case toolRegion:
		c.draftX, c.draftY = bx, by
		c.draft = &board.Item{Kind: board.KindRegion, X: bx, Y: by}
	case toolConnector:
		c.connectorDown(hit)
	case toolPen:
		c.penActive = true
		c.penPoints = [][2]float32{{bx, by}}
	case toolEraser:
		c.erasing = true
		c.eraseAt(bx, by)
	case toolImage:
		c.pickImage()
	}
	return true
}

// doubleClick edits a note or region label under the pointer, or creates a
// note on empty space (the quick-capture gesture).
func (c *Canvas) doubleClick(hit *board.Item, bx, by float32) {
	switch {
	case itemKindNote(hit):
		c.beginEdit(hit)
	case hit != nil && hit.Kind == board.KindRegion && by-hit.Y <= labelBoxH && bx-hit.X <= labelBoxW:
		c.beginEdit(hit)
	case hit == nil:
		c.createNoteAt(bx, by)
		c.setTool(toolSelect)
	}
}

// selectDown selects the item under the pointer and starts a move or resize.
// On empty space it clears the selection and starts a pan.
func (c *Canvas) selectDown(hit *board.Item, bx, by float32, ev ui.InputEvent) bool {
	if hit == nil {
		c.selected = ""
		c.startPan(ev.X, ev.Y)
		return true
	}
	c.selected = hit.ID
	c.board.ToFront(hit.ID)
	if hit.Kind == board.KindConnector {
		return true
	}
	if (hit.Kind == board.KindNote || hit.Kind == board.KindImage) && nearCorner(hit, bx, by) {
		c.resizeID = hit.ID
		c.resizeW, c.resizeH = hit.W, hit.H
		c.resizeBX, c.resizeBY = bx, by
		return true
	}
	c.dragID = hit.ID
	c.dragOffX, c.dragOffY = bx-hit.X, by-hit.Y
	return true
}

// connectorDown picks the two notes of a connector: the first click selects
// a note, the second one links it to the first.
func (c *Canvas) connectorDown(hit *board.Item) {
	if !itemKindNote(hit) {
		c.connFrom = ""
		return
	}
	if c.connFrom == "" {
		c.connFrom = hit.ID
		c.selected = hit.ID
		return
	}
	if hit.ID != c.connFrom {
		c.addConnector(c.connFrom, hit.ID)
		c.setTool(toolSelect)
	}
}

func (c *Canvas) pointerMove(ctx *ui.Context, ev ui.InputEvent) bool {
	bx, by := c.windowToBoard(ev.X, ev.Y)
	switch {
	case c.penActive:
		c.addPenPoint(bx, by)
		ctx.Invalidate()
		return true
	case c.erasing:
		c.eraseAt(bx, by)
		ctx.Invalidate()
		return true
	case c.draft != nil:
		c.updateDraft(bx, by)
		ctx.Invalidate()
		return true
	case c.resizeID != "":
		it := c.boardItem(c.resizeID)
		if it != nil {
			minW, minH := noteMinW, noteMinH
			if it.Kind == board.KindImage {
				minW, minH = imageMinW, imageMinW*it.H/max(it.W, 1)
			}
			it.W = max(c.resizeW+(bx-c.resizeBX), minW)
			it.H = max(c.resizeH+(by-c.resizeBY), minH)
			ctx.Invalidate()
			c.changed()
		}
		return true
	case c.dragID != "":
		if it := c.boardItem(c.dragID); it != nil {
			it.X = bx - c.dragOffX
			it.Y = by - c.dragOffY
			ctx.Invalidate()
			c.changed()
		}
		return true
	case c.draggingP:
		c.panTo(ev.X, ev.Y)
		ctx.Invalidate()
		return true
	}
	return false
}

func (c *Canvas) pointerUp(ev ui.InputEvent) bool {
	switch ev.Button {
	case 2:
		if c.draggingP {
			c.draggingP = false
			return true
		}
	case 0:
		switch {
		case c.penActive:
			c.endPen()
		case c.erasing:
			c.erasing = false
		case c.draft != nil:
			c.finishDraft()
		case c.resizeID != "":
			c.resizeID = ""
		case c.dragID != "":
			c.dragID = ""
		case c.draggingP:
			c.draggingP = false
		default:
			return false
		}
		return true
	}
	return false
}

// updateDraft sizes the region being drawn from its start point to (bx, by).
func (c *Canvas) updateDraft(bx, by float32) {
	x0, y0 := min(c.draftX, bx), min(c.draftY, by)
	c.draft.X, c.draft.Y = x0, y0
	c.draft.W = max(c.draftX, bx) - x0
	c.draft.H = max(c.draftY, by) - y0
}

// finishDraft adds the drawn region. A plain click (no drag) makes a region
// of the default size instead.
func (c *Canvas) finishDraft() {
	d := c.draft
	c.draft = nil
	if d.W < 20 || d.H < 20 {
		c.createRegionAt(c.draftX-160, c.draftY-100, 320, 200)
	} else {
		c.createRegionAt(d.X, d.Y, max(d.W, regionMinW), max(d.H, regionMinH))
	}
	c.setTool(toolSelect)
}

// startPan begins panning from a window point.
func (c *Canvas) startPan(x, y float32) {
	c.draggingP = true
	c.panStartX, c.panStartY = x, y
}

// panTo moves the view so the drag keeps pace with the pointer.
func (c *Canvas) panTo(x, y float32) {
	c.panX += (c.panStartX - x) / c.zoom
	c.panY += (c.panStartY - y) / c.zoom
	c.panStartX, c.panStartY = x, y
	c.changed()
}

// zoomAt zooms by a factor from the wheel, keeping the point under the
// pointer fixed.
func (c *Canvas) zoomAt(wx, wy, dy float32) {
	const step = 0.0015
	factor := min(max(1+dy*step, 0.3), 3)
	newZoom := min(max(c.zoom*factor, 0.2), 5)
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

// startPen, addPenPoint and endPen draw a freehand stroke.
func (c *Canvas) addPenPoint(bx, by float32) {
	last := c.penPoints[len(c.penPoints)-1]
	dx, dy := bx-last[0], by-last[1]
	// Only add a point when it moves far enough to keep strokes light.
	if dx*dx+dy*dy < 4 {
		return
	}
	c.penPoints = append(c.penPoints, [2]float32{bx, by})
}

func (c *Canvas) endPen() {
	if len(c.penPoints) >= 2 {
		c.addStroke(c.penPoints)
	}
	c.penActive = false
	c.penPoints = nil
}
