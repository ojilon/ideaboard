package ui

import (
	"github.com/egoist/mygo/ui"
)

// drawBoard paints the board: a pan/zoom transform over a light grid so the
// infinite canvas is visible as you pan and zoom.
func (c *Canvas) drawBoard(p *ui.Painter, r ui.Rect) {
	// Map board coordinates into window coordinates:
	//   window = (board - pan) * zoom
	t := ui.Translation(-c.panX, -c.panY).Then(ui.Scaling(c.zoom, c.zoom))
	p.Transform(t, 0, 0, func() {
		// Grid every 100 board DIPs, limited to the visible window.
		// Visible board range:
		vx0 := c.panX
		vx1 := c.panX + r.W/c.zoom
		vy0 := c.panY
		vy1 := c.panY + r.H/c.zoom
		step := float32(100)
		gc := p.Theme().Border
		if gc == (ui.Color{}) {
			gc = ui.RGB(200, 200, 200)
		}
		// faint grid
		gc = ui.RGBA(0, 0, 0, 0.08)

		x0 := floorTo(vx0, step)
		for x := x0; x <= vx1; x += step {
			p.Line(x, vy0, x, vy1, 1, gc)
		}
		y0 := floorTo(vy0, step)
		for y := y0; y <= vy1; y += step {
			p.Line(vx0, y, vx1, y, 1, gc)
		}
	})
}

func floorTo(v, step float32) float32 {
	return (float32(int(v/step)) * step)
}

// input turns pointer and key events into pan and zoom.
func (c *Canvas) input(ctx *ui.Context, ev ui.InputEvent) bool {
	switch ev.Kind {
	case ui.InputKeyDown:
		if ev.Key == ui.KeySpace {
			c.spaceHeld = true
			return true
		}
	case ui.InputKeyUp:
		if ev.Key == ui.KeySpace {
			c.spaceHeld = false
			return true
		}
	case ui.InputPointerDown:
		if ev.Button == 2 {
			// middle button: pan
			c.startPan(ev.X, ev.Y)
			return true
		}
		if ev.Button == 0 && c.spaceHeld {
			c.startPan(ev.X, ev.Y)
			return true
		}
	case ui.InputPointerMove:
		if c.draggingP {
			c.panTo(ev.X, ev.Y)
			ctx.Invalidate()
			return true
		}
	case ui.InputPointerUp:
		if c.draggingP && (ev.Button == 2 || (ev.Button == 0 && c.spaceHeld)) {
			c.draggingP = false
			return true
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

func (c *Canvas) startPan(x, y float32) {
	c.draggingP = true
	c.panStartX, c.panStartY = x, y
}

func (c *Canvas) panTo(x, y float32) {
	// Window moved by dx; board moves opposite in board coords.
	dx := (c.panStartX - x) / c.zoom
	dy := (c.panStartY - y) / c.zoom
	c.panX += dx
	c.panY += dy
	c.panStartX, c.panStartY = x, y
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
	// Keep the board point under the pointer fixed.
	invOld := 1 / c.zoom
	invNew := 1 / newZoom
	c.panX += wx * (invOld - invNew)
	c.panY += wy * (invOld - invNew)
	c.zoom = newZoom
}
