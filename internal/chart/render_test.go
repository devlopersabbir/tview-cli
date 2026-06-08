package chart

import (
	"bytes"
	"testing"
	"time"

	"github.com/devlopersabbir/tview-cli/internal/model"
)

func TestCompactNumber(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{name: "plain", value: 538.54, want: "538.54"},
		{name: "thousands", value: 2370, want: "2.37K"},
		{name: "millions", value: 12_500_000, want: "12.50M"},
		{name: "negative", value: -1420, want: "-1.42K"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compactNumber(tt.value); got != tt.want {
				t.Fatalf("compactNumber(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestRenderPNG(t *testing.T) {
	candles := make([]model.Candle, 12)
	for i := range candles {
		open := 100.0 + float64(i)
		close := open + 2
		if i%2 == 0 {
			close = open - 2
		}
		candles[i] = model.Candle{
			Open:      open,
			High:      open + 5,
			Low:       open - 5,
			Close:     close,
			Volume:    10 + float64(i),
			IsBull:    close >= open,
			Timestamp: time.Date(2026, time.June, 8, 9, i*5, 0, 0, time.UTC),
		}
	}

	got, err := RenderPNG("BTCUSDT", "5m", candles)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(got, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("RenderPNG did not return PNG bytes")
	}
}

func TestTimeAxisLabel(t *testing.T) {
	candle := model.Candle{
		Timestamp: time.Date(2026, time.June, 8, 14, 37, 0, 0, time.UTC),
	}

	tests := []struct {
		name     string
		interval string
		want     string
	}{
		{name: "minute interval", interval: "15m", want: "2:37PM"},
		{name: "hour interval", interval: "4h", want: "2PM"},
		{name: "daily interval", interval: "1d", want: "Mon"},
		{name: "weekly interval", interval: "w", want: "Mon"},
		{name: "monthly interval", interval: "1m_month", want: "Jun"},
		{name: "fallback interval", interval: "custom", want: "06/08"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := timeAxisLabel(tt.interval, candle); got != tt.want {
				t.Fatalf("timeAxisLabel(%q) = %q, want %q", tt.interval, got, tt.want)
			}
		})
	}
}

func TestTimeAxisLabelsStayWithinCandleWidth(t *testing.T) {
	candles := make([]model.Candle, 16)
	for i := range candles {
		candles[i].Timestamp = time.Date(2026, time.June, 8, 14, i, 0, 0, time.UTC)
	}

	got := timeAxisLabels("15m", candles)
	if len(got) != len(candles) {
		t.Fatalf("len(timeAxisLabels) = %d, want %d", len(got), len(candles))
	}
	if got != "2:00PM    2:15PM" {
		t.Fatalf("timeAxisLabels = %q, want %q", got, "2:00PM    2:15PM")
	}
}
