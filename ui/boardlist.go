package ui

import (
	"fmt"
	"math"
	"time"

	"ideaboard/storage"

	"github.com/egoist/mygo/ui"
)

// BoardList is the start screen: a grid of boards, with New board. Cards
// show a soft preview, the board name and when it was last edited.
type BoardList struct {
	Infos      []storage.BoardInfo
	OnOpen     func(name string)
	OnNew      func()
	OnSettings func()
	// Now is the clock for "Edited today" labels; nil uses time.Now.
	Now func() time.Time
}

// previewTints are the soft fills of the card previews (design system 2.1).
var previewTints = []string{"#DCE9E0", "#F8E5DE", "#EEE8F8", "#FFF4C7", "#E5EDF8"}

const (
	listPad   float32 = 28
	listGap   float32 = 20
	cardH     float32 = 196
	previewH  float32 = 108
	cardMaxW  float32 = 440
	titleSize float32 = 24
)

// View builds the board list for one frame.
func (l *BoardList) View(ctx *ui.Context) ui.Element {
	th := ctx.Theme()
	pal := paletteFor(th)
	w, _ := ctx.Size()

	cols := 3
	switch {
	case w < 560:
		cols = 1
	case w < 900:
		cols = 2
	}
	cardW := min((w-2*listPad-float32(cols-1)*listGap)/float32(cols), cardMaxW)

	box := ui.Box(ctx).Fill().Background(canvasColor(th))
	box.Children(func() {
		ui.Column(ctx).Padding(listPad).Gap(listGap).Children(func() {
			// Header: identity on the left, settings on the right. The header
			// is the window's drag handle, since the window is frameless.
			ui.Row(ctx).Height(36).Gap(10).DragWindow().Children(func() {
				ui.Box(ctx).Size(20, 20).Draw(func(p *ui.Painter, r ui.Rect) {
					drawLogo(p, r, pal.accent)
				})
				ui.Text(ctx, "Ideaboard").FontSize(15).FontWeight(600).TextColor(inkColor(th))
				ui.Box(ctx).Grow(1)
				iconButton(ctx, iconSettings, "Settings (Ctrl+,)", false, pal, l.OnSettings)
			})

			ui.Row(ctx).Gap(16).Children(func() {
				ui.Text(ctx, "Your boards").FontSize(titleSize).FontWeight(600).TextColor(inkColor(th))
				ui.Box(ctx).Grow(1)
				newBoardButton(ctx, pal, l.OnNew)
			})

			if len(l.Infos) == 0 {
				ui.Text(ctx, "No boards yet. Create one to start capturing ideas.").
					FontSize(14).TextColor(mutedColor(th))
				return
			}
			for i := 0; i < len(l.Infos); i += cols {
				ui.Row(ctx).Gap(listGap).Children(func() {
					for j := i; j < i+cols && j < len(l.Infos); j++ {
						l.card(ctx, l.Infos[j], cardW, pal)
					}
				})
			}
		})
	})
	return box
}

// newBoardButton is the one primary action on the list.
func newBoardButton(ctx *ui.Context, pal palette, onNew func()) {
	ui.Row(ctx).Height(40).Padding(0, 16).Gap(8).Radius(6).Background(pal.focus).
		Tooltip("Create a new board").
		Children(func() {
			ui.Icon(ctx, iconPlus)
			ui.Text(ctx, "New board").FontSize(14).FontWeight(500).TextColor(mustHex("#FFFFFF"))
		}).
		HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputPointerDown && ev.Button == 0 && onNew != nil {
				onNew()
				return true
			}
			return false
		})
}

// card is one board: preview, name and the edited date. The whole card
// opens the board on click.
func (l *BoardList) card(ctx *ui.Context, info storage.BoardInfo, w float32, pal palette) {
	name := info.Name
	ui.Box(ctx).Size(w, cardH).Background(pal.surface).Radius(6).Border(1, pal.line).
		HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputPointerDown && ev.Button == 0 && l.OnOpen != nil {
				l.OnOpen(name)
				return true
			}
			return false
		}).
		Children(func() {
			ui.Column(ctx).Padding(12).Gap(10).Children(func() {
				ui.Box(ctx).Size(w-26, previewH).Radius(4).
					Draw(func(p *ui.Painter, r ui.Rect) {
						drawPreview(p, r, name)
					})
				ui.Text(ctx, name).FontSize(15).FontWeight(500).TextColor(pal.inkPrimary)
				ui.Text(ctx, editedLabel(info.Modified, l.now())).FontSize(12).TextColor(pal.inkSecond)
			})
		})
}

func (l *BoardList) now() time.Time {
	if l.Now != nil {
		return l.Now()
	}
	return time.Now()
}

// drawPreview paints a quiet placeholder: a soft tint with two small notes.
// The tint and the note spots are chosen from the board name, so each board
// keeps its own look without loading its contents.
func drawPreview(p *ui.Painter, r ui.Rect, name string) {
	h := 0
	for _, ch := range name {
		h = h*31 + int(ch)
	}
	if h < 0 {
		h = -h
	}
	tint := mustHex(previewTints[h%len(previewTints)])
	p.Fill(r, tint, 4)
	n1 := mustHex(noteStyles[(h/7)%len(noteStyles)].fill)
	n2 := mustHex(noteStyles[(h/11+2)%len(noteStyles)].fill)
	p.Fill(ui.Rect{X: r.X + r.W*0.14, Y: r.Y + r.H*0.22, W: r.W * 0.30, H: r.H * 0.36}, n1, 4)
	p.Fill(ui.Rect{X: r.X + r.W*0.52, Y: r.Y + r.H*0.46, W: r.W * 0.30, H: r.H * 0.32}, n2, 4)
}

// editedLabel is the card caption: "Edited today", "Edited yesterday",
// "Edited 3 days ago", or the date for older boards.
func editedLabel(mod, now time.Time) string {
	days := int(math.Floor(startOfDay(now).Sub(startOfDay(mod)).Hours() / 24))
	switch {
	case days <= 0:
		return "Edited today"
	case days == 1:
		return "Edited yesterday"
	case days < 7:
		return fmt.Sprintf("Edited %d days ago", days)
	default:
		return "Edited " + mod.Format("2 Jan 2006")
	}
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
