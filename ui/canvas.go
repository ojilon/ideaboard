// Package ui builds the IdeaBoard window from the board model, using MyGo's
// native UI toolkit. This file holds the canvas: an infinite pan/zoom
// whiteboard that fills the window.
package ui

import (
	"ideaboard/board"

	"github.com/egoist/mygo/ui"
)

// Canvas is a full-window pan/zoom board. It holds pan/zoom state and turns
// pointer input into panning (middle-drag, space+drag) and zooming
// (ctrl+scroll).
type Canvas struct {
	board *board.Board // the model being shown

	panX, panY float32 // board coords of the window's top-left
	zoom       float32

	draggingP bool
	panStartX, panStartY float32
	spaceHeld bool
}

// NewCanvas makes a canvas showing b.
func NewCanvas(b *board.Board) *Canvas {
	return &Canvas{board: b, zoom: 1}
}

// View builds the canvas element for one frame.
func (c *Canvas) View(ctx *ui.Context) ui.Element {
	return ui.Box(ctx).Fill().
		Background(boardColor(ctx)).
		Draw(func(p *ui.Painter, r ui.Rect) {
			c.drawBoard(p, r)
		}).
		HandleInput(func(ev ui.InputEvent) bool {
			return c.input(ctx, ev)
		})
}

func boardColor(ctx *ui.Context) ui.Color {
	if ctx.Theme() != nil {
		return ctx.Theme().Background
	}
	return ui.RGB(245, 245, 247)
}
