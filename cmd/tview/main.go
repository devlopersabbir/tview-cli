// tview - A terminal candlestick chart viewer powered by Bybit market data.
//
// Author:  Sabbir Hossain Shuvo
// GitHub:  https://github.com/devlopersabbir
// License: MIT
package main

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"os/signal"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/spf13/cobra"

	"github.com/devlopersabbir/tview-cli/internal/chart"
	"github.com/devlopersabbir/tview-cli/internal/color"
	"github.com/devlopersabbir/tview-cli/internal/exchange"
	"github.com/devlopersabbir/tview-cli/internal/model"
	"github.com/devlopersabbir/tview-cli/internal/notify"
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
	var full bool
	var box bool
	var watch bool

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
  tview ada 1d
  tview btc 15m --full
  tview btc 15m --box
  tview btc 15m --watch`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if full && box {
				return fmt.Errorf("--full and --box cannot be used together")
			}
			return runChart(cmd, args, full, watch)
		},
	}

	root.Flags().BoolVar(&full, "full", false, "use the full terminal width for the chart")
	root.Flags().BoolVar(&box, "box", false, "use the default boxed chart width")
	root.Flags().BoolVar(&watch, "watch", false, "watch for bullish/bearish changes and notify Telegram")
	root.SetVersionTemplate(versionOutput())
	root.AddCommand(newVersionCmd())
	return root
}

func runChart(_ *cobra.Command, args []string, full, watch bool) error {
	symbol := strings.ToUpper(args[0])
	if !strings.HasSuffix(symbol, "USDT") {
		symbol += "USDT"
	}
	interval := args[1]

	limit := model.NumCandles
	if full {
		limit = fullWidthCandles()
	}

	if watch {
		return watchIndicator(symbol, interval)
	}

	candles, err := exchange.FetchBybit(symbol, interval, limit)
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

func watchIndicator(symbol, interval string) error {
	loadDotenv(".env")

	client := notify.Telegram{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID: os.Getenv("TELEGRAM_CHAT_ID"),
	}
	if client.Token == "" || client.ChatID == "" {
		return fmt.Errorf("set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID before using --watch")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

	fmt.Printf("%sWatching%s %s %s every 30s. Press Ctrl+C to stop.\n",
		color.Gray, color.Reset, symbol, strings.ToUpper(interval),
	)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var state indicatorWatchState
	for {
		if err := checkIndicator(symbol, interval, client, &state); err != nil {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", color.Red, err, color.Reset)
		}

		select {
		case <-ticker.C:
			continue
		case <-interrupt:
			fmt.Printf("\n%sStopped watch.%s\n", color.Gray, color.Reset)
			return nil
		}
	}
}

type indicatorWatchState struct {
	initialized bool
	signal      indicatorSignal
	markerTime  time.Time
}

type indicatorSignal int

const (
	indicatorNone indicatorSignal = iota
	indicatorBullish
	indicatorBearish
)

func loadDotenv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		_ = os.Setenv(key, value)
	}
}

func checkIndicator(symbol, interval string, client notify.Telegram, state *indicatorWatchState) error {
	candles, err := exchange.FetchBybit(symbol, interval, model.NumCandles)
	if err != nil {
		return err
	}
	if len(candles) < 2 {
		return fmt.Errorf("not enough data for %s", symbol)
	}

	marker, ok := latestClosedMarker(candles)
	if !ok {
		return fmt.Errorf("no closed marker data for %s", symbol)
	}
	closed, ok := latestClosedCandle(candles)
	if !ok {
		return fmt.Errorf("not enough data for %s", symbol)
	}
	signal := marker.Signal
	label := indicatorSignalLabel(signal)
	timestamp := marker.Candle.Timestamp.Local().Format("Jan 02 3:04PM")

	if !state.initialized {
		state.initialized = true
		state.signal = signal
		state.markerTime = marker.Candle.Timestamp
		fmt.Printf("%s[%s]%s %s %s baseline marker: %s at %.2f (%s)\n",
			color.Gray, time.Now().Format("3:04:05PM"), color.Reset,
			symbol, strings.ToUpper(interval), label, marker.Candle.Close, timestamp,
		)
		return nil
	}

	if !marker.Candle.Timestamp.After(state.markerTime) {
		fmt.Printf("%s[%s]%s %s %s marker still %s at %.2f (%s); latest closed %.2f\n",
			color.Gray, time.Now().Format("3:04:05PM"), color.Reset,
			symbol, strings.ToUpper(interval), label, marker.Candle.Close, timestamp, closed.Close,
		)
		return nil
	}

	if state.signal == signal {
		state.markerTime = marker.Candle.Timestamp
		fmt.Printf("%s[%s]%s %s %s new marker but still %s at %.2f (%s)\n",
			color.Gray, time.Now().Format("3:04:05PM"), color.Reset,
			symbol, strings.ToUpper(interval), label, marker.Candle.Close, timestamp,
		)
		return nil
	}

	oldLabel := indicatorSignalLabel(state.signal)
	message := indicatorMessage(symbol, interval, oldLabel, label, marker.Candle, timestamp)
	preview, err := chart.RenderPNG(symbol, interval, candles)
	if err != nil {
		return fmt.Errorf("render chart preview: %w", err)
	}

	if err := client.SendPhoto(fmt.Sprintf("%s-%s.png", symbol, strings.ToLower(interval)), preview, message); err != nil {
		return err
	}

	state.signal = signal
	state.markerTime = marker.Candle.Timestamp
	fmt.Printf("%s[%s]%s notified yellow marker: %s -> %s at %.2f (%s)\n",
		color.Gray, time.Now().Format("3:04:05PM"), color.Reset,
		oldLabel, label, marker.Candle.Close, timestamp,
	)
	return nil
}

type indicatorMarker struct {
	Signal indicatorSignal
	Candle model.Candle
}

func latestClosedMarker(candles []model.Candle) (indicatorMarker, bool) {
	if len(candles) < 2 {
		return indicatorMarker{}, false
	}

	closedCandles := candles[:len(candles)-1]
	highIdx, lowIdx := 0, 0
	maxHigh, minLow := closedCandles[0].High, closedCandles[0].Low
	for i, c := range closedCandles {
		if c.High > maxHigh {
			maxHigh = c.High
			highIdx = i
		}
		if c.Low < minLow {
			minLow = c.Low
			lowIdx = i
		}
	}

	if lowIdx > highIdx {
		return indicatorMarker{Signal: indicatorBullish, Candle: closedCandles[lowIdx]}, true
	}
	return indicatorMarker{Signal: indicatorBearish, Candle: closedCandles[highIdx]}, true
}

func latestClosedMarkerSignal(candles []model.Candle) indicatorSignal {
	marker, ok := latestClosedMarker(candles)
	if !ok {
		return indicatorNone
	}
	return marker.Signal
}

func latestClosedCandle(candles []model.Candle) (model.Candle, bool) {
	if len(candles) < 2 {
		return model.Candle{}, false
	}

	// The latest Bybit candle can still be forming, so use the previous candle
	// as the confirmed signal source to avoid intrabar notification noise.
	return candles[len(candles)-2], true
}

func indicatorMessage(symbol, interval, oldLabel, newLabel string, candle model.Candle, timestamp string) string {
	change := candle.Close - candle.Open
	changePct := 0.0
	if candle.Open != 0 {
		changePct = change / candle.Open * 100
	}

	return fmt.Sprintf(
		"<b>%s %s indicator changed</b>\n\n"+
			"%s -> <b>%s</b>\n"+
			"Price: <b>%.2f</b>\n"+
			"Move: <b>%+.2f%%</b>\n"+
			"Time: %s\n\n"+
			"<pre>O %.2f\nH %.2f\nL %.2f\nC %.2f\nV %.2f</pre>",
		html.EscapeString(symbol),
		html.EscapeString(strings.ToUpper(interval)),
		html.EscapeString(oldLabel),
		html.EscapeString(newLabel),
		candle.Close,
		changePct,
		html.EscapeString(timestamp),
		candle.Open,
		candle.High,
		candle.Low,
		candle.Close,
		candle.Volume,
	)
}

func indicatorLabel(isBull bool) string {
	if isBull {
		return "BULLISH"
	}
	return "BEARISH"
}

func indicatorSignalLabel(signal indicatorSignal) string {
	switch signal {
	case indicatorBullish:
		return "BULLISH"
	case indicatorBearish:
		return "BEARISH"
	default:
		return "NONE"
	}
}

func fullWidthCandles() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return model.NumCandles
	}

	// Price column and separators take most of the non-chart space. Keep this
	// slightly generous so --full visually reaches the terminal edge.
	candles := width - 10
	if candles < model.NumCandles {
		return model.NumCandles
	}
	if candles > model.MaxCandles {
		return model.MaxCandles
	}
	return candles
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
