# Cursor Bar

A sleek macOS menu bar app that displays your [Cursor](https://cursor.com) usage at a glance.

## Features

- **Live Usage Percentage** — See your current request usage right in the menu bar
- **Request Tracking** — Monitor requests used vs. your plan limit with visual progress bars
- **On-Demand Spending** — Track your on-demand usage costs ($0-$30)
- **Recent Activity** — View your last 10 API calls with model, tokens, and cost
- **Auto-Refresh** — Updates every 5 minutes automatically
- **Dark Mode Support** — Menu bar icon adapts to your system theme
- **Native macOS** — Lightweight, runs as a true menu bar app (no dock icon)

## Screenshot

```
┌─────────────────────────────────────────────┐
│ Request Usage                               │
│   1847/2000 = 92%                           │
│   ●●●●●●●●●●●●●●●●●●○○                      │
│   Resets on Jan 10, 2026                    │
├─────────────────────────────────────────────┤
│ Request Usage (Money)                       │
│   $0.00/$30                                 │
│   ○○○○○○○○○○○○○○○○○○○○                      │
├─────────────────────────────────────────────┤
│ Recent Usage                                │
│   01/04 13:35   Pro   claude-sonnet   2.3K  │
│   01/04 13:20   Pro   claude-sonnet   1.8K  │
│   01/04 12:45   Pro   gemini-pro     15.2K  │
│   ...                                       │
├─────────────────────────────────────────────┤
│ Refresh                                     │
│ Open Dashboard                              │
│ Edit Config                                 │
├─────────────────────────────────────────────┤
│ Quit                                        │
└─────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- macOS 10.14+
- Go 1.21+ (for building from source)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/cursor-mac-bar.git
cd cursor-mac-bar

# Build the app
./build.sh

# Run it
open CursorBar.app
```

## Setup

On first launch, you'll be prompted to enter your Cursor credentials:

### 1. Get your Session Token

1. Open [cursor.com](https://cursor.com) in your browser
2. Open DevTools (`Cmd + Option + I`)
3. Go to **Application** → **Cookies** → `cursor.com`
4. Find `WorkosCursorSessionToken` and copy its value

### 2. Get your User ID

1. In the same DevTools window, go to **Network** tab
2. Refresh the page
3. Look for any API request and find the `user` parameter in the URL
4. It looks like: `user_01ABC...`

### 3. Enter Credentials

When prompted by the app, paste your token and user ID. They'll be saved securely in:
```
~/.config/cursor-bar/config.json
```

## Usage

| Menu Bar Display | Meaning |
|------------------|---------|
| `45%` | Normal usage (45% of limit used) |
| `⚠️92%` | High usage warning (≥90%) |
| `!` | Error (check config or connection) |
| `...` | Loading/refreshing |

### Menu Actions

- **Refresh** — Manually refresh usage data
- **Open Dashboard** — Open Cursor settings in browser
- **Edit Config** — Open config file in default text editor
- **Quit** — Exit the app

## Configuration

Config file location: `~/.config/cursor-bar/config.json`

```json
{
  "token": "your-workos-cursor-session-token",
  "user_id": "user_01ABC..."
}
```

## Development

### Project Structure

```
cursor-mac-bar/
├── main.go                 # App entry point and UI
├── build.sh                # Build script for binary + .app bundle
├── internal/
│   ├── api/api.go          # Cursor API client
│   ├── config/config.go    # Configuration management
│   └── icon/icon.go        # Menu bar icon (embedded PNG)
└── cursor-brand-assets/    # Cursor brand assets (optional)
```

### Building

```bash
# Build Go binary + macOS app bundle
./build.sh

# Output:
# - ./cursor-bar       (CLI binary)
# - ./CursorBar.app    (macOS app bundle)
```

### Regenerating the Menu Bar Icon

If you have the `cursor-brand-assets` folder:

```bash
# The build script automatically uses:
# cursor-brand-assets/General Logos/Cube/PNG/CUBE_2D_DARK.png
```

## API Endpoints Used

| Endpoint | Purpose |
|----------|---------|
| `/api/usage` | Request count and limits |
| `/api/usage-summary` | On-demand spending |
| `/api/dashboard/get-filtered-usage-events` | Recent activity |

## Tech Stack

- **Go** — Core application
- **[systray](https://github.com/getlantern/systray)** — Cross-platform system tray library
- **macOS native** — AppleScript dialogs, `open` command integration

## Troubleshooting

### "!" appears in menu bar
- Your session token may have expired
- Click **Edit Config** and update your token
- Or delete `~/.config/cursor-bar/config.json` and restart

### App doesn't start
- Check if another instance is running: `killall CursorBar`
- Try running from terminal to see errors: `./cursor-bar`

### Usage not updating
- Click **Refresh** to manually update
- Check your internet connection
- Verify your credentials are still valid

## License

MIT

## Acknowledgments

- [Cursor](https://cursor.com) for the amazing AI-powered IDE
- [systray](https://github.com/getlantern/systray) for the menu bar library
