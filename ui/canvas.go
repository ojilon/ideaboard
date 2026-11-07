// Package ui builds the IdeaBoard window from the board model, using MyGo's
// native UI toolkit. This file holds the canvas state and its frame: an
// infinite pan/zoom whiteboard that fills the window.
package ui

import (
	"ideaboard/board"

	"github.com/egoist/mygo/ui"
)

// change is one undoable edit: an added item (undo removes it) or a removed
// item (undo restores it).
type change struct {
	add bool
	it  *board.Item
}

// Canvas is a full-window pan/zoom board. It holds pan/zoom state, the
// active tool and the gesture in progress, and turns pointer and key input
// into board edits (see canvas_input.go and canvas_items.go).
type Canvas struct {
	board *board.Board

	// Viewport: board coords of the window's top-left, zoom, window size.
	panX, panY   float32
	zoom         float32
	viewW, viewH float32

	tool      tool
	spaceHeld bool

	// Panning.
	draggingP            bool
	panStartX, panStartY float32

	// Selection and in-place editing.
	selected   string  // id of the selected item ("" none)
	editID     string  // item being edited in place ("" none)
	editRegion bool    // true when editing a region label, not note text
	editText   string  // working text for the item being edited
	dragID     string  // item being moved ("" none)
	dragOffX   float32 // pointer offset into the moved item, board coords
	dragOffY   float32
	resizeID   string // item being resized ("" none)
	resizeW    float32
	resizeH    float32
	resizeBX   float32 // pointer board position when the resize began
	resizeBY   float32

	// Color picker for a note, opened with a right click.
	colorPickID string
	colorPickX  float32 // window coords
	colorPickY  float32

	// Connector tool: the first note picked, waiting for the second.
	connFrom string
	// Region tool: the region being drawn, and where the drag started.
	draft          *board.Item
	draftX, draftY float32

	// Pen and eraser.
	penActive bool
	penPoints [][2]float32 // board coords of the stroke being drawn
	erasing   bool

	// Undo and redo for adds and removals (moves and resizes are not undone).
	history []change
	redo    []change

	// Images: decoded bitmaps by asset id. A nil value means the load failed
	// and is not retried every frame.
	imageCache map[string]ui.ImageSource
	loadAsset  func(asset string) ui.ImageSource

	// Optional background image behind the board, and its dim (0..1).
	bgImage ui.ImageSource
	bgDim   float32
	// noteFont is the family used for note text ("" for the default).
	noteFont string

	curTheme *ui.Theme

	// Hooks set by the app.
	onChange         func()
	onAttachImage    func()
	onPaste          func()
	onSettingsToggle func()
	onBoards         func()
	onQuit           func()
	showSettings     bool
}

// NewCanvas makes a canvas showing b, at the view saved in the board.
func NewCanvas(b *board.Board) *Canvas {
	c := &Canvas{
		board:      b,
		panX:       b.View.PanX,
		panY:       b.View.PanY,
		zoom:       b.View.Zoom,
		imageCache: map[string]ui.ImageSource{},
	}
	if c.zoom <= 0 {
		c.zoom = 1
	}
	return c
}

// OnChange sets the change hook (used to trigger autosave).
func (c *Canvas) OnChange(fn func()) { c.onChange = fn }

// OnAttachImage sets the hook that opens the image picker.
func (c *Canvas) OnAttachImage(fn func()) { c.onAttachImage = fn }

// OnPaste sets the hook that pastes an image or image file from the clipboard.
func (c *Canvas) OnPaste(fn func()) { c.onPaste = fn }

// OnSettingsToggle sets the hook that opens or closes the settings panel.
func (c *Canvas) OnSettingsToggle(fn func()) { c.onSettingsToggle = fn }

// OnBoards sets the hook that returns to the board list.
func (c *Canvas) OnBoards(fn func()) { c.onBoards = fn }

// OnQuit sets the hook for Ctrl+Q.
func (c *Canvas) OnQuit(fn func()) { c.onQuit = fn }

// SetShowSettings tells the toolbar whether the settings panel is open.
func (c *Canvas) SetShowSettings(on bool) { c.showSettings = on }

// SetNoteFont sets the font family for note text.
func (c *Canvas) SetNoteFont(family string) { c.noteFont = family }

// SetAssetLoader sets how image assets are loaded from disk when first drawn.
func (c *Canvas) SetAssetLoader(fn func(asset string) ui.ImageSource) { c.loadAsset = fn }

// SetBackground sets the optional background image and its dim (0..1).
// A nil src removes the background.
func (c *Canvas) SetBackground(src ui.ImageSource, dim float32) {
	c.bgImage = src
	c.bgDim = dim
}

// ViewCenter returns the board point at the middle of the window.
func (c *Canvas) ViewCenter() (float32, float32) {
	return c.panX + c.viewW/2/c.zoom, c.panY + c.viewH/2/c.zoom
}

// changed marks the board dirty: it stores the view and calls the hook.
func (c *Canvas) changed() {
	c.board.View = board.View{PanX: c.panX, PanY: c.panY, Zoom: c.zoom}
	if c.onChange != nil {
		c.onChange()
	}
}

// View builds the canvas element for one frame.
func (c *Canvas) View(ctx *ui.Context) ui.Element {
	c.curTheme = ctx.Theme()
	c.viewW, c.viewH = ctx.Size()
	box := ui.Box(ctx).Fill().
		Background(canvasColor(c.curTheme)).
		Draw(func(p *ui.Painter, r ui.Rect) {
			c.drawBoard(p, r)
		}).
		HandleInput(func(ev ui.InputEvent) bool {
			return c.input(ctx, ev)
		})

	box.Children(func() {
		c.header(ctx)
		c.toolbar(ctx, c.viewW)
		if c.editID != "" {
			c.editOverlay(ctx)
		}
		if c.colorPickID != "" {
			if it := c.boardItem(c.colorPickID); itemKindNote(it) {
				c.colorPicker(ctx)
			}
		}
	})
	return box
}

// itemKindNote reports whether it is a note.
func itemKindNote(it *board.Item) bool { return it != nil && it.Kind == board.KindNote }
