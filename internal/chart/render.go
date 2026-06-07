// Package chart renders OHLCV candlestick charts in the terminal.
package chart

import (
	"fmt"
	"math"
	"strings"

	"github.com/devlopersabbir/tview-cli/internal/color"
	"github.com/devlopersabbir/tview-cli/internal/model"
)

// decimalsFor returns appropriate decimal precision based on price magnitude.
func decimalsFor(price float64) int {
	switch {
	case price < 0.01:
		return 6
	case price < 1:
		return 4
	case price < 100:
		return 3
	default:
		return 2
	}
}

// compactNumber formats large market values so footer labels stay readable.
func compactNumber(value float64) string {
	abs := math.Abs(value)
	units := []struct {
		value  float64
		suffix string
	}{
		{1_000_000_000_000, "T"},
		{1_000_000_000, "B"},
		{1_000_000, "M"},
		{1_000, "K"},
	}

	for _, unit := range units {
		if abs >= unit.value {
			return fmt.Sprintf("%.2f%s", value/unit.value, unit.suffix)
		}
	}
	return fmt.Sprintf("%.2f", value)
}

// priceToRowFunc returns a closure that maps a price to a chart row index.
func priceToRowFunc(maxPrice, priceRange float64) func(float64) int {
	return func(p float64) int {
		row := int(float64(model.ChartHeight) * (maxPrice - p) / priceRange)
		if row < 0 {
			return 0
		}
		if row >= model.ChartHeight {
			return model.ChartHeight - 1
		}
		return row
	}
}

