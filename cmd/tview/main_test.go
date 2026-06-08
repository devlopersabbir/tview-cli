package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndicatorLabel(t *testing.T) {
	tests := []struct {
		name   string
		isBull bool
		want   string
	}{
		{name: "bullish", isBull: true, want: "BULLISH"},
		{name: "bearish", isBull: false, want: "BEARISH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := indicatorLabel(tt.isBull); got != tt.want {
				t.Fatalf("indicatorLabel(%v) = %q, want %q", tt.isBull, got, tt.want)
			}
		})
	}
}

func TestLoadDotenv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("TVIEW_TEST_TOKEN=abc123\nTVIEW_TEST_KEEP=from_file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TVIEW_TEST_TOKEN", "")
	t.Setenv("TVIEW_TEST_KEEP", "from_env")

	loadDotenv(envPath)

	if got := os.Getenv("TVIEW_TEST_TOKEN"); got != "abc123" {
		t.Fatalf("TVIEW_TEST_TOKEN = %q, want %q", got, "abc123")
	}
	if got := os.Getenv("TVIEW_TEST_KEEP"); got != "from_env" {
		t.Fatalf("TVIEW_TEST_KEEP = %q, want %q", got, "from_env")
	}
}
