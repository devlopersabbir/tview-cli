package chart

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/devlopersabbir/tview-cli/internal/model"
)

var (
	pngBg       = color.RGBA{18, 23, 29, 255}
	pngPanel    = color.RGBA{25, 31, 38, 255}
	pngGrid     = color.RGBA{68, 76, 86, 255}
	pngText     = color.RGBA{214, 219, 226, 255}
	pngMuted    = color.RGBA{145, 151, 160, 255}
	pngBull     = color.RGBA{0, 220, 185, 255}
	pngBear     = color.RGBA{255, 73, 91, 255}
	pngMarker   = color.RGBA{255, 213, 0, 255}
	pngWhite    = color.RGBA{245, 247, 250, 255}
	pngFontFace = basicfont.Face7x13
)

// RenderPNG renders a static candlestick chart preview for Telegram.
func RenderPNG(symbol, interval string, candles []model.Candle) ([]byte, error) {
	if len(candles) == 0 {
		return nil, fmt.Errorf("no candles to render")
	}

	width, height := 1000, 680
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{pngBg}, image.Point{}, draw.Src)

	left, right := 88, 44
	top, chartBottom := 116, 470
	volumeTop, volumeBottom := 502, 590
	plotW := width - left - right
	plotH := chartBottom - top

	fillRect(img, image.Rect(left, top, width-right, chartBottom), pngPanel)
	fillRect(img, image.Rect(left, volumeTop, width-right, volumeBottom), pngPanel)

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

	latest := candles[len(candles)-1]
	arrow := "▲"
	stateColor := pngBull
	if !latest.IsBull {
		arrow = "▼"
		stateColor = pngBear
	}
	changePct := 0.0
	if latest.Open != 0 {
		changePct = (latest.Close - latest.Open) / latest.Open * 100
	}
	decimals := decimalsFor(latest.Close)
	priceFmt := fmt.Sprintf("%%.%df", decimals)

	drawText(img, 18, 38, symbol, pngWhite)
	drawText(img, 95, 38, "•", pngGrid)
	drawText(img, 120, 38, strings.ToUpper(interval), pngMuted)
	drawText(img, 168, 38, arrow+" "+fmt.Sprintf(priceFmt, latest.Close), stateColor)
	drawText(img, 300, 38, fmt.Sprintf("%+.2f%%", changePct), stateColor)
	drawText(img, 18, 64, fmt.Sprintf("O "+priceFmt+"   H "+priceFmt+"   L "+priceFmt+"   V %s", latest.Open, latest.High, latest.Low, compactNumber(latest.Volume)), pngMuted)
	drawText(img, 18, 650, fmt.Sprintf("Last "+priceFmt+"   Move %+.2f%%   Vol %s   Max Vol %s", latest.Close, changePct, compactNumber(latest.Volume), compactNumber(maxVolume)), pngMuted)

	for i := 0; i <= 5; i++ {
		y := top + i*plotH/5
		line(img, left, y, width-right, y, pngGrid)
		price := maxPrice - priceRange*float64(i)/5
		drawText(img, 18, y+4, fmt.Sprintf(priceFmt, price), pngMuted)
	}
	line(img, left, top, left, chartBottom, pngGrid)
	line(img, width-right, top, width-right, chartBottom, pngGrid)
	line(img, left, volumeTop, left, volumeBottom, pngGrid)
	line(img, width-right, volumeTop, width-right, volumeBottom, pngGrid)

	priceY := func(price float64) int {
		y := top + int(float64(plotH)*(maxPrice-price)/priceRange)
		if y < top {
			return top
		}
		if y > chartBottom {
			return chartBottom
		}
		return y
	}

	step := float64(plotW) / float64(len(candles))
	bodyW := int(math.Max(3, math.Min(10, step*0.7)))
	for i, c := range candles {
		x := left + int((float64(i)+0.5)*step)
		col := pngBear
		if c.IsBull {
			col = pngBull
		}

		wickTop := priceY(c.High)
		wickBot := priceY(c.Low)
		bodyTop := priceY(math.Max(c.Open, c.Close))
		bodyBot := priceY(math.Min(c.Open, c.Close))
		if bodyBot == bodyTop {
			bodyBot++
		}

		line(img, x, wickTop, x, wickBot, col)
		fillRect(img, image.Rect(x-bodyW/2, bodyTop, x+bodyW/2+1, bodyBot+1), col)

		if maxVolume > 0 {
			barH := int(float64(volumeBottom-volumeTop) * c.Volume / maxVolume)
			fillRect(img, image.Rect(x-bodyW/2, volumeBottom-barH, x+bodyW/2+1, volumeBottom), col)
		}
	}

	drawText(img, left+int((float64(highIdx)+0.5)*step)-4, priceY(candles[highIdx].High)-8, "▼", pngMarker)
	drawText(img, left+int((float64(lowIdx)+0.5)*step)-4, priceY(candles[lowIdx].Low)+16, "▲", pngMarker)

	labels := timeAxisLabels(interval, candles)
	for i := 0; i < len(labels); {
		if labels[i] == ' ' {
			i++
			continue
		}
		j := i
		for j < len(labels) && labels[j] != ' ' {
			j++
		}
		x := left + int((float64(i)+0.5)*step)
		drawText(img, x, 616, labels[i:j], pngMuted)
		i = j
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func fillRect(img *image.RGBA, rect image.Rectangle, col color.Color) {
	draw.Draw(img, rect, &image.Uniform{col}, image.Point{}, draw.Src)
}

func line(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	if x1 == x2 {
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		fillRect(img, image.Rect(x1, y1, x1+1, y2+1), col)
		return
	}
	if y1 == y2 {
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		fillRect(img, image.Rect(x1, y1, x2+1, y1+1), col)
	}
}

func drawText(img *image.RGBA, x, y int, text string, col color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{col},
		Face: pngFontFace,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}
