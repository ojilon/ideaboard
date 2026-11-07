// Package ui builds the IdeaBoard window from the board model, using MyGo's
// native UI toolkit. This file holds the canvas: an infinite pan/zoom
// whiteboard that fills the window, plus sticky-note rendering.
package ui

import (
	"math"
	"sort"

	"ideaboard/board"

	"github.com/egoist/mygo/ui"
)

// Canvas is a full-window pan/zoom board. It holds pan/zoom state and turns
// pointer input into panning (middle-drag, space+drag), zooming
// (ctrl+scroll), note creation, note drag and note edit.
type Canvas struct {
	board *board.Board // the model being shown

	panX, panY float32 // board coords of the window's top-left
	zoom       float32

	// Panning.
	draggingP bool
	panStartX, panStartY float32
	spaceHeld bool

	// Notes.
	// selected is the id of the note under the pointer / last clicked.
	selected string
	// editID is the note being edited in place ("" when none).
	editID string
	// editText is the working text for the note being edited.
	editText string
	// dragID is the note being dragged ("" when none).
	dragID string
	// dragOff is the pointer offset into the dragged note, in board coords.
	dragOffX, dragOffY float32
	// resizeID is the note being resized ("" when none).
	resizeID string
	// resizeStart is the note's W/H and the pointer board pos when resize began.
	resizeStartW, resizeStartH float32
	resizeStartBX, resizeStartBY float32

	// colorPickID is the note being recolored ("" when none).
	colorPickID string
	// colorPickX/Y is where to show the color picker overlay, in window coords.
	colorPickX, colorPickY float32

	// curTheme is the theme of the current frame (set in View).
	curTheme *ui.Theme

	// Pen tool (freehand drawing).
	penActive bool
	penPoints [][2]float32 // board coords of the stroke being drawn

	// Undo/redo stacks for strokes (stroke items, most recent last).
	undone []*board.Item

	// imageCache holds decoded images by asset id.
	imageCache map[string]ui.ImageSource

	// onChange is called whenever the board changes (to trigger autosave).
	onChange func()

	// onAttachImage opens an image file and attaches it to the board.
	onAttachImage func()
	// onQuit is called when the user asks to close the app (Escape).
	onQuit func()
}

// OnAttachImage sets the attach-image hook.
func (c *Canvas) OnAttachImage(fn func()) { c.onAttachImage = fn }

// NewCanvas makes a canvas showing b.
func NewCanvas(b *board.Board) *Canvas {
	return &Canvas{board: b, zoom: 1, imageCache: map[string]ui.ImageSource{}}
}

// PanX and PanY return the board coords of the window's top-left.
func (c *Canvas) PanX() float32 { return c.panX }
func (c *Canvas) PanY() float32 { return c.panY }

// OnChange sets the change hook.
func (c *Canvas) OnChange(fn func()) { c.onChange = fn }

// OnQuit sets the quit hook.
func (c *Canvas) OnQuit(fn func()) { c.onQuit = fn }

func (c *Canvas) changed() {
	if c.onChange != nil {
		c.onChange()
	}
}

// View builds the canvas element for one frame.
func (c *Canvas) View(ctx *ui.Context) ui.Element {
	c.curTheme = ctx.Theme()
	box := ui.Box(ctx).Fill().
		Background(boardColor(ctx)).
		Draw(func(p *ui.Painter, r ui.Rect) {
			c.drawBoard(p, r)
		}).
		HandleInput(func(ev ui.InputEvent) bool {
			return c.input(ctx, ev)
		})

	box.Children(func() {
		if c.editID != "" {
			it := c.boardNote(c.editID)
			if it != nil {
				wx := (it.X - c.panX) * c.zoom
				wy := (it.Y - c.panY) * c.zoom
				ww := it.W * c.zoom
				wh := it.H * c.zoom
				ui.Box(ctx).Absolute().Left(wx).Top(wy).Size(ww, wh).
					Children(func() {
						ui.TextArea(ctx, &c.editText).
							FontSize(it.FontSize).
							FillWidth().FillHeight().
							OnShortcut(0, ui.KeyEscape, func() {
								c.commitEdit()
							})
					})
			}
			return
		}
		if c.colorPickID != "" {
			it := c.boardNote(c.colorPickID)
			if it != nil {
				c.colorPicker(ctx, it)
			}
			return
		}
	})

	return box
}

