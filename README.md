# CertVaultCLIX (cvx)

[![Release](https://img.shields.io/github/v/release/gregPerlinLi/CertVaultCLIX)](https://github.com/gregPerlinLi/CertVaultCLIX/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/gregPerlinLi/CertVaultCLIX)](https://goreportcard.com/report/github.com/gregPerlinLi/CertVaultCLIX)
[![License](https://img.shields.io/github/license/gregPerlinLi/CertVaultCLIX)](LICENSE)

**CertVaultCLIX** (`cvx`) is a modern, keyboard-driven interactive Terminal UI (TUI) client for the
[CertVault](https://github.com/gregPerlinLi/CertVault) self-signed SSL certificate management platform.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Cobra](https://github.com/spf13/cobra),
it brings the full power of CertVault to your terminal — no browser required.
Manage CA certificates, issue and renew SSL certificates, inspect sessions, convert key formats, and administer
users, all through a rich, color-coded TUI that adapts to your terminal size.

---

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
  - [Homebrew (macOS / Linux)](#homebrew-macos--linux)
  - [Scoop (Windows)](#scoop-windows)
  - [Download Pre-built Binary](#download-pre-built-binary)
  - [Go Install](#go-install)
  - [Build from Source](#build-from-source)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
  - [Config File](#config-file)
  - [Environment Variables](#environment-variables)
  - [Command-line Flag](#command-line-flag)
- [CLI Reference](#cli-reference)
- [TUI Usage Guide](#tui-usage-guide)
  - [Login](#login)
  - [Layout Overview](#layout-overview)
  - [Navigation Structure](#navigation-structure)
  - [Dashboard](#dashboard)
  - [CA Certificates](#ca-certificates)
  - [SSL Certificates](#ssl-certificates)
  - [Request a Certificate](#request-a-certificate)
  - [Profile](#profile)
  - [Sessions](#sessions)
  - [Tools](#tools)
  - [Admin Panel](#admin-panel)
  - [Superadmin Panel](#superadmin-panel)
  - [Settings](#settings)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [Role-based Access Control](#role-based-access-control)
- [Visual Design](#visual-design)
- [License](#license)

---

## Features

| Category | Capabilities |
|---|---|
| 🖥 **Interactive TUI** | Full-screen, keyboard-driven menus, tables, and forms |
| 🔐 **CA Certificates** | List, view details, export PEM/DER, request, renew, delete (admin) |
| 📜 **SSL Certificates** | List, view details, export PEM/DER/PFX, request, renew, delete |
| 👤 **Profile** | View current profile, update display name / email, change password |
| 📋 **Sessions** | List active sessions, view details, revoke individual sessions |
| 🛠 **Certificate Tools** | Analyze certificates/private keys, convert PEM↔DER |
| ⚙️ **Admin Panel** | User management, CA certificate management (Admin role+) |
| 👑 **Superadmin Panel** | Manage all users, view all sessions, force-logout any user (Superadmin only) |
| 🎨 **Color-coded Expiry** | Green / yellow / red for certificate validity status |
| ⚡ **Lightweight** | Single static binary, no runtime dependencies |

---

## Requirements

- A running [CertVault](https://github.com/gregPerlinLi/CertVault) server instance
- Go 1.22+ (only required for building from source)

---

## Installation

### Homebrew (macOS / Linux)

```bash
brew install gregPerlinLi/tap/cvx
```

### Scoop (Windows)

```powershell
scoop bucket add gregPerlinLi https://github.com/gregPerlinLi/scoop-bucket
scoop install cvx
```

### Download Pre-built Binary

Pre-built binaries for **Linux**, **macOS**, and **Windows** on **AMD64** and **ARM64** are available on the
[Releases](https://github.com/gregPerlinLi/CertVaultCLIX/releases) page.

```bash
# Example: Linux AMD64
curl -Lo cvx.tar.gz https://github.com/gregPerlinLi/CertVaultCLIX/releases/latest/download/cvx_<version>_linux_amd64.tar.gz
tar -xzf cvx.tar.gz
chmod +x cvx
sudo mv cvx /usr/local/bin/
```

Each release also ships a `cvx_<version>_checksums.txt` (SHA-256) file for verification.

### Go Install

Requires Go 1.22+:

```bash
go install github.com/gregPerlinLi/CertVaultCLIX@latest
```

The `cvx` binary will be placed in `$GOPATH/bin` (usually `~/go/bin`).
Make sure that directory is on your `$PATH`.

### Build from Source

```bash
git clone https://github.com/gregPerlinLi/CertVaultCLIX.git
cd CertVaultCLIX

# Build binary to ./cvx
make build

# Or install directly to $GOPATH/bin
make install
```

To embed version metadata:

```bash
make build VERSION=v1.2.3
```

---

## Quick Start

1. **Start a CertVault server** (see [CertVault](https://github.com/gregPerlinLi/CertVault) for setup).
   The default address is `http://localhost:1888`.

2. **Launch cvx:**

   ```bash
   cvx
   ```

3. At the **Login** screen, enter your CertVault username and password, then press `Enter`.

4. After a successful login, you land on the **Dashboard**.
   Use `↑`/`↓` (or `j`/`k`) in the left sidebar to navigate between sections and press `Enter` to open one.

5. Press `q` (or `Ctrl+C`) at any time to quit. Your session cookie is saved automatically.

---

## Configuration

cvx stores its configuration in a JSON file and also reads environment variables. The priority order is:

```
CLI flag  >  Environment variable  >  Config file  >  Default
```

### Config File

| Platform | Default path |
|---|---|
| Linux / BSD | `~/.config/certvaultclix/config.json` |
| macOS | `~/Library/Application Support/certvaultclix/config.json` |
| Windows | `%APPDATA%\certvaultclix\config.json` |

Example content:

```json
{
  "server_url": "http://localhost:1888",
  "session": "JSESSIONID_VALUE"
}
```

The `session` field is updated automatically whenever you log in or out of the TUI.
You can also set it manually to reuse an existing CertVault session token.

### Environment Variables

| Variable | Description |
|---|---|
| `CERTVAULT_URL` | CertVault server base URL — overrides the config file value |
| `CERTVAULT_SESSION` | JSESSIONID cookie value — overrides the config file value |

Example:

```bash
export CERTVAULT_URL=https://certvault.mycompany.com
export CERTVAULT_SESSION=abc123...
cvx
```

### Command-line Flag

```bash
# Override the server URL for a single invocation
cvx --server http://certvault-staging:1888
```

---

## CLI Reference

| Command | Description |
|---|---|
| `cvx` | Launch the interactive TUI (default) |
| `cvx ping` | Check connectivity to the CertVault server and print confirmation |
| `cvx version` | Print version, commit hash, and build date |
| `cvx --server <url>` | Use a different server URL for this invocation |
| `cvx --help` | Show help |

**Examples:**

```bash
# Verify that the server is reachable
cvx ping
# ✓ Connected to http://localhost:1888

# Print full version info
cvx version
# v1.2.3 (commit: a1b2c3d, built: 2025-01-15T10:30:00Z)

# Start TUI against a specific server
cvx --server https://certvault.example.com
```

---

## TUI Usage Guide

### Login

If cvx does not have a stored session (or the session has expired), the **Login** screen appears first.

```
Username:  _______________
Password:  _______________

                    [ Login ]
```

- Use `Tab` / `Shift+Tab` to move between fields.
- Press `Enter` on the **Login** button (or when the last field is focused) to submit.
- An error banner appears if credentials are invalid.

### Layout Overview

The TUI is split into two panes:

```
┌─────────────────┬─────────────────────────────────────┐
│  Sidebar        │  Content pane                       │
│                 │                                     │
│  📊 Dashboard   │  (current section is rendered here) │
│  🔐 CA Certs    │                                     │
│  📜 SSL Certs   │                                     │
│  ➕ Request     │                                     │
│  👤 Profile     │                                     │
│  📋 Sessions    │                                     │
│  🛠 Tools       │                                     │
│  ⚙️  Admin      │                                     │
│  👑 Superadmin  │                                     │
│  ⚡ Settings    │                                     │
└─────────────────┴─────────────────────────────────────┘
```

- `←`/`h` and `→`/`l` switch focus between the **sidebar** and the **content pane**.
- `↑`/`k` and `↓`/`j` navigate items within the focused pane.
- `Enter` opens the highlighted section or row.
- `Esc` goes back one level (from content pane → sidebar, from detail → list, etc.).

### Navigation Structure

```
📊 Dashboard
├── 🔐 CA Certificates      list → detail → export PEM / export DER
├── 📜 SSL Certificates     list → detail → export PEM / export DER / export PFX
├── ➕ Request Certificate  form (CA, CN, SANs, algorithm, expiry…)
├── 👤 Profile              view + edit display name / email / password
├── 📋 Sessions             list → detail → revoke
├── 🛠 Tools
│   ├── Analyze Certificate  (paste PEM → parsed fields)
│   ├── Analyze Private Key  (paste PEM key → algorithm / key size)
│   ├── Convert PEM → DER
│   └── Convert DER → PEM
├── ⚙️  Admin  [role ≥ Admin]
│   ├── User Management      list users
│   └── CA Management        list CAs → detail
├── 👑 Superadmin  [role = Superadmin]
│   ├── All Sessions         list all sessions → detail → force-logout
│   └── User Management      list / create / edit / delete users
└── ⚡ Settings
    ├── Server URL           edit + save
    └── About                version info
```

### Dashboard

The Dashboard is the home screen after login.
It shows a welcome message, your role, and **Quick Stats** cards.
The set of stats shown depends on your role:

| Stat | Visible to |
|---|---|
| Binded CA | Everyone |
| Requested SSL Certs | Everyone |
| Total Users | Admin+ |
| Requested CA Certs | Admin+ |
| Total CA Certs | Superadmin |
| Total SSL Certs | Superadmin |

Press `r` to refresh the stats.

### CA Certificates

Lists the CA certificates bound to your account.
Each row shows the comment, owner, type (Root CA / Int CA / Leaf CA), expiry date, remaining days, and availability.

| Key | Action |
|---|---|
| `↑`/`↓` | Select a row |
| `Enter` | View certificate details |
| `r` / `F5` | Refresh the list |
| `[` / `]` | Previous / next page |

**Detail view** shows all certificate fields (subject, issuer, validity, algorithm, fingerprint, SANs, extensions).
- Press `x` to export the certificate (PEM or DER).
- Press `Esc` to return to the list.

### SSL Certificates

Lists the SSL certificates you own.
Each row shows the comment, owner, expiry date, and remaining days color-coded by validity.

| Key | Action |
|---|---|
| `↑`/`↓` | Select a row |
| `Enter` | View certificate details |
| `n` | Request a new certificate (opens the Request form) |
| `d` | Delete the selected certificate (confirmation required) |
| `r` / `F5` | Refresh the list |
| `[` / `]` | Previous / next page |

**Detail view** shows all fields plus private key retrieval.
- Press `x` to export (PEM / DER / PFX).
- Press `e` to renew (extend the expiry).
- Press `Esc` to return to the list.

### Request a Certificate

A scrollable form for issuing a new SSL certificate.

| Field | Description |
|---|---|
| **CA** | Select the signing CA with `↑`/`↓` |
| **Common Name (CN)** | Primary domain or identifier (required) |
| **Country** | Two-letter ISO country code (e.g. `US`) |
| **Province** | State or province name (e.g. `California`) |
| **City** | Locality name (e.g. `San Francisco`) |
| **Organization** | Legal organization name |
| **SANs** | Comma-separated Subject Alternative Names (e.g. `example.com,*.example.com`) |
| **Algorithm** | Signing algorithm — use `↑`/`↓` to cycle: `RSA`, `EC`, `ED25519` |
| **Key Size** | `2048` or `4096` for RSA; `256` or `384` for EC; leave blank for ED25519 |
| **Expire Days** | Validity period in days (e.g. `365`) |
| **Comment** | Optional note shown in the certificate list |

Navigation inside the form:
- `Tab` / `Shift+Tab` move between fields.
- `↑`/`↓` cycle options on the **CA** and **Algorithm** selectors.
- `Enter` on the last field submits the request.
- Mouse wheel scrolls the form when it doesn't fit in the terminal.

### Profile

Shows your current username and role, with a form to update:

| Field | Notes |
|---|---|
| **Display Name** | Visible name used in the UI |
| **Email** | Contact email on file |
| **Old Password** | Required only when changing the password |
| **New Password** | Leave blank to keep the current password |

Press `Tab` to move between fields and `Enter` on the last field to save.
A success toast confirms the update; error details are shown in red on failure.

### Sessions

Lists your login sessions (including sessions from other devices / browsers).

| Column | Description |
|---|---|
| UUID | Unique session identifier |
| IP Address | Client IP at login time |
| Browser / OS | User-agent parsed info |
| Login At | Timestamp of the login |
| Online | `✓` if the session is currently active |

| Key | Action |
|---|---|
| `↑`/`↓` | Select a session |
| `Enter` | View session details (IP, region, city, browser, OS) |
| `d` / `Delete` | Revoke the selected session (confirmation dialog) |
| `r` / `F5` | Refresh |
| `[` / `]` | Previous / next page |

### Tools

A sub-menu with four utilities.
Use `↑`/`↓` and `Enter` to open a tool, `Esc` to return to the menu.

#### Analyze Certificate

Paste a PEM-encoded certificate into the text area and press `Ctrl+S` to analyze it.
The result pane shows the subject, issuer, validity period (color-coded), serial number, fingerprint,
public key details, and extensions / SANs.

#### Analyze Private Key

Paste a PEM-encoded private key and press `Ctrl+S` to analyze.
Returns the algorithm (RSA / EC / ED25519) and key size in bits.

#### Convert PEM → DER / DER → PEM

Paste the input and press `Ctrl+S`.
The converted output is shown in the result pane.

**Common tool keys:**

| Key | Action |
|---|---|
| `Ctrl+S` | Run the tool |
| `Tab` | Toggle focus: input ↔ result |
| `↑`/`↓` (result focused) | Scroll the result |
| `Ctrl+L` | Clear input and result |
| `Esc` | Return to tools menu |

### Admin Panel

Visible to users with the **Admin** or **Superadmin** role.
A top-level menu offers two sub-sections:

#### User Management

Tabular list of all users with columns: username, display name, email, role.
- `[` / `]` to page through users.
- `r` to refresh.

#### CA Management

Tabular list of all CA certificates (Root CA, Int CA, Leaf CA) with availability status.
- `Enter` on a row opens the full **CA Detail** view (same as the user CA detail view).
- `r` to refresh; `[` / `]` to page.

### Superadmin Panel

Visible to **Superadmin** users only.
Offers two sub-sections:

#### All Sessions

Displays every active session across all users with the same columns as the personal Sessions view,
plus the owning **username**.
- `d` / `Delete` force-logs out the selected session.

#### User Management

Full CRUD for all user accounts:
- `n` — create a new user (username, display name, email, password, role).
- `e` — edit the selected user (display name, email, password).
- `d` — delete the selected user (confirmation required).
- Role can be set to **User (1)**, **Admin (2)**, or **Superadmin (3)**.

### Settings

| Section | Description |
|---|---|
| **Server URL** | The CertVault base URL used by cvx |
| **Config File** | Absolute path of the on-disk config file |
| **Version** | Build version, commit hash, and build date |
| **GitHub** | Link to the source repository |

Press `e` to edit the Server URL, `Enter` to save, `Esc` to cancel.
Changes are written to the config file immediately.

---

## Keyboard Shortcuts

### Global

| Key | Action |
|---|---|
| `q` / `Ctrl+C` | Quit cvx |
| `?` | Toggle help overlay |
| `←`/`h` | Focus the sidebar |
| `→`/`l` | Focus the content pane |
| `↑`/`k` | Move selection up |
| `↓`/`j` | Move selection down |
| `PgUp` / `Ctrl+U` | Page up |
| `PgDn` / `Ctrl+D` | Page down |
| `Enter` | Select / confirm |
| `Esc` / `Backspace` | Go back / cancel |
| `r` / `F5` | Refresh current view |

### Lists and Tables

| Key | Action |
|---|---|
| `n` | New item |
| `d` / `Delete` | Delete selected item |
| `e` | Edit selected item |
| `x` | Export selected item |
| `/` | Search / filter |
| `[` | Previous page |
| `]` | Next page |

### Forms

| Key | Action |
|---|---|
| `Tab` | Move to next field |
| `Shift+Tab` | Move to previous field |
| `Enter` (last field) | Submit form |
| `Esc` | Cancel and go back |

### Tools

| Key | Action |
|---|---|
| `Ctrl+S` | Execute the selected tool |
| `Tab` | Toggle focus between input and result |
| `Ctrl+L` | Clear input and result |
| `Esc` | Back to tools menu |

---

## Role-based Access Control

CertVault uses three role levels.
The sidebar automatically shows or hides sections based on the logged-in user's role.

| Role | Value | Access |
|---|---|---|
| **User** | 1 | Dashboard, CA Certs, SSL Certs, Request Cert, Profile, Sessions, Tools, Settings |
| **Admin** | 2 | Everything above **+** Admin panel (user list, CA management) |
| **Superadmin** | 3 | Everything above **+** Superadmin panel (all sessions, full user CRUD) |

The Dashboard stats cards also expand as your role level increases (see [Dashboard](#dashboard)).

---

## Visual Design

- **Primary color:** Purple `#7C3AED` — matches CertVault branding
- **Certificate expiry color coding:**
  - 🟢 **Green** — more than 30 days remaining
  - 🟡 **Yellow** — 30 days or fewer remaining
  - 🔴 **Red** — certificate has expired
- **Role color coding in user tables:**
  - Regular users are displayed in the default color
  - Admin and Superadmin roles are highlighted
- **Responsive layout** — panels and tables adapt to the terminal window size
- **Unicode box-drawing characters** for borders and panel separators
- **Animated spinner** during API calls

---

## Building from Source

```bash
# Clone
git clone https://github.com/gregPerlinLi/CertVaultCLIX.git
cd CertVaultCLIX

# Build (outputs ./cvx)
make build

# Install to $GOPATH/bin
make install

# Run tests
make test

# Run linter (requires golangci-lint)
make lint

# Clean build artifacts
make clean
```

To cross-compile for a different platform:

```bash
GOOS=linux GOARCH=arm64 go build -o cvx-linux-arm64 .
GOOS=windows GOARCH=amd64 go build -o cvx-windows-amd64.exe .
GOOS=darwin GOARCH=arm64 go build -o cvx-macos-arm64 .
```

---

## License

Apache 2.0 — see [LICENSE](LICENSE)
