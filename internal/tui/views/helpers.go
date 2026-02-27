package views

import "time"

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

// dateFormats lists date format strings accepted by the API, in priority order.
var dateFormats = []string{
time.RFC3339,
"2006-01-02T15:04:05",
"2006-01-02T15:04:05.999",
"2006-01-02T15:04:05.999999999",
}
