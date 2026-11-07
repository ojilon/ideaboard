package ui

import (
	"math"
	"sort"
	"strings"

	"ideaboard/board"

	"github.com/egoist/mygo/ui"
)

// Sizes and tolerances, in board DIPs.
const (
	noteMinW   float32 = 180
	noteMinH   float32 = 112
	imageMinW  float32 = 40
	regionMinW float32 = 80
	regionMinH float32 = 60
	handleSize float32 = 8   // resize handle square at a bottom-right corner
	edgeBand   float32 = 8   // grab band along a region's edge
	labelBoxW  float32 = 120 // grab area of a region's label
	labelBoxH  float32 = 32
	eraserR    float32 = 10
	connGap    float32 = 4 // gap between a connector and its notes
)

// windowToBoard converts a window point to board coordinates.
func (c *Canvas) windowToBoard(wx, wy float32) (float32, float32) {
	return wx/c.zoom + c.panX, wy/c.zoom + c.panY
}

// boardItem returns the item with the given id, or nil.
func (c *Canvas) boardItem(id string) *board.Item {
	if id == "" {
		return nil
	}
	return c.board.Item(id)
}

// sortedItems returns the items in draw order (ascending z), or the reverse
// (front-most first) when front is true.
func (c *Canvas) sortedItems(front bool) []*board.Item {
	items := make([]*board.Item, len(c.board.Items()))
	copy(items, c.board.Items())
	sort.SliceStable(items, func(i, j int) bool {
		if front {
			return items[i].Z > items[j].Z
		}
		return items[i].Z < items[j].Z
	})
	return items
}

// hitItem returns the front-most item under board point (bx, by). Regions
// are grabbed by their label or edge band, so notes inside stay usable.
func (c *Canvas) hitItem(bx, by float32) *board.Item {
	for _, it := range c.sortedItems(true) {
		if c.hits(it, bx, by) {
			return it
		}
	}
	return nil
}

// hits reports whether board point (bx, by) grabs the item.
func (c *Canvas) hits(it *board.Item, bx, by float32) bool {
	switch it.Kind {
	case board.KindNote, board.KindImage:
		return bx >= it.X && bx <= it.X+it.W && by >= it.Y && by <= it.Y+it.H
	case board.KindRegion:
		if bx < it.X || bx > it.X+it.W || by < it.Y || by > it.Y+it.H {
			return false
		}
		onLabel := bx-it.X <= labelBoxW && by-it.Y <= labelBoxH
		onEdge := bx-it.X <= edgeBand || it.X+it.W-bx <= edgeBand ||
			by-it.Y <= edgeBand || it.Y+it.H-by <= edgeBand
		return onLabel || onEdge
	case board.KindConnector:
		x0, y0, x1, y1, ok := c.connectorEnds(it)
		if !ok {
			return false
		}
		return distToSegment(bx, by, x0, y0, x1, y1) <= 6/c.zoom
	}
	return false
}

// nearCorner reports whether (bx, by) is in an item's bottom-right resize handle.
func nearCorner(it *board.Item, bx, by float32) bool {
	return bx >= it.X+it.W-handleSize && bx <= it.X+it.W && by >= it.Y+it.H-handleSize && by <= it.Y+it.H
}

// connectorEnds returns the segment a connector draws: from the edge of its
// first note to the edge of the second, along the line between their centers.
func (c *Canvas) connectorEnds(it *board.Item) (x0, y0, x1, y1 float32, ok bool) {
	from, to := c.boardItem(it.From), c.boardItem(it.To)
	if from == nil || to == nil {
		return 0, 0, 0, 0, false
	}
	cx0, cy0 := from.X+from.W/2, from.Y+from.H/2
	cx1, cy1 := to.X+to.W/2, to.Y+to.H/2
	dx, dy := cx1-cx0, cy1-cy0
	length := float32(math.Hypot(float64(dx), float64(dy)))
	if length < 1 {
		return 0, 0, 0, 0, false
	}
	t0 := exitParam(from, dx, dy)
	t1 := exitParam(to, dx, dy)
	ux, uy := dx/length*connGap, dy/length*connGap
	x0, y0 = cx0+dx*t0+ux, cy0+dy*t0+uy
	x1, y1 = cx1-dx*t1-ux, cy1-dy*t1-uy
	return x0, y0, x1, y1, true
}

