package ui

import "github.com/egoist/mygo/ui"

const (
	toolBtn    float32 = 36 // size of one toolbar button
	toolGap    float32 = 2
	toolPad    float32 = 6
	headerTop  float32 = 14
	headerLeft float32 = 20
	divW       float32 = 14 // width of the divider between tool groups
)

// header draws the app identity and the board name at the top-left. It is a
// window drag handle, because the window is frameless.
func (c *Canvas) header(ctx *ui.Context) {
	ink := inkColor(c.curTheme)
	ui.Row(ctx).Absolute().Left(headerLeft).Top(headerTop).Height(36).Gap(10).
		DragWindow().Children(func() {
		ui.Box(ctx).Size(20, 20).Draw(func(p *ui.Painter, r ui.Rect) {
			drawLogo(p, r, paletteFor(c.curTheme).accent)
		})
		ui.Text(ctx, "Ideaboard").FontSize(15).FontWeight(600).TextColor(ink)
		ui.Box(ctx).Size(1, 16).Background(paletteFor(c.curTheme).line)
		ui.Text(ctx, c.board.Name).FontSize(13).TextColor(mutedColor(c.curTheme))
	})
}

// drawLogo paints the two-leaf Ideaboard mark inside r.
func drawLogo(p *ui.Painter, r ui.Rect, col ui.Color) {
	s := r.W / 2
	p.Fill(ui.Rect{X: r.X, Y: r.Y, W: s * 1.3, H: s * 1.3}, col, s*0.6)
	p.Fill(ui.Rect{X: r.X + s*0.7, Y: r.Y + s*0.7, W: s * 1.3, H: s * 1.3}, ui.RGBA(col.R, col.G, col.B, 0.6), s*0.6)
}

// toolbar draws the slim floating toolbar at the top center of the canvas.
func (c *Canvas) toolbar(ctx *ui.Context, viewW float32) {
	pal := paletteFor(c.curTheme)
	type btn struct {
		icon   *ui.SVG
		label  string
		active bool
		onTap  func()
	}
	btns := []btn{
		{iconSelect, toolName[toolSelect], c.tool == toolSelect, func() { c.setTool(toolSelect) }},
		{iconNote, toolName[toolNote], c.tool == toolNote, func() { c.setTool(toolNote) }},
		{iconRegion, toolName[toolRegion], c.tool == toolRegion, func() { c.setTool(toolRegion) }},
		{iconConnector, toolName[toolConnector], c.tool == toolConnector, func() { c.setTool(toolConnector) }},
		{iconPen, toolName[toolPen], c.tool == toolPen, func() { c.setTool(toolPen) }},
		{iconEraser, toolName[toolEraser], c.tool == toolEraser, func() { c.setTool(toolEraser) }},
		{iconImage, toolName[toolImage], c.tool == toolImage, func() { c.setTool(toolImage) }},
	}
	// Left group: tools. Right group: boards and settings. Sizes are fixed,
	// so the bar's width is known before it is built.
	n := len(btns)
	// n tools + divider + 2 right-hand buttons, with a gap between each of the n+3 children.
	barW := toolPad*2 + float32(n+2)*toolBtn + divW + float32(n+2)*toolGap
	barH := toolBtn + toolPad*2
	left := (viewW - barW) / 2
	top := headerTop

	// Shadow sits in its own box beneath the bar (design system 5.3).
	ui.Box(ctx).Absolute().Left(left).Top(top).Size(barW, barH).Radius(10).
		Draw(func(p *ui.Painter, r ui.Rect) {
			p.Shadow(r, 10, 0, 4, 16, 0, ui.RGBA(35, 45, 40, 0.08))
		})

	ui.Box(ctx).Absolute().Left(left).Top(top).Size(barW, barH).
		Background(pal.surface).Radius(10).Border(1, pal.line).
		Children(func() {
			ui.Row(ctx).Padding(toolPad).Gap(toolGap).Children(func() {
				for _, b := range btns {
					iconButton(ctx, b.icon, b.label, b.active, pal, b.onTap)
				}
				ui.Box(ctx).Size(divW, toolBtn).Draw(func(p *ui.Painter, r ui.Rect) {
					p.Fill(ui.Rect{X: r.X + divW/2, Y: r.Y + 8, W: 1, H: toolBtn - 16}, pal.line, 0)
				})
				iconButton(ctx, iconBoards, "All boards", false, pal, c.onBoards)
				iconButton(ctx, iconSettings, "Settings (Ctrl+,)", c.showSettings, pal, c.onSettingsToggle)
			})
		})
}

// iconButton is one toolbar button: an icon with a tooltip, a soft fill
// when active, and a click on the primary button.
func iconButton(ctx *ui.Context, icon *ui.SVG, label string, active bool, pal palette, onTap func()) {
	bg := ui.RGBA(0, 0, 0, 0)
	col := pal.inkPrimary
	if active {
		bg = pal.accentSoft
		col = pal.focus
	}
	ui.Box(ctx).Size(toolBtn, toolBtn).Radius(5).Background(bg).Tooltip(label).
		Draw(func(p *ui.Painter, r ui.Rect) {
			p.Icon(icon, ui.Rect{X: r.X + 9, Y: r.Y + 9, W: 18, H: 18}, col)
		}).
		HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputPointerDown && ev.Button == 0 && onTap != nil {
				onTap()
				return true
			}
			return false
		})
}

// setTool switches tools and drops any half-finished gesture.
func (c *Canvas) setTool(t tool) {
	c.tool = t
	c.connFrom = ""
	c.draft = nil
	c.penActive = false
	c.penPoints = nil
	c.erasing = false
}
