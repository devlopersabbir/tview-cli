package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/adshao/go-binance/v2"
	"github.com/guptarohit/asciigraph"
	"github.com/spf13/cobra"
)

const version = "1.0.0"

func main() {
	// Base command: tview
	var rootCmd = &cobra.Command{
		Use:   "tview [symbol] [interval]",
		Short: "tview is a lightweight CLI tool to view crypto prices in your terminal",
		Long:  `A quick terminal utility to fetch and display candlestick trends using ASCII charts.`,
		Args:  cobra.ExactArgs(2), // Requires exactly 2 arguments: symbol and interval
		Run: func(cmd *cobra.Command, args []string) {
			// Format arguments (e.g., btc -> BTCUSDT, 15m -> 15m)
			symbol := strings.ToUpper(args[0])
			if !strings.HasSuffix(symbol, "USDT") {
				symbol = symbol + "USDT"
			}
			interval := args[1]

			fmt.Printf("Fetching data for %s (%s)... \n\n", symbol, interval)
			fetchAndDrawChart(symbol, interval)
		},
	}

	// Version flag
	var versionFlag bool
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print the version of tview")

	// Intercept execution if --version is passed
	cobra.OnInitialize(func() {
		if versionFlag {
			fmt.Printf("tview version %s\n", version)
			os.Exit(0)
		}
	})

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func fetchAndDrawChart(symbol, interval string) {
	client := binance.NewClient("", "") // No API keys needed for public market data

	// Fetch the last 50 candlesticks
	klines, err := client.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(50).
		Do(context.Background())

	if err != nil {
		fmt.Printf("❌ Error fetching data from Binance: %v\n", err)
		return
	}

	// Extract Close prices to build the trendline
	var closingPrices []float64
	for _, kline := range klines {
		price, err := strconv.ParseFloat(kline.Close, 64)
		if err != nil {
			continue
		}
		closingPrices = append(closingPrices, price)
	}

	if len(closingPrices) == 0 {
		fmt.Println("No data found. Please check your symbol or interval.")
		return
	}

	// Render the ASCII chart
	chart := asciigraph.Plot(
		closingPrices,
		asciigraph.Height(12),
		asciigraph.Width(60),
		asciigraph.Caption(fmt.Sprintf("Price Trend for %s (%s)", symbol, interval)),
	)

	fmt.Println(chart)

	// Print current latest price
	latestPrice := closingPrices[len(closingPrices)-1]
	fmt.Printf("\n💰 Current Price: $%.2f\n", latestPrice)
}
