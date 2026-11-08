# IdeaBoard

A quiet, minimal note-board app for capturing small scattered ideas quickly
while reading. Open the app, double-click, type, done.

## Build and run

Requires Go 1.27.3+.

MyGo (the GUI library) is used through a `replace` directive in `go.mod`
pointing to a local clone. The clone is expected at `D:\projects\goui`
; if it lives elsewhere, edit the `replace` line in `go.mod`.

```bat
cd D:\projects\ideaboard
go run .
```

### Build a release binary

```bat
cd D:\projects\ideaboard
go build -trimpath -o ideaboard.exe .
```

### Develop

The app opens on the last board (or the board list). Tools are in the floating
toolbar at the top:

- Select (V): click and drag notes, regions, images and connectors; drag empty
  space to pan; drag a bottom-right corner to resize a note or image.
- Note (N): click to add a note; you can type at once. Double-click empty
  space to do the same. Click away or press Escape to finish.
- Region (R): drag to draw a labeled region; double-click its label to rename.
- Connector (C): click two notes to link them.
- Pen (P) / Eraser (E): freehand drawing and erasing.
- Image (I) or Ctrl+V: add an image (PNG or JPEG).
- Ctrl+Z / Ctrl+Y: undo and redo adds and removals. Ctrl+,: settings.
  Ctrl+Q: quit.

## Project layout

- `main.go` — app entry point (MyGo native UI window)
- `board/` — board model, storage, items (notes, regions, connectors, strokes, images)
- `ui/` — canvas view, pan/zoom, note/region/connector/image drawing, tools
- `settings/` — theme, font, background image settings
- `storage/` — board.json reading/writing, safe (temp + rename) writes
- `installer/` — Inno Setup script + build instructions (added last)
- `GOAL.md` — project goal
- `docs/` — architecture, data format, roadmap, MyGo notes

## Libraries

- [MyGo](https://github.com/egoist/mygo) (local fork at D:\projects\goui) —
  native UI, windows, transparency, input, fonts, images. No cgo, single
  Go binary of a few MB.

## License

MIT (MyGo is MIT; this app is yours to shape).
