package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/spf13/cobra"
)

type Candle struct {
	Open, High, Low, Close, Volume float64
	IsBull                         bool
	Timestamp                      time.Time
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "tview [symbol] [interval]",
		Short: "Beautiful inline crypto candlestick and volume viewer",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			symbol := strings.ToUpper(args[0])
			if !strings.HasSuffix(symbol, "USDT") {
				symbol += "USDT"
			}
			printPremiumChart(symbol, args[1])
		},
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printPremiumChart(symbol, interval string) {
	chartHeight := 12
	volumeHeight := 3
	numCandles := 40 // Fits beautifully on standard terminal windows

	client := binance.NewClient("", "")
	klines, err := client.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(numCandles).
		Do(context.Background())

	if err != nil {
		fmt.Printf("\033[31m❌ Error fetching data from Binance: %v\033[0m\n", err)
		return
	}

	var candles []Candle
	minPrice, maxPrice := math.MaxFloat64, 0.0
	maxVolume := 0.0
	var highestCandleIdx, lowestCandleIdx int

	for i, k := range klines {
		o, _ := strconv.ParseFloat(k.Open, 64)
		h, _ := strconv.ParseFloat(k.High, 64)
		l, _ := strconv.ParseFloat(k.Low, 64)
		c, _ := strconv.ParseFloat(k.Close, 64)
		v, _ := strconv.ParseFloat(k.Volume, 64)
		t := time.Unix(k.OpenTime/1000, 0)

		if l < minPrice {
			minPrice = l
			lowestCandleIdx = i
		}
		if h > maxPrice {
			maxPrice = h
			highestCandleIdx = i
		}
		if v > maxVolume {
			maxVolume = v
		}

		candles = append(candles, Candle{o, h, l, c, v, c >= o, t})
	}

	// ANSI Styles
	green := "\033[38;5;46m"
	red := "\033[38;5;196m"
	gray := "\033[38;5;244m"
	darkGray := "\033[38;5;237m"
	whiteBold := "\033[1;37m"
	reset := "\033[0m"

	// Header Summary Row
	latest := candles[len(candles)-1]
	statusColor := red
	if latest.IsBull {
		statusColor = green
	}
	fmt.Printf("\n📊 %s%s%s  %s•%s  %s%s%s\n", whiteBold, symbol, reset, gray, reset, statusColor, strings.ToUpper(interval), reset)
	fmt.Printf("%sHigh: %.2f   Low: %.2f%s\n\n", gray, maxPrice, minPrice, reset)

	// 1. RENDER CANDLESTICK GRID
	for y := 0; y < chartHeight; y++ {
		// Y-Axis Price Label
		targetPrice := maxPrice - ((maxPrice - minPrice) * float64(y) / float64(chartHeight))
		fmt.Printf("%s%9.1f │%s ", gray, targetPrice, reset)

		for i, candle := range candles {
			yHigh := int(float64(chartHeight) * (maxPrice - candle.High) / (maxPrice - minPrice))
			yLow := int(float64(chartHeight) * (maxPrice - candle.Low) / (maxPrice - minPrice))

			bodyTop := math.Max(candle.Open, candle.Close)
			bodyBottom := math.Min(candle.Open, candle.Close)
			yTop := int(float64(chartHeight) * (maxPrice - bodyTop) / (maxPrice - minPrice))
			yBottom := int(float64(chartHeight) * (maxPrice - bodyBottom) / (maxPrice - minPrice))

			// Clamp indexes to avoid out-of-bound micro rendering glitches
			if yHigh < 0 {
				yHigh = 0
			}
			if yLow >= chartHeight {
				yLow = chartHeight - 1
			}

			color := red
			if candle.IsBull {
				color = green
			}

			// Add Price Target indicator callouts inline if it hits global peak/floor
			if y == yHigh && i == highestCandleIdx {
				fmt.Printf("%s▲%s", whiteBold, reset)
				continue
			}
			if y == yLow && i == lowestCandleIdx {
				fmt.Printf("%s▼%s", whiteBold, reset)
				continue
			}

			// Core Block Rendering Engine
			if y >= yTop && y <= yBottom {
				fmt.Printf("%s█%s ", color, reset) // Thick body block
			} else if y >= yHigh && y <= yLow {
				fmt.Printf("%s│%s ", color, reset) // Wick line
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}

	// Clean Layout Divider
	fmt.Printf("%s──────────┼%s", gray, darkGray)
	for i := 0; i < numCandles; i++ {
		fmt.Print("──")
	}
	fmt.Printf("%s\n", reset)

	// 2. RENDER VOLUME BARS
	for y := 0; y < volumeHeight; y++ {
		if y == 1 {
			fmt.Printf("%s  Volume  │%s ", gray, reset)
		} else {
			fmt.Printf("%s          │%s ", gray, reset)
		}

		for _, candle := range candles {
			vHeight := int(float64(volumeHeight) * (candle.Volume / maxVolume))
			color := red
			if candle.IsBull {
				color = green
			}

			if (volumeHeight - y) <= vHeight {
				fmt.Printf("%s▄%s ", color, reset) // Uses bottom half blocks for smooth volume spikes
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}

	// 3. X-AXIS TIME STAMPS
	fmt.Printf("%s          │%s ", gray, reset)
	for i, candle := range candles {
		if i%8 == 0 { // Print time every 8 candles to avoid overlapping text
			fmt.Printf("%s%s%s ", gray, candle.Timestamp.Format("15:04"), reset)
		} else if i%8 > 2 {
			// Do nothing to reserve padding room for the 5-character layout timestamp
		} else {
			fmt.Print(" ")
		}
	}

	// Bottom metrics bar
	fmt.Printf("\n\n💰 Current Price: %s$%.2f%s  │  Vol: %s%.2f%s\n\n", statusColor, latest.Close, reset, gray, latest.Volume, reset)
}
