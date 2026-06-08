package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/devlopersabbir/tview-cli/internal/model"
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

func TestIndicatorMessage(t *testing.T) {
	message := indicatorMessage(
		"BTCUSDT",
		"15m",
		"BEARISH",
		"BULLISH",
		model.Candle{
			Open:   100,
			High:   112,
			Low:    98,
			Close:  110,
			Volume: 42.5,
		},
		"Jun 08 8:45AM",
	)

	wants := []string{
		"<b>BTCUSDT 15M indicator changed</b>",
		"BEARISH -> <b>BULLISH</b>",
		"Price: <b>110.00</b>",
		"Move: <b>+10.00%</b>",
		"Time: Jun 08 8:45AM",
		"O 100.00",
		"V 42.50",
	}

	for _, want := range wants {
		if !strings.Contains(message, want) {
			t.Fatalf("indicatorMessage missing %q in %q", want, message)
		}
	}
}

func TestLatestClosedCandleUsesPreviousCandle(t *testing.T) {
	candles := []model.Candle{
		{Timestamp: time.Date(2026, time.June, 8, 9, 0, 0, 0, time.UTC), Close: 100},
		{Timestamp: time.Date(2026, time.June, 8, 9, 5, 0, 0, time.UTC), Close: 105},
		{Timestamp: time.Date(2026, time.June, 8, 9, 10, 0, 0, time.UTC), Close: 99},
	}

	got, ok := latestClosedCandle(candles)
	if !ok {
		t.Fatal("latestClosedCandle returned false")
	}
	if got.Close != 105 {
		t.Fatalf("latestClosedCandle close = %v, want 105", got.Close)
	}
}

func TestLatestClosedCandleRequiresTwoCandles(t *testing.T) {
	if _, ok := latestClosedCandle([]model.Candle{{Close: 100}}); ok {
		t.Fatal("latestClosedCandle returned true with one candle")
	}
}

func TestLatestClosedMarkerSignal(t *testing.T) {
	tests := []struct {
		name    string
		candles []model.Candle
		want    indicatorSignal
	}{
		{
			name: "bullish when newest closed marker is lowest low",
			candles: []model.Candle{
				{High: 120, Low: 95},
				{High: 108, Low: 90},
				{High: 112, Low: 99},
			},
			want: indicatorBullish,
		},
		{
			name: "bearish when newest closed marker is highest high",
			candles: []model.Candle{
				{High: 110, Low: 90},
				{High: 120, Low: 98},
				{High: 112, Low: 99},
			},
			want: indicatorBearish,
		},
		{
			name: "uses existing marker even when latest closed candle has no marker",
			candles: []model.Candle{
				{High: 130, Low: 90},
				{High: 120, Low: 98},
				{High: 112, Low: 99},
			},
			want: indicatorBearish,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := latestClosedMarkerSignal(tt.candles); got != tt.want {
				t.Fatalf("latestClosedMarkerSignal = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLatestClosedMarkerIgnoresLiveCandle(t *testing.T) {
	candles := []model.Candle{
		{High: 110, Low: 90, Close: 100},
		{High: 120, Low: 95, Close: 110},
		{High: 200, Low: 10, Close: 150},
	}

	got, ok := latestClosedMarker(candles)
	if !ok {
		t.Fatal("latestClosedMarker returned false")
	}
	if got.Signal != indicatorBearish {
		t.Fatalf("latestClosedMarker signal = %v, want %v", got.Signal, indicatorBearish)
	}
	if got.Candle.Close != 110 {
		t.Fatalf("latestClosedMarker close = %v, want 110", got.Candle.Close)
	}
}

func TestIndicatorSignalLabel(t *testing.T) {
	tests := []struct {
		signal indicatorSignal
		want   string
	}{
		{signal: indicatorBullish, want: "BULLISH"},
		{signal: indicatorBearish, want: "BEARISH"},
		{signal: indicatorNone, want: "NONE"},
	}

	for _, tt := range tests {
		if got := indicatorSignalLabel(tt.signal); got != tt.want {
			t.Fatalf("indicatorSignalLabel(%v) = %q, want %q", tt.signal, got, tt.want)
		}
	}
}
