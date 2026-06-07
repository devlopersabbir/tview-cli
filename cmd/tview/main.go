// tview - A terminal candlestick chart viewer powered by Bybit market data.
//
// Author:  Sabbir Hossain Shuvo
// GitHub:  https://github.com/devlopersabbir
// License: MIT
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/devlopersabbir/tview-cli/internal/chart"
	"github.com/devlopersabbir/tview-cli/internal/color"
	"github.com/devlopersabbir/tview-cli/internal/exchange"
)

// Build-time variables injected by GoReleaser / ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "tview [symbol] [interval]",
		Short:   "Terminal candlestick chart viewer — powered by Bybit",
		Version: version,
		Long: fmt.Sprintf(`%stview%s — Real-time crypto candlestick charts in your terminal.

  Built by %sSabbir Hossain Shuvo%s  •  github.com/devlopersabbir
  Version %s%s%s  |  Commit %s  |  Built %s`,
			color.Bold+color.White, color.Reset,
			color.Green+color.Bold, color.Reset,
			color.Yellow, version, color.Reset,
			commit, date,
		),
		Example: `  tview eth 15m
  tview btc 4h
  tview SOLUSDT 1h
  tview ada 1d`,
		Args: cobra.ExactArgs(2),
		RunE: runChart,
	}

	root.SetVersionTemplate(versionOutput())
	root.AddCommand(newVersionCmd())
	return root
}

func runChart(_ *cobra.Command, args []string) error {
	symbol := strings.ToUpper(args[0])
	if !strings.HasSuffix(symbol, "USDT") {
		symbol += "USDT"
	}
	interval := args[1]

	candles, err := exchange.FetchBybit(symbol, interval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s❌  %v%s\n", color.Red, err, color.Reset)
		return err
	}
	if len(candles) == 0 {
		fmt.Fprintf(os.Stderr, "%s❌  No data returned. Check symbol name.%s\n", color.Red, color.Reset)
		return fmt.Errorf("no data for %s", symbol)
	}

	chart.Print(symbol, interval, candles)
	return nil
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Print(versionOutput())
		},
	}
}

func versionOutput() string {
	return fmt.Sprintf("%stview%s  version %s%s%s\n  commit  %s\n  built   %s\n  author  %sSabbir Hossain Shuvo%s  •  github.com/devlopersabbir\n",
		color.Bold+color.White, color.Reset,
		color.Yellow, version, color.Reset,
		commit,
		date,
		color.Green+color.Bold, color.Reset,
	)
}
