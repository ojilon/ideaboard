# IdeaBoard roadmap

Android is a later phase; everything before it is Windows desktop.

## Milestone 1 — Setup + docs (this session)

- `D:\projects\ideaboard` created, `git init`
- `GOAL.md`, `README.md`, `docs/ARCHITECTURE.md`, `docs/DATA_FORMAT.md`,
  `docs/ROADMAP.md`, `docs/MYGO_NOTES.md` written
- `go.mod` with `replace` directive to the local MyGo clone (`D:\projects\goui`)
- Commit: "setup: project folder, docs, go.mod"

## Milestone 2 — MyGo window with background and transparency, hello world

- Frameless/transparent window (`WindowOptions.Transparent`, `Opacity`,
  `CornerRadius`), hello-world text
- Window opens fast, shows last board
- Commit: "ui: transparent window, hello world"

## Milestone 3 — Canvas with pan/zoom

- Infinite pan/zoom canvas (ui.Transform), pan with middle-mouse / space+drag,
  zoom with ctrl+scroll
- Commit: "canvas: pan and zoom"

## Milestone 4 — Sticky notes + storage

- Create note (double-click empty space), drag, resize, edit in place,
  soft colors, delete
- board.json + assets/, autosave debounced + safe write
- Multiple boards, simple board list
- Commit: "notes: sticky notes + storage"

## Milestone 5 — Regions + connectors

- Dotted-outline regions with labels (Mon, Tue, …)
- Dotted connector lines between notes that follow movement
- Commit: "board: regions and connectors"

## Milestone 6 — Freehand drawing + images

- Pen tool, eraser, undo/redo for strokes
- Attach/paste images as movable, resizable items
- Commit: "board: freehand and images"

## Milestone 7 — Settings

- Light / dark theme, note font (user-chosen), background image + dim/blur
- Commit: "settings: theme, font, background"

## Milestone 8 — Installer (after the app runs)

- Inno Setup script in `installer/`, choose install dir, Start Menu shortcut,
  uninstaller, build instructions
- Commit: "installer: Inno Setup script"

## Later phase — Android

- Same `board.json` + `assets/` format read by an Android client
- Touch pan/zoom, create note by tap-and-hold or double-tap
- Not part of the Windows desktop milestones above
