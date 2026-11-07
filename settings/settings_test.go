package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	// Point the store at a temp dir by monkeypatching is not possible;
	// instead just exercise Save/Load against the real app-data location.
	// Save then Load should round-trip.
	s := Defaults()
	s.Theme = ThemeDark
	s.FontFamily = "Inter"
	s.BackgroundDim = 0.7
	if err := Save(s); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded := Load()
	if loaded.Theme != ThemeDark {
		t.Errorf("theme = %q, want %q", loaded.Theme, ThemeDark)
	}
	if loaded.FontFamily != "Inter" {
		t.Errorf("font = %q, want %q", loaded.FontFamily, "Inter")
	}
	if loaded.BackgroundDim != 0.7 {
		t.Errorf("dim = %v, want 0.7", loaded.BackgroundDim)
	}
	// Clean up the settings file we wrote.
	p, _ := path()
	os.Remove(p)
	os.Remove(filepath.Dir(p))
}

func TestDefaultsAreSane(t *testing.T) {
	d := Defaults()
	if d.Theme != ThemeSystem {
		t.Errorf("default theme = %q", d.Theme)
	}
	if d.BackgroundDim < 0 || d.BackgroundDim > 1 {
		t.Errorf("default dim out of range: %v", d.BackgroundDim)
	}
}
