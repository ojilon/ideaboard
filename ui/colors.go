// Package ui holds design-system color tokens for IdeaBoard, aligned with
// ideaboard-design-system.md. The light/dark palettes follow MyGo's
// theme.Theme (light/dark) so the board respects the chosen appearance.
package ui

import (
	"github.com/egoist/mygo/ui"
)

// Light canvas / note colors (design system 2.1, 2.2).
var light = palette{
	canvas:     ui.Hex("#F7F8F6"),
	surface:    ui.Hex("#FFFFFF"),
	inkPrimary: ui.Hex("#27313A"),
	inkSecond:  ui.Hex("#66717A"),
	line:       ui.Hex("#DCE2DE"),
	lineStrong: ui.Hex("#BCC7C0"),
	regionOut:  ui.Hex("#B8C6BD"),
	accent:     ui.Hex("#789A89"),
}

// Dark canvas / note colors (design system 2.3).
var dark = palette{
	canvas:     ui.Hex("#202127"),
	surface:    ui.Hex("#292B33"),
	inkPrimary: ui.Hex("#ECEEF0"),
	inkSecond:  ui.Hex("#B4BAC2"),
	line:       ui.Hex("#41454E"),
	lineStrong: ui.Hex("#5B626D"),
	regionOut:  ui.Hex("#555B64"),
	accent:     ui.Hex("#9EBEAA"),
}

type palette struct {
	canvas     ui.Color
	surface    ui.Color
	inkPrimary ui.Color
	inkSecond  ui.Color
	line       ui.Color
	lineStrong ui.Color
	regionOut  ui.Color
	accent     ui.Color
}

// notePalette maps a note's color id to its fill color.
var notePalette = map[string]ui.Color{
	"#FFF4C7": ui.Hex("#FFF4C7"), // note-yellow
	"#E2EEE3": ui.Hex("#E2EEE3"), // note-sage
	"#F8E5DE": ui.Hex("#F8E5DE"), // note-peach
	"#EEE8F8": ui.Hex("#EEE8F8"), // note-lavender
	"#E5EDF8": ui.Hex("#E5EDF8"), // note-blue
	"#F1F2EF": ui.Hex("#F1F2EF"), // note-neutral
}

// paletteFor returns the design palette matching the theme's appearance.
func paletteFor(th *ui.Theme) palette {
	if th == nil {
		return light
	}
	// Luminance of the theme background decides light vs dark.
	r, g, b := float64(th.Background.R), float64(th.Background.G), float64(th.Background.B)
	lum := 0.299*r + 0.587*g + 0.114*b
	if lum > 127 {
		return light
	}
	return dark
}

// canvasColor returns the board background for the theme.
func canvasColor(th *ui.Theme) ui.Color {
	return paletteFor(th).canvas
}

// noteFill returns the fill color for a note's color id, defaulting to
// note-yellow.
func noteFill(id string) ui.Color {
	if c, ok := notePalette[id]; ok {
		return c
	}
	return ui.Hex("#FFF4C7")
}

// strokeColor returns a mid-tone border for a note.
func strokeColor(id string) ui.Color {
	// a slightly darker variant of the fill
	return ui.Hex("#E8D991")
}

// regionColor returns the region outline color for the theme.
func regionColor(th *ui.Theme) ui.Color {
	return paletteFor(th).regionOut
}

// connectorColor returns the connector color for the theme.
func connectorColor(th *ui.Theme) ui.Color {
	return paletteFor(th).lineStrong
}

// inkColor returns primary text color for the theme.
func inkColor(th *ui.Theme) ui.Color {
	return paletteFor(th).inkPrimary
}

// mutedColor returns secondary text color for the theme.
func mutedColor(th *ui.Theme) ui.Color {
	return paletteFor(th).inkSecond
}
