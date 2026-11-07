// Package settings holds IdeaBoard's user settings: theme, note font,
// background image + dim, and window state. Persisted to settings.json next
// to the app data folder.
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Theme is the appearance source.
type Theme string

const (
	ThemeSystem Theme = "system"
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
)

// Settings is the user-configurable state of the app.
type Settings struct {
	Theme        Theme   `json:"theme"`
	FontFamily   string  `json:"fontFamily"`
	FontFile     string  `json:"fontFile"`   // path to a .ttf/.otf the user chose (optional)
	Background   string  `json:"background"` // path to a user-chosen background image (optional)
	BackgroundDim float32 `json:"backgroundDim"` // 0 (no dim) to 1 (fully dimmed)
	Window       WindowState `json:"window"`
}

type WindowState struct {
	X, Y       int  `json:"x,omitempty"`
	W, H       int  `json:"w,omitempty"`
	Maximized bool `json:"maximized,omitempty"`
}

// Defaults returns calm default settings.
func Defaults() Settings {
	return Settings{
		Theme:         ThemeSystem,
		FontFamily:    "",
		FontFile:      "",
		Background:    "",
		BackgroundDim: 0.45,
		Window:        WindowState{W: 900, H: 600},
	}
}

// dataDir returns the folder where settings.json lives.
func dataDir() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "IdeaBoard"), nil
}

// path returns the settings file path.
func path() (string, error) {
	d, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "settings.json"), nil
}

// Load reads settings from disk, returning Defaults on any error.
func Load() Settings {
	p, err := path()
	if err != nil {
		return Defaults()
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return Defaults()
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Defaults()
	}
	return s
}

// Save writes settings safely (temp file + rename).
func Save(s Settings) error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
