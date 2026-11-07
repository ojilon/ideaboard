# IdeaBoard architecture

## Folder structure

```
ideaboard/
  main.go              # entry point: MyGo app, window, board list, open board, settings, image pick/paste
  go.mod               # module + replace directive -> local MyGo clone
  board/               # board domain model
    item.go            # Board, Layer, Item (note, region, connector, stroke, image), add/remove/restore
  storage/             # persistence
    store.go           # board folder, board.json + assets/, safe write, autosave debounce, board list
  ui/                  # native UI (MyGo `ui` package)
    canvas.go          # Canvas state, View (one frame), public hooks
    canvas_input.go    # pointer/keyboard input -> tool actions, pan, zoom
    canvas_items.go    # hit-testing, create/remove with undo history, edit mode, connector geometry
    canvas_draw.go     # drawing: notes, regions, connectors, strokes, images, background
    tools.go           # the tool enum (select, note, region, connector, pen, eraser, image)
    toolbar.go         # header (identity, board name) and the floating toolbar
    icons.go           # SVG icons for the toolbar and panels
    boardlist.go       # board list screen: cards, New board
    colors.go          # design-system tokens (light/dark, note palette)
  settings/            # user settings
    settings.go        # settings.json: theme, font, background, dim, window opacity, last board
    ui.go              # settings panel (right side)
  installer/           # Inno Setup script + build instructions
  GOAL.md
  README.md
  docs/
```

## Modules and responsibilities

- **main.go** wires everything: creates the MyGo window (`mygo.NewWindow` with
  native UI `ui.View`), owns the board list and settings, and runs the app
  loop (`mygo.App.Run`).
- **board package** holds the in-memory model. A `Board` has a list of
  `Layer`s; each layer holds `Item`s. Items know their kind (note, region,
  connector, stroke, image), position, size, color, and text. The board
  exposes `AddItem`, `RemoveItem`, `ItemAt`, `MoveItem`, `ResizeItem`, and
  hit-testing for double-click create and drag select.
- **storage package** persists one board to one folder:
  `<boardsdir>/<boardname>/board.json` + `<boardsdir>/<boardname>/assets/`.
  Writes are safe: serialize to a temp file in the same folder, then rename.
  Autosave runs on every change, debounced (~1 s), plus a save on window
  close/quit.
- **ui package** builds the MyGo native UI view each frame from the board
  model. The active tool decides what a click does (see `ui/tools.go`). One full-window element paints the canvas with a pan/zoom
  `ui.Transform`. Pointer input is read each frame (`PointerPosition`,
  `PressedButton`, `InputScroll`, `HandleInput`) to pan, zoom, drag notes,
  and double-click to create.
- **settings package** stores `settings.json` in the app data folder: theme
  (light/dark/system), note font family, background image path + dim, window
  opacity, window position/size, and the last board opened.

## Data flow

1. `main.go` loads settings + board list at startup, opens the last board.
2. The board model lives in memory; every user action mutates it and flags it
   dirty.
3. A debounced timer writes `board.json` (safe write) after the last change.
4. Each frame, `ui.View(app.view)` walks the model and paints the canvas.
5. Pointer/keyboard input mutates the model directly (no separate event bus;
   keep it small).
6. Settings changes re-theme the view and invalidate the frame.

## Constraints honored

- Single Go binary, no cgo, `CGO_ENABLED=0`.
- No constant redraws: frames only on input, invalidation, or animation.
- Low memory: canvases/bitmaps released when a board closes; images decoded
  once and kept as `ui.Bitmap`.
- Plain Go + standard library wherever possible; MyGo for windowing/UI.
