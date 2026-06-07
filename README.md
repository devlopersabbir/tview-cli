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

Recommended with Homebrew:

```bash
brew tap devlopersabbir/tview-cli https://github.com/devlopersabbir/tview-cli
brew install --cask devlopersabbir/tview-cli/tview
tview version
```

If you installed the older formula before, remove it first:

```bash
brew uninstall tview 2>/dev/null || true
brew install --cask devlopersabbir/tview-cli/tview
```

If `brew install` succeeds but `tview` is not found, add Homebrew to your shell PATH:

```bash
# Apple Silicon Macs
eval "$(/opt/homebrew/bin/brew shellenv)"

# Intel Macs
eval "$(/usr/local/bin/brew shellenv)"
```

Manual install without Homebrew:

```bash
curl -fsSL https://raw.githubusercontent.com/devlopersabbir/tview-cli/main/scripts/install.sh | sh
```

The manual installer places `tview` in `/usr/local/bin` when possible, otherwise in `~/.local/bin`, and prints the PATH line to add if needed.

If macOS Gatekeeper blocks a manually downloaded binary, remove the quarantine attribute:

```bash
xattr -d com.apple.quarantine "$(command -v tview)" 2>/dev/null || true
xattr -d com.apple.provenance "$(command -v tview)" 2>/dev/null || true
```

### Linux

Ubuntu or Debian:

```bash
curl -fsSL https://raw.githubusercontent.com/devlopersabbir/tview-cli/main/scripts/install.sh | sh
```

Or download the `.deb` package from [GitHub Releases](https://github.com/devlopersabbir/tview-cli/releases), then install it:

```bash
sudo apt install ./tview_*_linux_amd64.deb
```

RPM and APK packages are also available from [GitHub Releases](https://github.com/devlopersabbir/tview-cli/releases). The installer supports Linux amd64 and arm64.

### Windows

Recommended with PowerShell:

```powershell
iwr https://raw.githubusercontent.com/devlopersabbir/tview-cli/main/scripts/install.ps1 -UseBasicParsing | iex
```

Or install with Scoop:

```powershell
scoop bucket add devlopersabbir-tview https://github.com/devlopersabbir/tview-cli
scoop install tview
```

You can also download the Windows `.zip` from [GitHub Releases](https://github.com/devlopersabbir/tview-cli/releases). Put `tview.exe` in a folder on your PATH, for example `%LOCALAPPDATA%\Programs\tview\bin`.

### Go

```bash
go install github.com/devlopersabbir/tview-cli/cmd/tview@latest
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
