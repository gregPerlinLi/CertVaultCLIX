# CertVaultCLIX (cvx)

[![Release](https://img.shields.io/github/v/release/gregPerlinLi/CertVaultCLIX)](https://github.com/gregPerlinLi/CertVaultCLIX/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/gregPerlinLi/CertVaultCLIX)](https://goreportcard.com/report/github.com/gregPerlinLi/CertVaultCLIX)
[![License](https://img.shields.io/github/license/gregPerlinLi/CertVaultCLIX)](LICENSE)

**CertVaultCLIX** (`cvx`) is a modern interactive Terminal UI (TUI) client for the [CertVault](https://github.com/gregPerlinLi/CertVault) self-signed SSL certificate management platform.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), it provides a rich, keyboard-driven terminal interface for managing CA certificates, SSL certificates, users, and sessions — all from your terminal.

## Features

- 🖥 **Full interactive TUI** — keyboard-driven, navigable menus and tables
- 🔐 **CA Certificate management** — list, view, and export CA certificates
- 📜 **SSL Certificate management** — list, view, request, renew, delete SSL certificates
- 👤 **Profile management** — view and update your profile, change password
- 📋 **Session management** — list and revoke sessions
- 🛠 **Certificate tools** — analyze certificates/keys, convert formats (PEM↔DER, PEM→PFX)
- ⚙️ **Admin panel** — manage users and CA certificates (admin role)
- 👑 **Superadmin panel** — manage all users and force logout (superadmin role)
- 🎨 **Color-coded expiry** — green/yellow/red for certificate expiry status
- ⚡ **Fast and lightweight** — single binary, no dependencies

## Installation

### Go Install

```bash
go install github.com/gregPerlinLi/CertVaultCLIX@latest
```

### From Releases

Download the latest binary from the [Releases](https://github.com/gregPerlinLi/CertVaultCLIX/releases) page.

### Build from Source

```bash
git clone https://github.com/gregPerlinLi/CertVaultCLIX.git
cd CertVaultCLIX
make build
# Binary: ./cvx
make install  # installs to $GOPATH/bin
```

## Usage

### Interactive TUI Mode (default)

```bash
cvx
```

Launches the full-screen interactive TUI. If not logged in, you'll see the login screen.

### Direct Commands

```bash
# Check server connectivity
cvx ping

# Print version
cvx version

# Use a different server URL
cvx --server http://my-certvault:1888
```

## Configuration

Configuration is stored at `~/.config/certvaultclix/config.json`:

```json
{
  "server_url": "http://localhost:1888",
  "session": "JSESSIONID_VALUE"
}
```

### Environment Variables

| Variable | Description |
|---|---|
| `CERTVAULT_URL` | CertVault server URL (overrides config) |
| `CERTVAULT_SESSION` | JSESSIONID session cookie value |

## Keyboard Shortcuts

| Key | Action |
|---|---|
| `↑`/`k`, `↓`/`j` | Move up/down |
| `←`/`h`, `→`/`l` | Navigate panels |
| `Enter` | Select / confirm |
| `Esc` | Back / cancel |
| `r` / `F5` | Refresh |
| `/` | Search |
| `n` | New item |
| `d` | Delete item |
| `e` | Edit item |
| `x` | Export |
| `Tab` | Next field |
| `?` | Toggle help overlay |
| `q` | Quit |

## Navigation Structure

```
📊 Dashboard
├── 🔐 CA Certificates      (list/view)
├── 📜 SSL Certificates     (list/view/request/renew/delete)
├── ➕ Request Certificate  (form)
├── 👤 Profile              (view/update/change password)
├── 📋 Sessions             (list/revoke)
├── 🛠 Tools
│   ├── Analyze Certificate
│   ├── Analyze Private Key
│   └── Convert Formats (PEM↔DER, PEM→PFX)
├── ⚙️ Admin (role >= Admin)
│   ├── User Management
│   └── CA Management
├── 👑 Superadmin (role == Superadmin)
│   ├── All Sessions
│   └── User Management
└── ⚡ Settings
    ├── Server URL
    └── About
```

## Visual Design

- **Color scheme:** Purple (#7C3AED) primary matching CertVault branding
- **Certificate expiry colors:**
  - 🟢 Green: >30 days remaining
  - 🟡 Yellow: <30 days remaining
  - 🔴 Red: Expired
- **Responsive layout** adapting to terminal size
- **Unicode box drawing** for borders and panels

## Requirements

- Go 1.22+
- A running [CertVault](https://github.com/gregPerlinLi/CertVault) server

## License

Apache 2.0 — see [LICENSE](LICENSE)