// exitParam returns the t at which a ray from the center of it, along the
// direction (dx, dy) scaled by t, leaves its box.
func exitParam(it *board.Item, dx, dy float32) float32 {
	t := float32(math.Inf(1))
	if dx != 0 {
		t = min(t, it.W/2/absf(dx))
	}
	if dy != 0 {
		t = min(t, it.H/2/absf(dy))
	}
	return t
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// distToSegment is the distance from (px, py) to the segment (x0,y0)-(x1,y1).
func distToSegment(px, py, x0, y0, x1, y1 float32) float32 {
	dx, dy := x1-x0, y1-y0
	l2 := dx*dx + dy*dy
	t := float32(0)
	if l2 > 0 {
		t = ((px-x0)*dx + (py-y0)*dy) / l2
		t = min(max(t, 0), 1)
	}
	ex, ey := x0+t*dx-px, y0+t*dy-py
	return float32(math.Sqrt(float64(ex*ex + ey*ey)))
}

// record adds an undoable change and clears the redo stack.
func (c *Canvas) record(ch change) {
	c.history = append(c.history, ch)
	c.redo = nil
}

// addItem adds it to the board as an undoable change and returns it.
func (c *Canvas) addItem(it *board.Item) *board.Item {
	c.board.AddItem(it)
	c.record(change{add: true, it: it})
	c.changed()
	return it
}

// removeItem removes an item as an undoable change.
func (c *Canvas) removeItem(it *board.Item) {
	c.board.RemoveItem(it.ID)
	if c.selected == it.ID {
		c.selected = ""
	}
	c.record(change{add: false, it: it})
	c.changed()
}

// deleteSelected removes the selected item, if any.
func (c *Canvas) deleteSelected() bool {
	it := c.boardItem(c.selected)
	if it == nil {
		return false
	}
	c.removeItem(it)
	return true
}

// undo reverts the most recent add or removal.
func (c *Canvas) undo() {
	if len(c.history) == 0 {
		return
	}
	ch := c.history[len(c.history)-1]
	c.history = c.history[:len(c.history)-1]
	c.apply(ch, true)
	c.redo = append(c.redo, ch)
	c.changed()
}

// redoLast re-applies the most recently undone change.
func (c *Canvas) redoLast() {
	if len(c.redo) == 0 {
		return
	}
	ch := c.redo[len(c.redo)-1]
	c.redo = c.redo[:len(c.redo)-1]
	c.apply(ch, false)
	c.history = append(c.history, ch)
	c.changed()
}

// apply performs a change in one direction. undoing reverses it. An added
// item is present after the change is done; a removed one is present after
// its removal is undone.
func (c *Canvas) apply(ch change, undoing bool) {
	wantPresent := ch.add != undoing
	present := c.board.Item(ch.it.ID) != nil
	switch {
	case wantPresent && !present:
		c.board.Restore(ch.it)
	case !wantPresent && present:
		c.board.RemoveItem(ch.it.ID)
		if c.selected == ch.it.ID {
			c.selected = ""
		}
	}
}

// createNoteAt adds a note centered on (bx, by) and opens it for typing.
func (c *Canvas) createNoteAt(bx, by float32) {
	it := c.addItem(&board.Item{
		Kind:     board.KindNote,
		X:        bx - noteMinW/2,
		Y:        by - noteMinH/2,
		W:        noteMinW,
		H:        noteMinH,
		Color:    noteStyles[0].fill,
		FontSize: 16,
	})
	c.selected = it.ID
	c.beginEdit(it)
}

// createRegionAt adds a labeled region with the given box.
func (c *Canvas) createRegionAt(x, y, w, h float32) *board.Item {
	n := 0
	for _, it := range c.board.Items() {
		if it.Kind == board.KindRegion {
			n++
		}
	}
	it := c.addItem(&board.Item{
		Kind: board.KindRegion,
		X:    x, Y: y, W: w, H: h,
		Color: regionLabels[n%len(regionLabels)],
		Label: "Region",
	})
	c.selected = it.ID
	return it
}

// addConnector links two notes with a connector.
func (c *Canvas) addConnector(from, to string) {
	it := c.addItem(&board.Item{Kind: board.KindConnector, From: from, To: to})
	c.selected = it.ID
}

// addStroke stores a finished pen stroke.
func (c *Canvas) addStroke(points [][2]float32) {
	minX, minY := points[0][0], points[0][1]
	maxX, maxY := minX, minY
	for _, pt := range points {
		minX, maxX = min(minX, pt[0]), max(maxX, pt[0])
		minY, maxY = min(minY, pt[1]), max(maxY, pt[1])
	}
	c.addItem(&board.Item{
		Kind: board.KindStroke,
		X:    minX, Y: minY,
		W: maxX - minX, H: maxY - minY,
		Width:  2,
		Points: points,
	})
}

// eraseAt removes every pen stroke that passes within reach of (bx, by).
// It checks the segments between samples, so a fast drag still erases.
func (c *Canvas) eraseAt(bx, by float32) {
	for _, it := range c.board.Items() {
		if it.Kind != board.KindStroke {
			continue
		}
		reach := eraserR + it.Width
		for i := 0; i+1 < len(it.Points) || (i == 0 && len(it.Points) == 1); i++ {
			a := it.Points[i]
			b := a
			if i+1 < len(it.Points) {
				b = it.Points[i+1]
			}
			if distToSegment(bx, by, a[0], a[1], b[0], b[1]) <= reach {
				c.removeItem(it)
				break
			}
		}
	}
}

// beginEdit opens an item's text (or region label) for in-place editing.
func (c *Canvas) beginEdit(it *board.Item) {
	c.editID = it.ID
	c.selected = it.ID
	c.editRegion = it.Kind == board.KindRegion
	if c.editRegion {
		c.editText = it.Label
	} else {
		c.editText = it.Text
	}
}

// commitEdit applies the edited text and leaves edit mode. A note left
// empty is removed, so an accidental double-click leaves no clutter behind.
func (c *Canvas) commitEdit() {
	if c.editID == "" {
		return
	}
	it := c.boardItem(c.editID)
	text := c.editText
	c.editID, c.editText = "", ""
	if it == nil {
		return
	}
	if it.Kind == board.KindRegion {
		if strings.TrimSpace(text) != "" {
			it.Label = strings.TrimSpace(text)
		}
		c.changed()
		return
	}
	if strings.TrimSpace(text) == "" {
		c.removeItem(it)
		return
	}
	it.Text = text
	c.changed()
}

// editOverlay places a text area over the item being edited.
func (c *Canvas) editOverlay(ctx *ui.Context) {
	it := c.boardItem(c.editID)
	if it == nil {
		return
	}
	var x, y, w, h, size float32
	if c.editRegion {
		x, y = (it.X+12-c.panX)*c.zoom, (it.Y+10-c.panY)*c.zoom
		w, h = 160*c.zoom, 24*c.zoom
		size = 14 * c.zoom
	} else {
		x, y = (it.X-c.panX)*c.zoom, (it.Y-c.panY)*c.zoom
		w, h = it.W*c.zoom, it.H*c.zoom
		size = noteFontSize(it) * c.zoom
	}
	ui.Box(ctx).Absolute().Left(x).Top(y).Size(w, h).Children(func() {
		ui.TextArea(ctx, &c.editText).
			FontSize(size).
			FillWidth().FillHeight().
			OnShortcut(0, ui.KeyEscape, func() {
				c.commitEdit()
			})
	})
}

// colorPicker shows the note palette next to the right-clicked note.
func (c *Canvas) colorPicker(ctx *ui.Context) {
	it := c.boardItem(c.colorPickID)
	if it == nil {
		return
	}
	pal := paletteFor(c.curTheme)
	ui.Box(ctx).Absolute().Left(c.colorPickX).Top(c.colorPickY).
		Size(3*28+2*6+16, 2*28+6+16).Background(pal.surface).Radius(10).
		Border(1, pal.line).
		Children(func() {
			ui.Column(ctx).Gap(6).Padding(8).Children(func() {
				for row := 0; row < 2; row++ {
					ui.Row(ctx).Gap(6).Children(func() {
						for col := 0; col < 3; col++ {
							s := noteStyles[row*3+col]
							c.colorSwatch(ctx, it, s)
						}
					})
				}
			})
		})
}

// colorSwatch is one palette entry in the color picker.
func (c *Canvas) colorSwatch(ctx *ui.Context, it *board.Item, s noteStyle) {
	ui.Box(ctx).Size(28, 28).Background(mustHex(s.fill)).Radius(6).
		Border(1, mustHex(s.border)).Tooltip(s.name).
		HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputPointerDown && ev.Button == 0 {
				it.Color = s.fill
				c.colorPickID = ""
				c.changed()
				return true
			}
			return false
		})
}

// noteFontSize is a note's text size, with the default for unset notes.
func noteFontSize(it *board.Item) float32 {
	if it.FontSize == 0 {
		return 16
	}
	return it.FontSize
}

// AddImage caches an image under its asset id and adds it as a board item.
// The caller places the item (position and size) before calling.
func (c *Canvas) AddImage(assetID string, it *board.Item, src ui.ImageSource) {
	if src == nil {
		return
	}
	c.imageCache[assetID] = src
	it.Kind = board.KindImage
	it.Asset = assetID
	c.addItem(it)
	c.selected = it.ID
}

// image returns the decoded bitmap for an asset, loading it once. It returns
// nil when the asset cannot be loaded.
func (c *Canvas) image(asset string) ui.ImageSource {
	if src, ok := c.imageCache[asset]; ok {
		return src
	}
	var src ui.ImageSource
	if c.loadAsset != nil {
		src = c.loadAsset(asset)
	}
	c.imageCache[asset] = src
	return src
}
