You are building a new desktop app from scratch. Work in small, verifiable steps. Do not rewrite files wholesale once they exist; make surgical edits.

## Setup (do this first)
1. Create a new folder: [D:\projects\ideaboard] and run `git init` in it.
2. The GUI library is MyGo (a Go library for building PC apps).
   - My fork: [fork URL]
   - Local clone: [local path to MyGo clone]
   Read the library's README, examples and source BEFORE writing UI code. Use it via a `replace` directive in go.mod pointing to the local clone. If MyGo cannot do something I ask for, tell me clearly and propose the closest workable alternative. Do not silently fake it.
3. Write these files before any app code:
   - `GOAL.md`: the project goal in plain words (below)
   - `README.md`: what it is, how to build and run
   - `docs/ARCHITECTURE.md`: folder structure, modules, data flow
   - `docs/DATA_FORMAT.md`: the storage format
   - `docs/ROADMAP.md`: milestones, with Android as a later phase
   - `docs/MYGO_NOTES.md`: what MyGo supports and its limits, as you discover them

## Goal
A quiet, minimal note-board app for capturing small scattered ideas quickly while I am reading. I lose ideas because I don't write them down. Capturing must take under 3 seconds: open the app, double-click, type, done.

## Look and feel
- Clean, calm, minimal. Clear/transparent-feeling background with slight window transparency if MyGo supports it; support an optional background image (user-chosen, with dim/blur control).
- Minimal rounded corners (small radius, 4-6px) on windows, notes and buttons.
- A pleasant readable font for notes. Make font choice a setting.
- Text highlight style: a short underline (a thin line under isolated words) instead of a heavy background highlight.
- Light and dark theme, restrained palette. No clutter, no heavy toolbars. Prefer a slim floating toolbar.

## Core feature: the board (central screen)
An infinite or very large pan/zoom canvas, like a whiteboard.
1. Sticky notes: create (double-click empty space), drag, resize, edit text in place, choose a few soft colors, delete.
2. Regions: dotted-outline rectangles with a label, to mark areas such as days ("Mon", "Tue"). Notes dropped inside belong visually to the region.
3. Connectors: straight dotted lines between two notes that follow when the notes move.
4. Freehand drawing: a pen tool for hand-drawn lines on the board, with an eraser and undo/redo.
5. Images: attach/paste images onto the board as movable, resizable items.
6. Pan with middle-mouse/space+drag, zoom with ctrl+scroll.

## Storage
- Fully local, no network, no accounts.
- One board = one folder: `board.json` (items, positions, connectors, strokes) + `assets/` for images.
- Autosave on every change (debounced) and safe writes (write temp file, then rename) so nothing is lost on crash.
- Keep the format simple and documented so an Android version can read the same files later.
- Multiple boards, with a simple board list.

## Installer
- A small Windows installer that lets me choose the install directory, creates a Start Menu shortcut and an uninstaller. Use the lightest option (e.g. Inno Setup script or NSIS). Put it in `installer/` with build instructions. Do this AFTER the app runs.

## Build order (stop and show me after each)
1. Setup + docs
2. MyGo window with background and transparency, hello world
3. Canvas with pan/zoom
4. Sticky notes + storage
5. Regions + connectors
6. Freehand + images
7. Settings (theme, font, background)
8. Installer

## Rules
- My machine is a low-end Windows PC (Celeron). Keep it light: no heavy dependencies, low memory, no constant redraws.
- Keep code small and readable. Prefer plain Go and the standard library.
- Commit after each step with a clear message.
- If something fails or you are unsure, say so. Do not claim a feature works until you ran it.