package views

import (
	"errors"
	"strings"
	"time"

	"github.com/gregPerlinLi/CertVaultCLIX/internal/api"
)

// inlineAnalysisMsg is sent when an inline cert analysis (from a detail view) completes.
type inlineAnalysisMsg struct {
	result string
	err    string
}

// wrapText wraps a plain-text string so that each output line is at most maxWidth
// rune-columns wide. Continuation lines are prefixed with indent.
func wrapText(text string, maxWidth int, indent string) string {
	if maxWidth <= 0 {
		return text
	}
	runes := []rune(text)
	if len(runes) <= maxWidth {
		return text
	}
	indentRunes := []rune(indent)
	// Continuation lines may be shorter due to indent; ensure minimum usable width.
	contWidth := maxWidth - len(indentRunes)
	if contWidth < 10 {
		contWidth = maxWidth // indent too wide: fall back to maxWidth, no extra indent
	}

	breakAt := func(r []rune, width int) int {
		if len(r) <= width {
			return len(r)
		}
		// Search backwards from width-1 for a space to break on a word boundary.
		w := width - 1
		for w > 0 && r[w] != ' ' {
			w--
		}
		if w == 0 {
			return width // hard break: no space found within width
		}
		return w
	}

	var b strings.Builder
	// First line.
	cut := breakAt(runes, maxWidth)
	b.WriteString(string(runes[:cut]))
	rest := runes[cut:]
	// Skip leading spaces on the remainder.
	for len(rest) > 0 && rest[0] == ' ' {
		rest = rest[1:]
	}

	for len(rest) > 0 {
		b.WriteByte('\n')
		b.WriteString(string(indentRunes))
		cut = breakAt(rest, contWidth)
		b.WriteString(string(rest[:cut]))
		rest = rest[cut:]
		for len(rest) > 0 && rest[0] == ' ' {
			rest = rest[1:]
		}
	}
	return b.String()
}

// SessionExpiredMsg is sent when any API call returns 401 Unauthorized.
type SessionExpiredMsg struct{}

// LoggedOutMsg is sent after an explicit user-initiated logout.
type LoggedOutMsg struct{}

// isUnauthorized returns true when err indicates a session expiry (HTTP 401).
func isUnauthorized(err error) bool {
	return err != nil && errors.Is(err, api.ErrUnauthorized)
}

// parseDaysLeft parses a date string and returns the number of days until expiry.
// Handles multiple date formats used by the CertVault API.
func parseDaysLeft(notAfter string) int {
for _, f := range dateFormats {
if t, err := time.Parse(f, notAfter); err == nil {
return int(time.Until(t).Hours() / 24)
}
}
return 0
}

// formatNotAfter parses a date string and returns it formatted as YYYY-MM-DD.
func formatNotAfter(notAfter string) string {
for _, f := range dateFormats {
if t, err := time.Parse(f, notAfter); err == nil {
return t.Format("2006-01-02")
}
}
return notAfter
}

// certAlgos is the ordered list of key algorithms supported when requesting
// CA and SSL certificates. It is shared by CARequest and CertRequest.
var certAlgos = []string{"RSA", "EC", "ED25519"}

// dateFormats lists date format strings accepted by the API, in priority order.
var dateFormats = []string{
time.RFC3339,
"2006-01-02T15:04:05",
"2006-01-02T15:04:05.999",
"2006-01-02T15:04:05.999999999",
}
