# tview-cli

> Real-time crypto candlestick charts, rendered directly in your terminal — powered by [Bybit](https://bybit.com).

```
BTCUSDT  •  1H ▲ 67,420.00
  H: 68,100.00  L: 66,980.00

  68100.00 │                          ▼
  67877.77 │              █ █  █ █ █ │
  67655.55 │    │  │ █  █ █ █  █ █ █ │
  ...
```

## Features

- 📊 Candlestick chart with wick and body rendering
- 📦 Volume bars per candle
- 🕐 Time axis with date labels
- ▲▼ High/Low markers in yellow
- 🎨 ANSI color output (bull = green, bear = red)
- ⚡ Zero config — works globally via Bybit public API

## Installation

### Homebrew (macOS / Linux)
```bash
brew install devlopersabbir/tap/tview
```

### Go install
```bash
go install github.com/devlopersabbir/tview/cmd/tview@latest
```

### Binary download
Download the latest binary for your platform from the [Releases](https://github.com/devlopersabbir/tview-cli/releases) page.

## Usage

```
tview [symbol] [interval]
```

| Argument   | Description                                      | Example      |
|------------|--------------------------------------------------|--------------|
| `symbol`   | Trading pair (USDT suffix added automatically)   | `btc`, `eth`, `SOLUSDT` |
| `interval` | Candle interval                                  | `1m` `5m` `15m` `1h` `4h` `1d` |

### Examples

```bash
tview btc 1h        # Bitcoin — 1-hour candles
tview eth 15m       # Ethereum — 15-minute candles
tview SOLUSDT 4h    # Solana — 4-hour candles
tview ada 1d        # Cardano — daily candles
```

### Version

```bash
tview version
```

## Supported intervals

`1m` · `3m` · `5m` · `15m` · `30m` · `1h` · `2h` · `4h` · `6h` · `12h` · `1d` · `1w`

## Project structure

```
tview-cli/
├── cmd/
│   └── tview/
│       └── main.go          # CLI entry point & cobra commands
├── internal/
│   ├── chart/
│   │   └── render.go        # Terminal chart renderer
│   ├── color/
│   │   └── color.go         # ANSI color constants
│   ├── exchange/
│   │   └── bybit.go         # Bybit API client
│   └── model/
│       └── candle.go        # Candle type & layout constants
├── .github/
│   └── workflows/
│       └── release.yml      # Automated binary release pipeline
├── .goreleaser.yml           # Cross-platform build config
├── Makefile                  # Dev helpers
└── README.md
```

## Development

```bash
# Build
make build

# Build & run
make run

# Run tests
make test

# Local snapshot release (requires goreleaser)
make snapshot
```

## Release

Releases are automated via GitHub Actions. Push a semver tag to trigger a cross-platform build:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GoReleaser will build binaries for **Linux**, **macOS**, and **Windows** on **amd64** and **arm64**, attach them to the GitHub Release, and generate a changelog from your commit history.

## Author

**Sabbir Hossain Shuvo**  
GitHub: [@devlopersabbir](https://github.com/devlopersabbir)

## License

[MIT](LICENSE) © Sabbir Hossain Shuvo
