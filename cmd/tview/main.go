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
	"os/exec"
	"os/signal"
	"runtime"
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
	var forex bool
	var forexWatch bool
	var update bool

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
  tview btc 15m --watch
  tview xau 5m --forex
  tview xau 1m --fw`,
		Args: func(cmd *cobra.Command, args []string) error {
			if update && len(args) == 0 {
				return nil
			}
			return cobra.ExactArgs(2)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if update {
				return runUpdate()
			}
			if full && box {
				return fmt.Errorf("--full and --box cannot be used together")
			}
			if forexWatch {
				forex = true
				watch = true
			}
			return runChart(cmd, args, full, watch, forex)
		},
	}

	root.Flags().BoolVar(&full, "full", false, "use the full terminal width for the chart")
	root.Flags().BoolVar(&box, "box", false, "use the default boxed chart width")
	root.Flags().BoolVar(&watch, "watch", false, "watch for bullish/bearish changes and notify Telegram")
	root.Flags().BoolVarP(&forex, "forex", "f", false, "use forex/metal symbol mapping, e.g. xau -> XAUUSD")
	root.Flags().BoolVar(&forexWatch, "fw", false, "shortcut for --forex --watch")
	root.Flags().BoolVar(&update, "update", false, "update tview to the latest release")
	root.SetVersionTemplate(versionOutput())
	root.AddCommand(newVersionCmd())
	return root
}

func runChart(_ *cobra.Command, args []string, full, watch, forex bool) error {
	market, err := resolveMarket(args[0], forex)
	if err != nil {
		return err
	}
	interval := args[1]

	limit := model.NumCandles
	if full {
		limit = fullWidthCandles()
	}

	if watch {
		return watchIndicator(market, interval)
	}

	candles, err := fetchMarketCandles(market, interval, limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s❌  %v%s\n", color.Red, err, color.Reset)
		return err
	}
	if len(candles) == 0 {
		fmt.Fprintf(os.Stderr, "%s❌  No data returned. Check symbol name.%s\n", color.Red, color.Reset)
		return fmt.Errorf("no data for %s", market.DisplaySymbol)
	}

	chart.Print(market.DisplaySymbol, interval, candles)
	return nil
}

type marketRequest struct {
	DisplaySymbol string
	BybitSymbol   string
}

func resolveMarket(rawSymbol string, forex bool) (marketRequest, error) {
	symbol := strings.ToUpper(strings.ReplaceAll(rawSymbol, "/", ""))
	if forex {
		switch symbol {
		case "XAU", "XAUUSD":
			return marketRequest{DisplaySymbol: "XAUUSD", BybitSymbol: "XAUTUSDT"}, nil
		default:
			return marketRequest{}, fmt.Errorf("--forex currently supports xau or xauusd")
		}
	}

	if !strings.HasSuffix(symbol, "USDT") {
		symbol += "USDT"
	}
	return marketRequest{DisplaySymbol: symbol, BybitSymbol: symbol}, nil
}

func fetchMarketCandles(market marketRequest, interval string, limit int) ([]model.Candle, error) {
	return exchange.FetchBybit(market.BybitSymbol, interval, limit)
}

func watchIndicator(market marketRequest, interval string) error {
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
		color.Gray, color.Reset, market.DisplaySymbol, strings.ToUpper(interval),
	)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var state indicatorWatchState
	for {
		if err := checkIndicator(market, interval, client, &state); err != nil {
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

func checkIndicator(market marketRequest, interval string, client notify.Telegram, state *indicatorWatchState) error {
	candles, err := fetchMarketCandles(market, interval, model.NumCandles)
	if err != nil {
		return err
	}
	if len(candles) < 2 {
		return fmt.Errorf("not enough data for %s", market.DisplaySymbol)
	}

	marker, ok := latestClosedMarker(candles)
	if !ok {
		return fmt.Errorf("no closed marker data for %s", market.DisplaySymbol)
	}
	closed, ok := latestClosedCandle(candles)
	if !ok {
		return fmt.Errorf("not enough data for %s", market.DisplaySymbol)
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
			market.DisplaySymbol, strings.ToUpper(interval), label, marker.Candle.Close, timestamp,
		)
		return nil
	}

	if !marker.Candle.Timestamp.After(state.markerTime) {
		fmt.Printf("%s[%s]%s %s %s marker still %s at %.2f (%s); latest closed %.2f\n",
			color.Gray, time.Now().Format("3:04:05PM"), color.Reset,
			market.DisplaySymbol, strings.ToUpper(interval), label, marker.Candle.Close, timestamp, closed.Close,
		)
		return nil
	}

	if state.signal == signal {
		state.markerTime = marker.Candle.Timestamp
		fmt.Printf("%s[%s]%s %s %s new marker but still %s at %.2f (%s)\n",
			color.Gray, time.Now().Format("3:04:05PM"), color.Reset,
			market.DisplaySymbol, strings.ToUpper(interval), label, marker.Candle.Close, timestamp,
		)
		return nil
	}

	oldLabel := indicatorSignalLabel(state.signal)
	message := indicatorMessage(market.DisplaySymbol, interval, oldLabel, label, marker.Candle, timestamp)
	preview, err := chart.RenderPNG(market.DisplaySymbol, interval, candles)
	if err != nil {
		return fmt.Errorf("render chart preview: %w", err)
	}

	if err := client.SendPhoto(fmt.Sprintf("%s-%s.png", market.DisplaySymbol, strings.ToLower(interval)), preview, message); err != nil {
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

func runUpdate() error {
	exe, _ := os.Executable()
	if shouldUseHomebrewUpdater(exe) {
		return runUpdateCommand("brew", "reinstall", "--cask", "devlopersabbir/tview-cli/tview")
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		return runUpdateCommand("sh", "-c", "curl -fsSL https://raw.githubusercontent.com/devlopersabbir/tview-cli/main/scripts/install.sh | sh")
	case "windows":
		return runUpdateCommand("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", "iwr https://raw.githubusercontent.com/devlopersabbir/tview-cli/main/scripts/install.ps1 -UseBasicParsing | iex")
	default:
		return fmt.Errorf("self-update is not supported on %s; download the latest release from https://github.com/devlopersabbir/tview-cli/releases", runtime.GOOS)
	}
}

func shouldUseHomebrewUpdater(exe string) bool {
	if runtime.GOOS != "darwin" || exe == "" {
		return false
	}
	if !strings.Contains(exe, "/homebrew/") && !strings.HasPrefix(exe, "/usr/local/") {
		return false
	}
	_, err := exec.LookPath("brew")
	return err == nil
}

func runUpdateCommand(name string, args ...string) error {
	fmt.Printf("%sUpdating tview...%s\n", color.Gray, color.Reset)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