// Print renders a full candlestick chart for the given candles to stdout.
func Print(symbol, interval string, candles []model.Candle) {
	// ── Compute chart metrics ─────────────────────────────────────────────
	minPrice, maxPrice := math.MaxFloat64, 0.0
	maxVolume := 0.0
	highIdx, lowIdx := 0, 0

	for i, c := range candles {
		if c.Low < minPrice {
			minPrice = c.Low
			lowIdx = i
		}
		if c.High > maxPrice {
			maxPrice = c.High
			highIdx = i
		}
		if c.Volume > maxVolume {
			maxVolume = c.Volume
		}
	}

	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}

	priceToRow := priceToRowFunc(maxPrice, priceRange)

	latest := candles[len(candles)-1]
	statusColor := color.For(latest.IsBull)
	arrow := "▲"
	if !latest.IsBull {
		arrow = "▼"
	}
	change := latest.Close - latest.Open
	changePct := 0.0
	if latest.Open != 0 {
		changePct = change / latest.Open * 100
	}

	decimals := decimalsFor(latest.Close)
	priceFmt := fmt.Sprintf("%%.%df", decimals)
	priceText := fmt.Sprintf(priceFmt, latest.Close)
	rangeText := fmt.Sprintf(priceFmt+" - "+priceFmt, minPrice, maxPrice)
	latestVolText := compactNumber(latest.Volume)
	maxVolText := compactNumber(maxVolume)

	// ── Header ────────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Printf("%s%s%s  %s%s%s  %s%s%s %s%s%s  %s%+.2f%%%s\n",
		color.Bold+color.White, symbol, color.Reset,
		color.DarkGray, "•", color.Reset,
		color.Gray, strings.ToUpper(interval), color.Reset,
		statusColor+color.Bold, arrow+" "+priceText, color.Reset,
		statusColor, changePct, color.Reset,
	)
	fmt.Printf("%s  O%s "+priceFmt+"  %sH%s "+priceFmt+"  %sL%s "+priceFmt+"  %sV%s %s  %sRange%s %s\n\n",
		color.Gray, color.Reset, latest.Open,
		color.Gray, color.Reset, latest.High,
		color.Gray, color.Reset, latest.Low,
		color.Gray, color.Reset, latestVolText,
		color.Gray, color.Reset, rangeText,
	)

	highRow := priceToRow(candles[highIdx].High)
	lowRow := priceToRow(candles[lowIdx].Low)
	n := len(candles)
	axisWidth := n + 1

	fmt.Printf("%s          ┌%s", color.Gray, color.DarkGray)
	for i := 0; i < axisWidth; i++ {
		fmt.Print("─")
	}
	fmt.Printf("%s\n", color.Reset)

	// ── Candlestick grid ──────────────────────────────────────────────────
	for y := 0; y < model.ChartHeight; y++ {
		price := maxPrice - priceRange*float64(y)/float64(model.ChartHeight)
		fmt.Printf("%s%9.*f │%s ", color.Gray, decimals, price, color.Reset)

		for i, c := range candles {
			col := color.For(c.IsBull)
			wickTop := priceToRow(c.High)
			wickBot := priceToRow(c.Low)
			bodyTop := priceToRow(math.Max(c.Open, c.Close))
			bodyBot := priceToRow(math.Min(c.Open, c.Close))

			// High/low markers
			if y == highRow && i == highIdx {
				fmt.Printf("%s▼%s", color.Bold+color.Yellow, color.Reset)
				continue
			}
			if y == lowRow && i == lowIdx {
				fmt.Printf("%s▲%s", color.Bold+color.Yellow, color.Reset)
				continue
			}

			switch {
			case y >= bodyTop && y <= bodyBot:
				fmt.Printf("%s█%s", col, color.Reset)
			case y >= wickTop && y <= wickBot:
				fmt.Printf("%s│%s", col, color.Reset)
			default:
				fmt.Print(" ")
			}
		}
		fmt.Printf("%s│%s\n", color.Gray, color.Reset)
	}

	// ── X divider ─────────────────────────────────────────────────────────
	fmt.Printf("%s──────────┼%s", color.Gray, color.DarkGray)
	for i := 0; i < axisWidth; i++ {
		fmt.Print("─")
	}
	fmt.Printf("%s\n", color.Reset)

	// ── Volume bars ───────────────────────────────────────────────────────
	for y := 0; y < model.VolumeHeight; y++ {
		if y == 1 {
			fmt.Printf("%s  volume  │%s ", color.Gray, color.Reset)
		} else {
			fmt.Printf("%s          │%s ", color.Gray, color.Reset)
		}
		for _, c := range candles {
			col := color.For(c.IsBull)
			barH := 0
			if maxVolume > 0 {
				barH = int(math.Round(float64(model.VolumeHeight) * c.Volume / maxVolume))
			}
			if (model.VolumeHeight - y) <= barH {
				fmt.Printf("%s▄%s", col, color.Reset)
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Printf("%s│%s\n", color.Gray, color.Reset)
	}

	// ── Time axis ─────────────────────────────────────────────────────────
	fmt.Printf("%s          └%s", color.Gray, color.DarkGray)
	for i := 0; i < axisWidth; i++ {
		fmt.Print("─")
	}
	fmt.Printf("%s\n", color.Reset)

	fmt.Printf("%s           %s", color.Gray, color.Reset)
	skip := 0
	for _, c := range candles {
		if skip == 0 {
			ts := c.Timestamp.Format("01/02")
			fmt.Printf("%s%s%s", color.Gray, ts, color.Reset)
			skip = len(ts) - 1
		} else {
			fmt.Print(" ")
			skip--
		}
	}
	fmt.Println()

	// ── Footer ────────────────────────────────────────────────────────────
	fmt.Printf("\n%s  Last%s %s%s%s   %sMove%s %s%+.2f%%%s   %sVol%s %s   %sMax Vol%s %s\n\n",
		color.Gray, color.Reset, statusColor+color.Bold, priceText, color.Reset,
		color.Gray, color.Reset, statusColor+color.Bold, changePct, color.Reset,
		color.Gray, color.Reset, latestVolText,
		color.Gray, color.Reset, maxVolText,
	)
}
