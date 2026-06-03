// Package chart renders OHLCV candlestick charts in the terminal.
package chart

import (
	"fmt"
	"math"
	"strings"

	"github.com/devlopersabbir/tview/internal/color"
	"github.com/devlopersabbir/tview/internal/model"
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

	decimals := decimalsFor(latest.Close)
	priceFmt := fmt.Sprintf("%%.%df", decimals)

	// ── Header ────────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Printf("%s%s%s  %s•%s  %s%s %s%s  %s%s\n",
		color.Bold+color.White, symbol, color.Reset,
		color.Gray, color.Reset,
		statusColor+color.Bold, strings.ToUpper(interval), color.Reset,
		color.Gray, arrow+" "+fmt.Sprintf(priceFmt, latest.Close),
		color.Reset,
	)
	fmt.Printf("%s  H: %s"+priceFmt+"%s  L: %s"+priceFmt+"%s\n\n",
		color.DarkGray,
		color.White, maxPrice, color.Reset+color.DarkGray,
		color.White, minPrice, color.Reset,
	)

	highRow := priceToRow(candles[highIdx].High)
	lowRow := priceToRow(candles[lowIdx].Low)
	n := len(candles)

	// ── Candlestick grid ──────────────────────────────────────────────────
	for y := 0; y < model.ChartHeight; y++ {
		price := maxPrice - priceRange*float64(y)/float64(model.ChartHeight)
		fmt.Printf("%s%9.*f │%s", color.Gray, decimals, price, color.Reset)

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
		fmt.Println()
	}

	// ── X divider ─────────────────────────────────────────────────────────
	fmt.Printf("%s──────────┼%s", color.Gray, color.DarkGray)
	for i := 0; i < n; i++ {
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
			barH := int(math.Round(float64(model.VolumeHeight) * c.Volume / maxVolume))
			if (model.VolumeHeight - y) <= barH {
				fmt.Printf("%s▄%s", col, color.Reset)
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}

	// ── Time axis ─────────────────────────────────────────────────────────
	fmt.Printf("%s          │%s ", color.Gray, color.Reset)
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
	fmt.Printf("\n%s  Price%s %s"+priceFmt+"%s    %sVol%s %.2f\n\n",
		color.Gray, color.Reset,
		statusColor+color.Bold, latest.Close, color.Reset,
		color.Gray, color.Reset, latest.Volume,
	)
}
