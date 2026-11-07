// Package settings renders the settings panel (theme, font, background).
package settings

import (
	"github.com/egoist/mygo/ui"
)

// Callbacks the panel reports to the shell.
type Callbacks struct {
	OnTheme      func(Theme)
	OnFontFamily func(string)
	OnFontFile   func(string)
	OnBackground func(path string, dim float32)
	OnClose      func()
}

// Panel builds a floating settings panel from s, reporting changes via cb.
func Panel(ctx *ui.Context, s Settings, cb Callbacks) ui.Element {
	fontFamily := s.FontFamily
	dim := float64(s.BackgroundDim)
	w, _ := ctx.Size()

	return ui.Box(ctx).Absolute().
		Left((w - 360) / 2).Top(100).Size(360, 440).
		Background(panelBg(ctx)).Radius(10).
		Children(func() {
			ui.Column(ctx).Gap(12).Padding(18).Children(func() {
				ui.Text(ctx, "Settings").FontSize(20).Bold()
				ui.Text(ctx, "Appearance").FontSize(13).TextColor(muted(ctx)).FontWeight(600)
				ui.Row(ctx).Gap(8).Children(func() {
					themeButton(ctx, "System", ThemeSystem, s.Theme, cb.OnTheme)
					themeButton(ctx, "Light", ThemeLight, s.Theme, cb.OnTheme)
					themeButton(ctx, "Dark", ThemeDark, s.Theme, cb.OnTheme)
				})

				ui.Text(ctx, "Note font").FontSize(13).TextColor(muted(ctx)).FontWeight(600)
				ui.Row(ctx).Gap(8).Children(func() {
					ui.TextInput(ctx, &fontFamily).Placeholder("Font family").Width(210).
						OnChange(func() { cb.OnFontFamily(fontFamily) })
					if ui.Button(ctx, "Choose .ttf…").Width(100).Clicked() {
						cb.OnFontFile("")
					}
				})

				ui.Text(ctx, "Background image").FontSize(13).TextColor(muted(ctx)).FontWeight(600)
				ui.Row(ctx).Gap(8).Children(func() {
					if ui.Button(ctx, "Choose image…").Width(110).Clicked() {
						cb.OnBackground(s.Background, s.BackgroundDim)
					}
					ui.Text(ctx, shortPath(s.Background)).FontSize(11).TextColor(muted(ctx)).Width(170)
				})
				ui.Row(ctx).Gap(8).Children(func() {
					ui.Text(ctx, "Dim").FontSize(12).Width(30)
					ui.Slider(ctx, &dim, 0, 1).Step(0.05).Width(200)
					if float32(dim) != s.BackgroundDim {
						cb.OnBackground(s.Background, float32(dim))
					}
				})

				ui.Box(ctx).Grow(1)
				ui.Row(ctx).Children(func() {
					if ui.Button(ctx, "Close").Width(70).Clicked() {
						cb.OnClose()
					}
				})
			})
		})
}

func themeButton(ctx *ui.Context, label string, t Theme, cur Theme, onPick func(Theme)) ui.Element {
	b := ui.Button(ctx, label).Width(80)
	if cur == t {
		b.Background(ctx.Theme().Accent)
	}
	if b.Clicked() {
		onPick(t)
	}
	return b
}

func muted(ctx *ui.Context) ui.Color {
	if ctx.Theme() != nil {
		return ctx.Theme().TextMuted
	}
	return ui.RGB(102, 113, 122)
}

func panelBg(ctx *ui.Context) ui.Color {
	if ctx.Theme() != nil {
		return ctx.Theme().Surface
	}
	return ui.RGB(255, 255, 255)
}

func shortPath(p string) string {
	if p == "" {
		return "(none)"
	}
	return p
}
