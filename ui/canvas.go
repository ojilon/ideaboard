// Package ui builds the IdeaBoard window from the board model, using MyGo's
// native UI toolkit. This file holds the canvas: an infinite pan/zoom
// whiteboard that fills the window, plus sticky-note rendering.
package ui

import (
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

	// onChange is called whenever the board changes (to trigger autosave).
	onChange func()
	// onQuit is called when the user asks to close the app (Escape).
	onQuit func()
}

// NewCanvas makes a canvas showing b.
func NewCanvas(b *board.Board) *Canvas {
	return &Canvas{board: b, zoom: 1}
}

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
	if ctx.Theme() != nil {
		return ctx.Theme().Background
	}
	return ui.RGB(245, 245, 247)
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
		for _, l := range c.board.Layers {
			for _, it := range l.Items {
				if it.Kind == board.KindNote {
					c.drawNote(p, it)
				}
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
	col := noteColor(it.Color)
	r := ui.Rect{it.X, it.Y, it.W, it.H}
	if c.selected == it.ID {
		p.Stroke(r, ui.RGB(59, 130, 246), 4, 1.5)
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
			font.Size = 14
		}
		layout := ui.ShapeText(it.Text, font, it.W-16)
		p.TextLayout(layout, it.X+8, it.Y+8, textColor())
	}
}

func noteColor(s string) ui.Color {
	if s == "" {
		return ui.RGB(255, 245, 157) // soft yellow
	}
	col := ui.Hex(s)
	if col == (ui.Color{}) {
		return ui.RGB(255, 245, 157)
	}
	return col
}

func textColor() ui.Color {
	return ui.RGB(35, 35, 40)
}


