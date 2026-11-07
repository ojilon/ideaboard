// Design-system color tokens for IdeaBoard (ideaboard-design-system.md,
// sections 2.1-2.3). Light and dark palettes are chosen from the MyGo
// theme's background, so the board follows the chosen appearance.
package ui

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// palette holds the UI tokens for one appearance.
type palette struct {
	canvas      ui.Color
	surface     ui.Color
	surfaceMute ui.Color
	surfaceHov  ui.Color
	inkPrimary  ui.Color
	inkSecond   ui.Color
	inkTertiary ui.Color
	line        ui.Color
	lineStrong  ui.Color
	regionOut   ui.Color
	accent      ui.Color
	accentSoft  ui.Color
	focus       ui.Color
	danger      ui.Color
}

// light is design system 2.1.
var light = palette{
	canvas:      mustHex("#F7F8F6"),
	surface:     mustHex("#FFFFFF"),
	surfaceMute: mustHex("#F2F4F1"),
	surfaceHov:  mustHex("#EEF2EE"),
	inkPrimary:  mustHex("#27313A"),
	inkSecond:   mustHex("#66717A"),
	inkTertiary: mustHex("#89939A"),
	line:        mustHex("#DCE2DE"),
	lineStrong:  mustHex("#BCC7C0"),
	regionOut:   mustHex("#B8C6BD"),
	accent:      mustHex("#789A89"),
	accentSoft:  mustHex("#DCE9E0"),
	focus:       mustHex("#507D68"),
	danger:      mustHex("#B85C55"),
}

// dark is design system 2.3.
var dark = palette{
	canvas:      mustHex("#202127"),
	surface:     mustHex("#292B33"),
	surfaceMute: mustHex("#30333C"),
	surfaceHov:  mustHex("#393D47"),
	inkPrimary:  mustHex("#ECEEF0"),
	inkSecond:   mustHex("#B4BAC2"),
	inkTertiary: mustHex("#8C949E"),
	line:        mustHex("#41454E"),
	lineStrong:  mustHex("#5B626D"),
	regionOut:   mustHex("#555B64"),
	accent:      mustHex("#9EBEAA"),
	accentSoft:  mustHex("#35463D"),
	focus:       mustHex("#B1D0BC"),
	danger:      mustHex("#E18B82"),
}

// noteStyle is one entry of the note palette (design system 2.2): the fill
// and the border/accent that goes with it.
type noteStyle struct {
	name   string
	fill   string
	border string
}

// noteStyles is the note palette, in the order the color picker shows it.
var noteStyles = []noteStyle{
	{"yellow", "#FFF4C7", "#E8D991"},
	{"sage", "#E2EEE3", "#B9D0BD"},
	{"peach", "#F8E5DE", "#E7C5B8"},
	{"lavender", "#EEE8F8", "#D0C3E8"},
	{"blue", "#E5EDF8", "#C4D2E8"},
	{"neutral", "#F1F2EF", "#D9DEDA"},
}

// regionLabels are the label colors offered for regions (muted accents,
// dark enough for text on the canvas).
var regionLabels = []string{"#789A89", "#B07F69", "#8C7FB0", "#6E8FB8"}

// noteInk is the note text color. Notes keep light-theme ink in both
// themes, because pastel fills need dark text (design system 2.3).
var noteInk = mustHex("#27313A")

// noteFillDefault is the fill of a note with no or an unknown color.
const noteFillDefault = "#FFF4C7"

// paletteFor returns the palette matching the theme's appearance.
func paletteFor(th *ui.Theme) palette {
	if th == nil {
		return light
	}
	if luminance(th.Background) > 127 {
		return light
	}
	return dark
}

// canvasColor returns the board background for the theme.
func canvasColor(th *ui.Theme) ui.Color { return paletteFor(th).canvas }

// inkColor returns the primary UI text/ink color for the theme.
func inkColor(th *ui.Theme) ui.Color { return paletteFor(th).inkPrimary }

// mutedColor returns the secondary text color for the theme.
func mutedColor(th *ui.Theme) ui.Color { return paletteFor(th).inkSecond }

// regionOutline returns the dashed region outline color for the theme.
func regionOutline(th *ui.Theme) ui.Color { return paletteFor(th).regionOut }

// connectorColor returns the connector color for the theme.
func connectorColor(th *ui.Theme) ui.Color { return paletteFor(th).inkSecond }

// noteColorID normalizes a note color id for lookups.
func noteColorID(id string) string { return strings.ToUpper(id) }

// noteFill returns the fill for a note's color id. Unknown valid hex colors
// (from older files) are used as they are; anything else gets the default.
func noteFill(id string) ui.Color {
	for _, s := range noteStyles {
		if noteColorID(s.fill) == noteColorID(id) {
			return mustHex(s.fill)
		}
	}
	if c, ok := parseColor(id); ok {
		return c
	}
	return mustHex(noteFillDefault)
}

// noteBorder returns the border color for a note's color id.
func noteBorder(id string) ui.Color {
	for _, s := range noteStyles {
		if noteColorID(s.fill) == noteColorID(id) {
			return mustHex(s.border)
		}
	}
	return mustHex("#D9DEDA")
}

// regionLabel returns the label color for a region item's color id, or the
// theme's muted ink when it has none.
func regionLabel(id string, th *ui.Theme) ui.Color {
	if c, ok := parseColor(id); ok {
		return c
	}
	return mutedColor(th)
}

// parseColor reads "#rrggbb". It reports false for anything else.
func parseColor(s string) (ui.Color, bool) {
	if len(s) != 7 || s[0] != '#' {
		return ui.Color{}, false
	}
	v, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return ui.Color{}, false
	}
	return ui.RGB(uint8(v>>16), uint8(v>>8), uint8(v)), true
}

// mustHex parses a constant design token. Only compile-time constants go
// through it, so a bad token is a programming error.
func mustHex(s string) ui.Color {
	c, ok := parseColor(s)
	if !ok {
		panic("ui: bad color token " + s)
	}
	return c
}

// luminance returns the perceived brightness of a color, 0..255.
func luminance(c ui.Color) float64 {
	return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
}
