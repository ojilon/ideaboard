// Package settings renders the settings panel (theme, font, background,
// window opacity) and holds its persisted values.
package settings

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// Callbacks the panel reports to the shell.
type Callbacks struct {
	OnTheme          func(Theme)
	OnFontFamily     func(string)
	OnFontFile       func()
	OnPickBackground func()
	OnBackground     func(path string, dim float32) // path "" removes the background
	OnOpacity        func(opacity float32)
	OnClose          func()
}

const panelW float32 = 340

var closeIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#000" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"/></svg>`))

// Panel builds the floating settings panel on the right of the window.
func Panel(ctx *ui.Context, s Settings, cb Callbacks) ui.Element {
	w, _ := ctx.Size()
	fontFamily := s.FontFamily
	dim := float64(s.BackgroundDim)
	opacity := float64(s.WindowOpacity)

	return ui.Box(ctx).Absolute().
		Left(w-panelW-16).Top(64).Width(panelW).
		Background(panelBg(ctx)).Radius(10).Border(1, border(ctx)).
		Padding(20).
		Children(func() {
			ui.Column(ctx).Gap(20).Children(func() {
				ui.Row(ctx).Children(func() {
					ui.Text(ctx, "Settings").FontSize(18).FontWeight(600)
					ui.Box(ctx).Grow(1)
					ui.Box(ctx).Size(28, 28).Radius(5).Tooltip("Close settings").
						Draw(func(p *ui.Painter, r ui.Rect) {
							p.Icon(closeIcon, ui.Rect{X: r.X + 6, Y: r.Y + 6, W: 16, H: 16}, muted(ctx))
						}).
						HandleInput(func(ev ui.InputEvent) bool {
							if ev.Kind == ui.InputPointerDown && ev.Button == 0 {
								cb.OnClose()
								return true
							}
							return false
						})
				})

				section(ctx, "Theme", func() {
					ui.Row(ctx).Gap(0).Children(func() {
						themeButton(ctx, "System", ThemeSystem, s.Theme, cb.OnTheme)
						themeButton(ctx, "Light", ThemeLight, s.Theme, cb.OnTheme)
						themeButton(ctx, "Dark", ThemeDark, s.Theme, cb.OnTheme)
					})
				})

				section(ctx, "Note font", func() {
					ui.Row(ctx).Gap(8).Children(func() {
						ui.TextInput(ctx, &fontFamily).Placeholder("Font family").Width(200).
							OnChange(func() { cb.OnFontFamily(fontFamily) })
						if ui.Button(ctx, "Choose font…").Width(112).Clicked() {
							cb.OnFontFile()
						}
					})
				})

				section(ctx, "Background image", func() {
					ui.Text(ctx, shortPath(s.Background)).FontSize(12).TextColor(muted(ctx)).Width(300)
					ui.Row(ctx).Gap(8).Children(func() {
						if ui.Button(ctx, "Choose image…").Width(130).Clicked() {
							cb.OnPickBackground()
						}
						if s.Background != "" {
							if ui.Button(ctx, "Remove").Width(80).Clicked() {
								cb.OnBackground("", s.BackgroundDim)
							}
						}
					})
					ui.Row(ctx).Gap(10).Children(func() {
						ui.Text(ctx, "Dim").FontSize(13).Width(70).TextColor(muted(ctx))
						ui.Slider(ctx, &dim, 0, 1).Step(0.05).Width(160)
						ui.Text(ctx, percent(float32(dim))).FontSize(13).Width(44)
						if float32(dim) != s.BackgroundDim {
							cb.OnBackground(s.Background, float32(dim))
						}
					})
				})

				section(ctx, "Window opacity", func() {
					ui.Row(ctx).Gap(10).Children(func() {
						ui.Slider(ctx, &opacity, 0.5, 1).Step(0.02).Width(230)
						ui.Text(ctx, percent(float32(opacity))).FontSize(13).Width(44)
						if float32(opacity) != s.WindowOpacity {
							cb.OnOpacity(float32(opacity))
						}
					})
				})
			})
		})
}

// section is a titled group of controls.
func section(ctx *ui.Context, title string, body func()) {
	ui.Column(ctx).Gap(8).Children(func() {
		ui.Text(ctx, title).FontSize(13).FontWeight(600).TextColor(muted(ctx))
		body()
	})
}

// themeButton is one segment of the theme choice; the current one is filled.
func themeButton(ctx *ui.Context, label string, t Theme, cur Theme, onPick func(Theme)) ui.Element {
	b := ui.Button(ctx, label).Width(94)
	if cur == t {
		b.Background(accentSoft(ctx))
	}
	if b.Clicked() {
		onPick(t)
	}
	return b
}

func percent(v float32) string {
	return fmt.Sprintf("%d%%", int(v*100+0.5))
}

func muted(ctx *ui.Context) ui.Color {
	if ctx.Theme() != nil {
		return ctx.Theme().TextMuted
	}
	return ui.RGB(102, 113, 122)
}

func border(ctx *ui.Context) ui.Color {
	if ctx.Theme() != nil {
		return ctx.Theme().Border
	}
	return ui.RGB(220, 226, 222)
}

func accentSoft(ctx *ui.Context) ui.Color {
	if ctx.Theme() != nil && ctx.Theme().Background.R < 128 {
		return ui.RGB(53, 70, 61)
	}
	return ui.RGB(220, 233, 224)
}

func panelBg(ctx *ui.Context) ui.Color {
	if ctx.Theme() != nil {
		return ctx.Theme().Surface
	}
	return ui.RGB(255, 255, 255)
}

func shortPath(p string) string {
	if p == "" {
		return "No background image"
	}
	return p
}
