# IdeaBoard - project goal

A quiet, minimal note-board app for capturing small scattered ideas quickly
while I am reading. I lose ideas because I don't write them down. Capturing
must take under 3 seconds: open the app, double-click, type, done.

## What it is

A local, single-user desktop app (Windows first) with one infinite pan/zoom
canvas ("the board"). On the board you place sticky notes, group them into
labeled regions, connect notes with dotted lines, draw freehand, and paste
images. Everything is stored on disk in one folder per board.

## Core capture flow (must stay under 3 seconds)

1. Open the app (it opens on the last board, window-sized, no login).
2. Double-click empty board space.
3. A note appears with the caret inside; type the idea.
4. Done — the note autosaves. Click away or press Escape to finish.

## Guiding principles

- Quiet and calm: clean, minimal, transparent-feeling background, slim
  floating toolbar, no clutter, no heavy toolbars.
- Light: low memory, low CPU, no constant redraws. Runs on a Celeron.
- Local only: no network, no accounts, no cloud.
- One board = one folder: `board.json` + `assets/`. Simple, documented format
  so a later Android version can read the same files.
- Multiple boards with a simple board list.
- Light/dark theme, user-chosen font, optional background image with dim/blur.

## Out of scope for now

- Accounts, sync, collaboration, network of any kind.
- Mobile (Android is a later phase, see docs/ROADMAP.md).
- Rich text beyond a note's plain text and an optional underline highlight.
