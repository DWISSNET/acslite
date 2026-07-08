package pkg

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NewUUID generates a new random UUID string.
func NewUUID() string {
	return uuid.New().String()
}

// ExtractMACFromSerial tries to derive a MAC address from the last 12 hex
// characters of a serial number.
func ExtractMACFromSerial(serial string) string {
	s := strings.ToUpper(serial)
	if len(s) >= 12 {
		raw := s[len(s)-12:]
		parts := make([]string, 6)
		for i := 0; i < 6; i++ {
			parts[i] = raw[i*2 : i*2+2]
		}
		return strings.Join(parts, ":")
	}
	return "00:00:00:00:00:00"
}

// FormatTime returns a consistent RFC3339 string for use in responses.
func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// TruncateString limits a string to n characters.
func TruncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// BuildDeviceID constructs a unique device ID from manufacturer OUI and serial.
func BuildDeviceID(oui, serial string) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(oui), strings.ToUpper(serial))
}

// DetectVendorFromManufacturer maps a raw manufacturer string to a normalised
// vendor key (huawei, zte, tplink, dlink, mikrotik, etc.).
func DetectVendorFromManufacturer(manufacturer string) string {
	lower := strings.ToLower(manufacturer)
	switch {
	case strings.Contains(lower, "huawei"):
		return "huawei"
	case strings.Contains(lower, "zte"):
		return "zte"
	case strings.Contains(lower, "tp-link") || strings.Contains(lower, "tplink"):
		return "tplink"
	case strings.Contains(lower, "d-link") || strings.Contains(lower, "dlink"):
		return "dlink"
	case strings.Contains(lower, "mikrotik"):
		return "mikrotik"
	case strings.Contains(lower, "cisco"):
		return "cisco"
	case strings.Contains(lower, "calix"):
		return "calix"
	case strings.Contains(lower, "technicolor"):
		return "technicolor"
	default:
		return strings.ToLower(manufacturer)
	}
}
