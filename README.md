# IdeaBoard

A quiet, minimal note-board app for capturing small scattered ideas quickly
while reading. Open the app, double-click, type, done.

## Build and run

Requires Go 1.27.3+.

MyGo (the GUI library) is used through a `replace` directive in `go.mod`
pointing to a local clone. The clone is expected at `D:\projects\goui`
(adjust `IDEABOARD_MYGO` if it lives elsewhere).

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

The app opens a frameless, transparent window on the last board. Double-click
empty space to create a note; drag notes to move them; resize by the bottom
right corner; double-click a note's text to edit it in place.

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
