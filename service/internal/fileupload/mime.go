package fileupload

import (
	"mime"
	"strings"
)

func normalizeMediaType(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	mt, _, err := mime.ParseMediaType(s)
	if err != nil {
		if idx := strings.Index(s, ";"); idx >= 0 {
			return strings.ToLower(strings.TrimSpace(s[:idx]))
		}
		return strings.ToLower(s)
	}
	return strings.ToLower(mt)
}

func mimeAllowed(detected string, allowed []string) bool {
	d := normalizeMediaType(detected)
	if d == "" {
		return false
	}
	for _, raw := range allowed {
		if mimeRuleMatches(d, raw) {
			return true
		}
	}
	return false
}

func mimeRuleMatches(detectedNormalized, raw string) bool {
	a := strings.TrimSpace(strings.ToLower(raw))
	if a == "" {
		return false
	}
	if !strings.HasSuffix(a, "/*") {
		rule := normalizeMediaType(raw)
		return rule != "" && rule == detectedNormalized
	}
	prefix := strings.TrimSuffix(a, "/*")
	return strings.HasPrefix(detectedNormalized, prefix+"/")
}
