# IdeaBoard Windows Installer

This directory holds an Inno Setup script that builds a Windows installer for
IdeaBoard. The installer lets the user choose the install directory, creates a
Start Menu shortcut and an uninstaller, and installs the single-file Go binary.

## Prerequisites

- [Inno Setup 6+](https://jrsoftware.org/isinfo.php) (`iscc` on PATH).
- A built `ideaboard.exe` (see the root README for how to build).

## Build

From the repository root:

```bat
cd D:\projects\ideaboard
go build -trimpath -o ideaboard.exe .
cd installer
iscc install.iss
```

The output installer is written to `installer/output/IdeaBoard-<version>-setup.exe`.

## What the installer does

- Copies `ideaboard.exe` into the chosen install directory (`%LOCALAPPDATA%\Programs\IdeaBoard` by default, per-user, no admin elevation).
- Creates a Start Menu shortcut ("IdeaBoard") plus an uninstaller entry.
- Optionally creates a desktop icon (unchecked by default).
- Launches IdeaBoard after install (optional, can be declined).

## Notes

- MyGo is pure Go with no external DLLs, so the installer only needs the one
  executable. WebView2 is not used (native UI), so there's no webview
  runtime dependency.
- Boards and settings are stored under `%APPDATA%\IdeaBoard`, not in the
  install directory, so upgrades keep your boards.
- No code signing is applied; Windows SmartScreen may flag an unsigned
  installer on first run.
