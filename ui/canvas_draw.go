package ui

import (
	"math"

	"ideaboard/board"

	"github.com/egoist/mygo/ui"
)

// The canvas is plain on purpose (design concepts 01-05): no grid is drawn,
// which also means fewer lines to paint each frame.

// drawBoard paints one frame: the background image (if any), then the board
// items in z-order under the pan/zoom transform, then gestures in progress.
func (c *Canvas) drawBoard(p *ui.Painter, r ui.Rect) {
	if c.bgImage != nil {
		p.Image(c.bgImage, r, ui.Cover)
		if c.bgDim > 0 {
			dimCol := ui.RGBA(0, 0, 0, c.bgDim)
			if c.curTheme != nil && luminance(c.curTheme.Background) > 127 {
				dimCol = ui.RGBA(255, 255, 255, c.bgDim)
			}
			p.Fill(r, dimCol, 0)
		}
	}

	t := ui.Translation(-c.panX, -c.panY).Then(ui.Scaling(c.zoom, c.zoom))
	p.Transform(t, 0, 0, func() {
		for _, it := range c.sortedItems(false) {
			switch it.Kind {
			case board.KindRegion:
				c.drawRegion(p, it)
			case board.KindConnector:
				c.drawConnector(p, it)
			case board.KindStroke:
				c.drawStroke(p, it)
			case board.KindImage:
				c.drawImage(p, it)
			case board.KindNote:
				c.drawNote(p, it)
			}
		}
		if c.draft != nil {
			p.StrokeDashed(ui.Rect{X: c.draft.X, Y: c.draft.Y, W: c.draft.W, H: c.draft.H}, regionOutline(c.curTheme), 10, 1)
		}
		if c.penActive && len(c.penPoints) > 1 {
			pts := make([][2]float32, len(c.penPoints))
			copy(pts, c.penPoints)
			p.StrokePolyline(pts, 2, c.strokeInk(""))
		}
	})
}

// drawNote paints a sticky note: soft fill, thin border, a small shadow, and
// its text. The text is hidden while the note is being edited.
func (c *Canvas) drawNote(p *ui.Painter, it *board.Item) {
	pal := paletteFor(c.curTheme)
	r := ui.Rect{X: it.X, Y: it.Y, W: it.W, H: it.H}
	if c.selected == it.ID {
		p.Stroke(ui.Rect{X: it.X - 3, Y: it.Y - 3, W: it.W + 6, H: it.H + 6}, pal.accent, 8, 1.5)
	}
	p.Shadow(r, 6, 0, 2, 6, 0, ui.RGBA(35, 45, 40, 0.08))
	p.Fill(r, noteFill(it.Color), 6)
	p.Stroke(r, noteBorder(it.Color), 6, 1)
	if it.ID == c.selected {
		handle := ui.Rect{X: it.X + it.W - handleSize, Y: it.Y + it.H - handleSize, W: handleSize, H: handleSize}
		p.Fill(handle, pal.accent, 2)
	}
	if it.Text != "" && it.ID != c.editID {
		font := ui.Font{Size: noteFontSize(it), Family: c.noteFont}
		layout := ui.ShapeText(it.Text, font, max(it.W-32, 40))
		p.TextLayout(layout, it.X+16, it.Y+16, noteInk)
	}
}

// drawRegion paints a dotted region outline with its label in the top-left.
func (c *Canvas) drawRegion(p *ui.Painter, it *board.Item) {
	pal := paletteFor(c.curTheme)
	r := ui.Rect{X: it.X, Y: it.Y, W: it.W, H: it.H}
	if c.selected == it.ID {
		p.Stroke(r, pal.accent, 10, 1.5)
	} else {
		p.StrokeDashed(r, regionOutline(c.curTheme), 10, 1)
	}
	if it.Label != "" && it.ID != c.editID {
		layout := ui.ShapeText(it.Label, ui.Font{Size: 14, Weight: 500}, max(it.W-32, 0))
		p.TextLayout(layout, it.X+14, it.Y+12, regionLabel(it.Color, c.curTheme))
	}
}

// drawConnector paints a dotted line between two notes. The ends sit on the
// notes' edges, so the line follows them when they move.
func (c *Canvas) drawConnector(p *ui.Painter, it *board.Item) {
	x0, y0, x1, y1, ok := c.connectorEnds(it)
	if !ok {
		return
	}
	col := connectorColor(c.curTheme)
	width := float32(1.5)
	if c.selected == it.ID {
		col = paletteFor(c.curTheme).accent
		width = 2.5
	}
	// Dots: short dashes with a steady gap, so the line reads as dotted.
	length := float32(math.Hypot(float64(x1-x0), float64(y1-y0)))
	if length == 0 {
		return
	}
	ux, uy := (x1-x0)/length, (y1-y0)/length
	const step, dot = float32(5), float32(1.6)
	for d := float32(0); d < length; d += step {
		p.Line(x0+ux*d, y0+uy*d, x0+ux*(d+dot), y0+uy*(d+dot), width, col)
	}
}

// drawStroke paints a freehand stroke in the ink color.
func (c *Canvas) drawStroke(p *ui.Painter, it *board.Item) {
	if len(it.Points) < 2 {
		return
	}
	width := it.Width
	if width == 0 {
		width = 2
	}
	pts := make([][2]float32, len(it.Points))
	copy(pts, it.Points)
	p.StrokePolyline(pts, width, c.strokeInk(it.Color))
}

// strokeInk is the color of a stroke: its own color when it has one, else
// the theme's ink so it stays visible in dark mode.
func (c *Canvas) strokeInk(color string) ui.Color {
	if col, ok := parseColor(color); ok {
		return col
	}
	return inkColor(c.curTheme)
}

// drawImage paints an image item. A missing or unreadable asset shows as a
// muted placeholder box.
func (c *Canvas) drawImage(p *ui.Painter, it *board.Item) {
	pal := paletteFor(c.curTheme)
	r := ui.Rect{X: it.X, Y: it.Y, W: it.W, H: it.H}
	p.Shadow(r, 4, 0, 2, 6, 0, ui.RGBA(35, 45, 40, 0.10))
	src := c.image(it.Asset)
	if src == nil {
		p.Fill(r, pal.surfaceMute, 4)
		p.Stroke(r, pal.line, 4, 1)
	} else {
		fit := ui.Contain
		if it.Fit == "cover" {
			fit = ui.Cover
		}
		p.Image(src, r, fit)
	}
	if c.selected == it.ID {
		p.Stroke(ui.Rect{X: it.X - 3, Y: it.Y - 3, W: it.W + 6, H: it.H + 6}, pal.accent, 6, 1.5)
	}
}
