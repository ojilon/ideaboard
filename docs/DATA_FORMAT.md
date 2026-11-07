# IdeaBoard storage format

One board = one folder:

```
<boards dir>/<board name>/
  board.json
  assets/
    <image id>.png   (or .jpg/.webp — whatever the user pasted)
```

## board.json

Top-level object:

```json
{
  "version": 1,
  "name": "My board",
  "view": {
    "panX": 120.5,
    "panY": -40.0,
    "zoom": 1.25
  },
  "items": [
    { ... note ... },
    { ... region ... },
    { ... connector ... },
    { ... stroke ... },
    { ... image ... }
  ]
}
```

## Common item fields

Every item has:

| field | type | meaning |
|-------|------|---------|
| id | string | unique id (random at creation) |
| kind | string | "note", "region", "connector", "stroke", "image" |
| x, y | number | top-left position in board coordinates (DIPs) |
| w, h | number | size (notes, regions, images); for connectors/strokes it is the bounds |
| color | string | hex "#rrggbb" or "#rrggbbaa" (note fill, region label color, etc.) |
| z | number | draw order; larger draws on top |
| locked | boolean | true to prevent moving/resizing (optional) |

## Notes

```json
{
  "id": "n1",
  "kind": "note",
  "x": 100, "y": 60, "w": 200, "h": 120,
  "color": "#fff59d",
  "z": 0,
  "text": "Buy milk",
  "fontSize": 16
}
```

- `text` is plain text (one paragraph). The underline highlight style is a
  rendering choice, not stored per word.
- `fontSize` is the note's text size in DIPs; defaults to the setting.

## Regions (dotted-outline rectangles with a label)

```json
{
  "id": "r1",
  "kind": "region",
  "x": 20, "y": 20, "w": 400, "h": 300,
  "color": "#9e9e9e",
  "z": 0,
  "label": "Mon",
  "labelColor": "#616161"
}
```

- Drawn as a dotted rectangle with the label in a corner.
- Notes "dropped inside" belong visually to the region (no ownership stored;
  membership is spatial at draw time).

## Connectors (straight dotted lines between two notes)

```json
{
  "id": "c1",
  "kind": "connector",
  "x": 100, "y": 120, "w": 200, "h": 0,
  "color": "#9e9e9e",
  "z": 0,
  "from": "n1",
  "fromAnchor": "right",
  "to": "n2",
  "toAnchor": "left",
  "dash": [4, 4]
}
```

- `from`/`to` reference note ids; the line is redrawn between the notes'
  anchor points each frame so it follows when notes move.
- `w`/`h` are the connector's bounding box (recomputed, may be stored for
  hit testing).

## Strokes (freehand drawing)

```json
{
  "id": "s1",
  "kind": "stroke",
  "x": 0, "y": 0, "w": 300, "h": 200,
  "color": "#000000",
  "z": 0,
  "points": [[10, 10], [20, 12], [35, 9], ...],
  "width": 2,
  "tool": "pen"   // "pen" or "eraser"
}
```

- `points` is a list of [x, y] board coordinates.
- `tool` "eraser" strokes are drawn by cutting holes (or stored separately
  and composited); simplest: keep them as strokes drawn with the background
  color, or a separate erase mask. Undo/redo removes the last stroke of the
  current freehand session.

## Images

```json
{
  "id": "i1",
  "kind": "image",
  "x": 50, "y": 50, "w": 320, "h": 240,
  "color": "#ffffff",
  "z": 0,
  "asset": "img_abc123.png",
  "fit": "contain"
}
```

- `asset` names a file in `assets/` (copied in at attach/paste time).
- Movable and resizable like a note.

## Versioning

`version` is 1. The format is intentionally simple, flat JSON so a future
Android version can read the same files with any JSON parser. Fields apps
don't understand are ignored; unknown items are skipped.

## Safe writes

Every change is written to `board.json.tmp` in the same folder, then renamed
over `board.json`, so a crash never leaves a half-written board.json.

## Autosave

Changes are debounced (~1 second) and written automatically; a final write
happens when the board closes or the app quits.
