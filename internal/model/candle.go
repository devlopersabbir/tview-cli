// Package model defines the core data types used across tview-cli.
package model

import "time"

// Candle represents a single OHLCV candlestick data point.
type Candle struct {
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	IsBull    bool
	Timestamp time.Time
}

// Chart layout constants.
const (
	ChartHeight  = 18
	VolumeHeight = 4
	NumCandles   = 60
	MaxCandles   = 1000
)
