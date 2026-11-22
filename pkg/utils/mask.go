package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Sensitive field names (case-insensitive) whose values are masked in logs.
var sensitiveFields = []string{
	"password", "token", "secret", "authorization", "refresh_token",
	"access_token", "api_key", "cookie", "jwt", "bearer", "email",
}

const maskedValue = "***MASKED***"

// MaskEmail redacts most of the local part for safe logging.
func MaskEmail(email string) string {
	email = strings.TrimSpace(email)
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return maskedValue
	}
	local := email[:at]
	domain := email[at+1:]
	if len(local) == 0 {
		return maskedValue
	}
	if len(local) == 1 {
		return "*" + "@" + domain
	}
	return local[:1] + "***@" + domain
}

// MaskSensitiveJSON masks known sensitive keys in a JSON string.
// Used for request/response body before logging.
func MaskSensitiveJSON(raw string) string {
	if raw == "" {
		return ""
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "{}" {
		return trimmed
	}
	if !json.Valid([]byte(raw)) {
		return raw
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw
	}
	maskValue(v)
	b, _ := json.Marshal(v)
	return string(b)
}

// MaskQueryString masks sensitive key=value pairs in a query string for logging.
func MaskQueryString(q string) string {
	if q == "" {
		return ""
	}
	return maskQueryString(q)
}

func maskValue(v any) {
	switch val := v.(type) {
	case map[string]any:
		maskMap(val)
	case []any:
		for i := range val {
			maskValue(val[i])
		}
	}
}

func maskMap(m map[string]any) {
	for k, v := range m {
		if isSensitiveKey(k) {
			if strings.EqualFold(k, "email") {
				if s, ok := v.(string); ok {
					m[k] = MaskEmail(s)
					continue
				}
			}
			m[k] = maskedValue
			continue
		}
		maskValue(v)
	}
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, s := range sensitiveFields {
		if lower == s || strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// maskQueryString masks key=value pairs where key is sensitive.
func maskQueryString(q string) string {
	parts := strings.Split(q, "&")
	for i, part := range parts {
		idx := strings.Index(part, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(part[:idx])
		if isSensitiveKey(key) {
			parts[i] = key + "=" + maskedValue
		}
	}
	return strings.Join(parts, "&")
}

// FormatNonJSONBodyLog returns a safe placeholder for non-JSON bodies in logs.
func FormatNonJSONBodyLog(contentType string, size int) string {
	if contentType == "" {
		contentType = "unknown"
	}
	return fmt.Sprintf("<non-json content-type=%q size=%d>", contentType, size)
}

// TruncateForLog limits logged payload size.
func TruncateForLog(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes] + "...(truncated)"
}

// MaskBodyForLog returns a safe string for logging request/response body.
func MaskBodyForLog(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	return MaskSensitiveJSON(string(body))
}