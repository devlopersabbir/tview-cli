# tview-cli

Real-time crypto candlestick charts in your terminal, powered by [Bybit](https://bybit.com).

```text
BTCUSDT  |  1H ▲ 67,420.00
  H: 68,100.00  L: 66,980.00

  68100.00 |                          ▼
  67877.77 |              █ █  █ █ █ |
  67655.55 |    |  | █  █ █ █  █ █ █ |
  ...
```

## Features

- Candlestick charts with wick and body rendering
- Volume bars for each candle
- Time axis with date labels
- High and low markers
- ANSI color output for bullish and bearish candles
- Zero config, using Bybit public market data

## Installation

### macOS

```bash
brew tap devlopersabbir/tview-cli https://github.com/devlopersabbir/tview-cli
brew install --cask tview
```

### Linux

For Ubuntu or Debian, download the `.deb` package from [GitHub Releases](https://github.com/devlopersabbir/tview-cli/releases), then install it:

```bash
sudo apt install ./tview_*_linux_amd64.deb
```

RPM and APK packages are also available from [GitHub Releases](https://github.com/devlopersabbir/tview-cli/releases).

### Windows

Download the Windows `.zip` from [GitHub Releases](https://github.com/devlopersabbir/tview-cli/releases), or install with Scoop:

```powershell
scoop bucket add devlopersabbir-tview https://github.com/devlopersabbir/tview-cli
scoop install tview
```

### Go

```bash
go install github.com/devlopersabbir/tview/cmd/tview@latest
```

## Usage

```bash
tview [symbol] [interval]
```

| Argument | Description | Example |
| --- | --- | --- |
| `symbol` | Trading pair. `USDT` is added automatically when omitted. | `btc`, `eth`, `SOLUSDT` |
| `interval` | Candle interval. | `1m`, `5m`, `15m`, `1h`, `4h`, `1d` |

Supported intervals:

```text
1m  3m  5m  15m  30m  1h  2h  4h  6h  12h  1d  1w
```

### Examples

```bash
tview btc 1h
tview eth 15m
tview SOLUSDT 4h
tview ada 1d
```

### Version

```bash
tview version
```

## Author

**Sabbir Hossain Shuvo**

GitHub: [@devlopersabbir](https://github.com/devlopersabbir)

## License

[MIT](LICENSE) © Sabbir Hossain Shuvo
