// Package exchange provides clients for fetching market data from crypto exchanges.
package exchange

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devlopersabbir/tview/internal/model"
)

const bybitBaseURL = "https://api.bybit.com/v5/market/kline"

// intervalToBybit converts common interval strings (e.g. "1h", "4h") to the
// Bybit kline API format. Supported Bybit intervals: 1 3 5 15 30 60 120 240 360 720 D W M.
func intervalToBybit(iv string) string {
	switch strings.ToLower(iv) {
	case "1m":
		return "1"
	case "3m":
		return "3"
	case "5m":
		return "5"
	case "15m":
		return "15"
	case "30m":
		return "30"
	case "1h":
		return "60"
	case "2h":
		return "120"
	case "4h":
		return "240"
	case "6h":
		return "360"
	case "12h":
		return "720"
	case "1d", "d":
		return "D"
	case "1w", "w":
		return "W"
	case "1m_month": // disambiguate monthly
		return "M"
	}
	return iv
}

// bybitResponse mirrors the Bybit v5 kline API response envelope.
type bybitResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		List [][]string `json:"list"`
	} `json:"result"`
}

// FetchBybit retrieves the latest candlestick data for a given symbol and interval
// from the Bybit Spot market. Returns candles ordered oldest → newest.
func FetchBybit(symbol, interval string) ([]model.Candle, error) {
	url := fmt.Sprintf(
		"%s?category=spot&symbol=%s&interval=%s&limit=%d",
		bybitBaseURL, symbol, intervalToBybit(interval), model.NumCandles,
	)

	resp, err := http.Get(url) //nolint:noctx // simple CLI, no context needed
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	var result bybitResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}
	if result.RetCode != 0 {
		return nil, fmt.Errorf("bybit API error %d: %s", result.RetCode, result.RetMsg)
	}

	// Bybit returns rows newest-first; reverse to get chronological order.
	rows := result.Result.List
	candles := make([]model.Candle, 0, len(rows))

	for i := len(rows) - 1; i >= 0; i-- {
		r := rows[i]
		// Row layout: [startTime, open, high, low, close, volume, turnover]
		if len(r) < 6 {
			continue
		}

		tsMs, _ := strconv.ParseInt(r[0], 10, 64)
		o, _ := strconv.ParseFloat(r[1], 64)
		h, _ := strconv.ParseFloat(r[2], 64)
		l, _ := strconv.ParseFloat(r[3], 64)
		c, _ := strconv.ParseFloat(r[4], 64)
		v, _ := strconv.ParseFloat(r[5], 64)

		candles = append(candles, model.Candle{
			Open:      o,
			High:      h,
			Low:       l,
			Close:     c,
			Volume:    v,
			IsBull:    c >= o,
			Timestamp: time.Unix(tsMs/1000, 0),
		})
	}

	return candles, nil
}
