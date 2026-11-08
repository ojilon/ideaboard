package ui

// tool is the active canvas tool. It decides what a click or drag does.
type tool int

const (
	toolSelect    tool = iota // select, move, resize; drag empty space to pan
	toolNote                  // click to add a note and type in it
	toolRegion                // drag to draw a labeled region
	toolConnector             // click two notes to connect them
	toolPen                   // freehand drawing
	toolEraser                // erase freehand strokes
	toolImage                 // open the image picker
)

// toolName is the accessible name of each tool, used as the tooltip.
var toolName = map[tool]string{
	toolSelect:    "Select (V)",
	toolNote:      "Note (N)",
	toolRegion:    "Region (R)",
	toolConnector: "Connector (C)",
	toolPen:       "Pen (P)",
	toolEraser:    "Eraser (E)",
	toolImage:     "Image (I)",
}
