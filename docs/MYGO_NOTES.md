# MyGo notes — what it supports and its limits

MyGo (local clone at D:\projects\goui, module github.com/egoist/mygo v0.8.3)
is a pure-Go desktop framework, no cgo. Each window shows either a web page in
the system webview or **native UI drawn by MyGo itself** (package `ui`, on the
GPU where possible). This app uses native UI: low memory, instant window open,
no HTML/JavaScript.

## What MyGo supports (confirmed by reading source + running the app)

- **Windows native UI:** `mygo.NewWindow` with `Content: ui.View(...)`, view
  runs on the main thread each frame. `Transparent` + `Opacity` + `Frameless`
  + `CornerRadius` all work (frameless transparent window opens on Windows).
- **2D drawing:** `ui.Painter.Fill`, `Stroke`, `StrokeDashed`, `Line`,
  `TextLayout` (via `ui.ShapeText`), `Image`, `Clip`, `Transform`
  (`ui.Translation`, `ui.Scaling`). Pan/zoom canvas confirmed working.
- **Input:** `Element.HandleInput` with `InputEvent` (pointer down/up/move,
  scroll with `Ctrl` mod = zoom, keys); `ctx.Invalidate()` from any goroutine.
  Confirmed: middle-drag/space+drag pan, ctrl+scroll zoom, double-click create,
  drag/resize notes, delete.
- **Text input:** `ui.TextArea(ctx, &value)` for in-place note editing, with
  `OnShortcut` (Escape commits). Confirmed working.
- **Quit hook:** `mygo.App.OnWillQuit` for save-on-quit. Confirmed working.
- **Images:** `ui.DecodeImage(data []byte) (ImageSource, error)` loads PNG/JPG
  into a `ui.Bitmap` (not yet wired into the app). `ui.RegisterFont` for
  custom fonts.
- **Themes:** `c.Theme()` light/dark, colors (Background, Surface, Border,
  Text, TextMuted, Accent).
- **Menus:** `ui.PopoverMenu` / `ui.Menu` available (not yet used).
- **No cgo:** builds with `CGO_ENABLED=0`; single Go binary of a few MB.
- **Installer:** MyGo has app-bundling for macOS/Linux, but no Windows
  installer baked in — we add an Inno Setup script in `installer/`.

## Limits / things MyGo cannot (or does not obviously) do

- **Window transparency on Windows:** `Transparent` + `Opacity` are supported
  in `WindowOptions`; verify they actually compose on Windows (Win32 layered
  window). If not, the fallback is an opaque window with a calm solid
  background color.
- **Blur behind the window / acrylic:** not exposed as a simple option; there
  is `Vibrancy` (macOS / Windows 11 material). Background-image dim/blur must
  be done by drawing a dimmed/blurred copy ourselves, or skipped. If we
  cannot blur the background image cheaply, we dim it (draw a translucent
  rect over it) instead of a real blur.
- **True "infinite" canvas:** MyGo paints a window-sized scene each frame.
  An "infinite" board is simulated: the canvas element fills the window and
  the view translates/scales board coordinates into window coordinates; items
  beyond the view are culled by not drawing them. Storage has no size limit.
- **Text underline highlight (short underline under isolated words):** MyGo
  paints text with `Painter.Text`; there is no built-in per-word underline
  style. Workaround: draw a thin line under a word ourselves at the text's
  baseline, or skip per-word highlighting and keep plain text. The underline
  is a rendering detail, not stored in board.json.
- **GIF animation / video:** `ui.gif.go` exists but is not a focus; images are
  static bitmaps.
- **Native file drag-and-drop of images onto the board:** `dragdrop.go`
  exists; confirm paste/attach works. If OS paste of an image is not
  directly readable, fall back to an "attach image file" dialog.
- **Constant redraws are discouraged:** frames happen on input/invalidation/
  animation only; freehand drawing while moving invalidates each frame, which
  is expected and fine.

## How we handle what MyGo can't do

| Want | MyGo reality | Workable alternative |
|------|--------------|----------------------|
| Transparent/blurred background | `Transparent`+`Opacity` supported; real blur not exposed | Transparent window + optional dimmed background image; fallback to solid calm background |
| Underline highlight | No per-word underline API | Draw thin line under word ourselves, or plain text only |
| Infinite canvas | Window-sized scene, transform maps board→window | Simulated infinite canvas with culling; no storage limit |
| Android later | Not MyGo's scope (desktop) | Same board.json + assets/ format; separate Android client |

If any of these prove impossible in practice, I will say so and propose the
closest alternative rather than faking it.