func (c *Canvas) colorPicker(ctx *ui.Context, it *board.Item) {
	ui.Box(ctx).Absolute().Left(c.colorPickX).Top(c.colorPickY).
		Size(132, 120).Background(ctx.Theme().Surface).Radius(6).
		Children(func() {
			ui.Column(ctx).Gap(6).Padding(8).Children(func() {
				for _, col := range softColors {
					c.colorSwatch(ctx, it, col)
				}
			})
		})
}

func (c *Canvas) colorSwatch(ctx *ui.Context, it *board.Item, col string) {
	ui.Box(ctx).Size(28, 28).Background(ui.Hex(col)).Radius(6).
		HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputPointerDown && ev.Button == 0 {
				it.Color = col
				c.colorPickID = ""
				c.changed()
				return true
			}
			return false
		})
}

func boardColor(ctx *ui.Context) ui.Color {
	return canvasColor(ctx.Theme())
}

// boardNote returns the note with the given id, or nil.
func (c *Canvas) boardNote(id string) *board.Item {
	for _, l := range c.board.Layers {
		for _, it := range l.Items {
			if it.ID == id && it.Kind == board.KindNote {
				return it
			}
		}
	}
	return nil
}

// nearCorner reports whether board point (bx, by) is in the note's
// bottom-right resize handle (8 board DIPs square at the corner).
func (c *Canvas) nearCorner(it *board.Item, bx, by float32) bool {
	return bx >= it.X+it.W-8 && bx <= it.X+it.W && by >= it.Y+it.H-8 && by <= it.Y+it.H
}

func (c *Canvas) drawRegion(p *ui.Painter, it *board.Item) {
	th := c.curTheme
	r := ui.Rect{it.X, it.Y, it.W, it.H}
	col := regionColor(th)
	// Dotted outline, 1-1.5 px.
	p.StrokeDashed(r, col, 4, 1)
	// Label inside the upper-left boundary.
	if it.Label != "" {
		layout := ui.ShapeText(it.Label, ui.Font{Size: 14}, max(it.W-32, 0))
		p.TextLayout(layout, it.X+12, it.Y+12, col)
	}
}

func (c *Canvas) drawConnector(p *ui.Painter, it *board.Item) {
	th := c.curTheme
	from := c.boardNote(it.From)
	to := c.boardNote(it.To)
	if from == nil || to == nil {
		return
	}
	// A dotted line between the two note centers; it follows the notes
	// because endpoints are recomputed from their positions every frame.
	x0, y0 := from.X+from.W/2, from.Y+from.H/2
	x1, y1 := to.X+to.W/2, to.Y+to.H/2
	c.drawDashedLine(p, x0, y0, x1, y1, 1.5, connectorColor(th))
}

func (c *Canvas) drawDashedLine(p *ui.Painter, x0, y0, x1, y1, width float32, col ui.Color) {
	dx, dy := x1-x0, y1-y0
	length := float32(0)
	if dx != 0 || dy != 0 {
		length = float32(math.Sqrt(float64(dx*dx + dy*dy)))
	}
	if length == 0 {
		return
	}
	ux, uy := dx/length, dy/length
	dash, gap := float32(6), float32(4)
	for d := float32(0); d < length; d += dash + gap {
		d0 := d
		d1 := min(d+dash, length)
		p.Line(x0+ux*d0, y0+uy*d0, x0+ux*d1, y0+uy*d1, width, col)
	}
}

func (c *Canvas) drawStroke(p *ui.Painter, it *board.Item) {
	if len(it.Points) == 0 {
		return
	}
	// Eraser strokes punch out by drawing with the canvas color over the
	// board; pen strokes draw in the stroke color.
	col := noteFill(it.Color)
	if it.Width == 0 {
		it.Width = 2
	}
	if eraser(it) {
		col = boardColorTheme(c.curTheme)
	}
	pts := make([][2]float32, len(it.Points))
	copy(pts, it.Points)
	p.StrokePolyline(pts, it.Width, col)
}

