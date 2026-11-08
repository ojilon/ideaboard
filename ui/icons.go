package ui

import "github.com/egoist/mygo/ui"

// svgIcon wraps a 24x24 stroke icon. Icons are drawn in the color the
// caller passes, so one shape serves every theme.
func svgIcon(body string) *ui.SVG {
	doc := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#000" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">` + body + `</svg>`
	return ui.MustParseSVG([]byte(doc))
}

var (
	iconSelect    = svgIcon(`<path d="M5 3.5l13.5 7.5-6 1.6-2.2 6z"/>`)
	iconNote      = svgIcon(`<path d="M5 4h14v10.5L13.5 20H5z"/><path d="M13.5 20v-5.5H19"/>`)
	iconRegion    = svgIcon(`<rect x="4" y="4" width="16" height="16" rx="2" stroke-dasharray="2.5 2.5"/>`)
	iconConnector = svgIcon(`<circle cx="6" cy="6" r="2"/><circle cx="18" cy="18" r="2"/><path d="M8 8l8 8" stroke-dasharray="0.1 3.2"/>`)
	iconPen       = svgIcon(`<path d="M4 20l1-4L16 5l3 3L8 19z"/><path d="M14 7l3 3"/>`)
	iconEraser    = svgIcon(`<path d="M4 16l8-10 6 6-6 6H7z"/><path d="M10 20h10"/>`)
	iconImage     = svgIcon(`<rect x="3.5" y="4.5" width="17" height="15" rx="2"/><circle cx="9" cy="10" r="1.6"/><path d="M4 18l5-5 4 4 2-2 5 5"/>`)
	iconBoards    = svgIcon(`<rect x="4" y="4" width="6.5" height="6.5" rx="1.5"/><rect x="13.5" y="4" width="6.5" height="6.5" rx="1.5"/><rect x="4" y="13.5" width="6.5" height="6.5" rx="1.5"/><rect x="13.5" y="13.5" width="6.5" height="6.5" rx="1.5"/>`)
	iconSettings  = svgIcon(`<path d="M4 7h9M17 7h3M4 17h3M11 17h9"/><circle cx="15" cy="7" r="2"/><circle cx="9" cy="17" r="2"/>`)
	iconPlus      = svgIcon(`<path d="M12 5v14M5 12h14"/>`)
	iconBack      = svgIcon(`<path d="M15 5l-7 7 7 7"/>`)
)