func boardColorTheme(th *ui.Theme) ui.Color { return canvasColor(th) }

func eraser(it *board.Item) bool { return it.Color == "_eraser" }

func (c *Canvas) drawImage(p *ui.Painter, it *board.Item) {
	src := c.imageCache[it.Asset]
	if src == nil {
		return
	}
	fit := ui.Contain
	if it.Fit == "cover" {
		fit = ui.Cover
	}
	p.Image(src, ui.Rect{it.X, it.Y, it.W, it.H}, fit)
}

// AddImage decodes and caches an image by asset id and adds it as a board
// item. The caller owns placing the item (position/size) before calling.
func (c *Canvas) AddImage(assetID string, it *board.Item, src ui.ImageSource) {
	if src == nil {
		return
	}
	c.imageCache[assetID] = src
	c.board.AddItem(it)
	c.changed()
}

// softColors are the few note colors offered.
var softColors = []string{
	"#fff59d", // soft yellow
	"#c8e6c9", // soft green
	"#bbdefb", // soft blue
	"#ffcc80", // soft orange
	"#e1bee7", // soft purple
	"#ffcdd2", // soft red
}

// drawBoard paints the board: a pan/zoom transform over a light grid plus
// the sticky notes.
func (c *Canvas) drawBoard(p *ui.Painter, r ui.Rect) {
	t := ui.Translation(-c.panX, -c.panY).Then(ui.Scaling(c.zoom, c.zoom))
	p.Transform(t, 0, 0, func() {
		c.drawGrid(p, r)
		items := c.board.Items()
		// Draw in z-order so regions sit behind notes and connectors follow.
		sorted := make([]*board.Item, len(items))
		copy(sorted, items)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Z < sorted[j].Z })
		for _, it := range sorted {
			switch it.Kind {
			case board.KindNote:
				c.drawNote(p, it)
			case board.KindRegion:
				c.drawRegion(p, it)
			case board.KindConnector:
				c.drawConnector(p, it)
			case board.KindStroke:
				c.drawStroke(p, it)
			case board.KindImage:
				c.drawImage(p, it)
			}
		}
	})
}

func (c *Canvas) drawGrid(p *ui.Painter, r ui.Rect) {
	vx0 := c.panX
	vx1 := c.panX + r.W/c.zoom
	vy0 := c.panY
	vy1 := c.panY + r.H/c.zoom
	step := float32(100)
	gc := ui.RGBA(0, 0, 0, 0.06)
	x0 := floorTo(vx0, step)
	for x := x0; x <= vx1; x += step {
		p.Line(x, vy0, x, vy1, 1, gc)
	}
	y0 := floorTo(vy0, step)
	for y := y0; y <= vy1; y += step {
		p.Line(vx0, y, vx1, y, 1, gc)
	}
}

func floorTo(v, step float32) float32 {
	return float32(int(v/step)) * step
}

func (c *Canvas) drawNote(p *ui.Painter, it *board.Item) {
	col := noteFill(it.Color)
	r := ui.Rect{it.X, it.Y, it.W, it.H}
	th := c.curTheme
	if c.selected == it.ID {
		p.Stroke(r, th.Accent, 6, 1.5)
	}
	p.Fill(r, col, 6)
	// Resize handle on the selected note's bottom-right corner.
	if c.selected == it.ID {
		handle := ui.Rect{it.X + it.W - 8, it.Y + it.H - 8, 8, 8}
		p.Fill(handle, ui.RGBA(0, 0, 0, 0.35), 2)
	}
	if it.Text != "" {
		font := ui.Font{Size: it.FontSize}
		if font.Size == 0 {
			font.Size = 16
		}
		layout := ui.ShapeText(it.Text, font, max(it.W-32, 40))
		p.TextLayout(layout, it.X+16, it.Y+16, inkColor(th))
	}
}


