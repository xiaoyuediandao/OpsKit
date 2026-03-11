package security

import (
	"regexp"
	"strings"
)

var redactPatterns []*regexp.Regexp

func init() {
	patterns := []string{
		// UUID-format API keys (e.g., 69979f2a-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
		`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
		// Endpoint IDs (ep-m-xxx or ep-xxx)
		`ep-[a-zA-Z0-9][\w-]{6,}`,
		// Feishu App IDs (cli_xxx)
		`cli_[a-zA-Z0-9]{10,}`,
		// User IDs (ou_xxx)
		`ou_[a-zA-Z0-9]{10,}`,
		// Secret keys (sk-xxx or sk_xxx)
		`sk[-_][a-zA-Z0-9]{10,}`,
		// Generic key=value patterns for sensitive fields
		`(?i)("(?:apiKey|appSecret|app_secret|token|api_key|secret_key|aes_key|encoding_key|password)")\s*:\s*"[^"]{8,}"`,
		// key_xxx patterns
		`key_[a-zA-Z0-9]{10,}`,
		// token_xxx patterns
		`token_[a-zA-Z0-9]{10,}`,
	}
	redactPatterns = make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		redactPatterns = append(redactPatterns, regexp.MustCompile(p))
	}
}

// Redact replaces sensitive patterns in the input string with [REDACTED].
func Redact(s string) string {
	result := s
	for _, re := range redactPatterns {
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			// For key-value patterns, only redact the value part
			if strings.Contains(match, ":") && strings.HasPrefix(match, "\"") {
				// Find the last quoted value and redact it
				idx := strings.LastIndex(match, "\"")
				if idx > 0 {
					prefix := match[:strings.Index(match, ": \"")+3]
					return prefix + "[REDACTED]\""
				}
			}
			return "[REDACTED]"
		})
	}
	return result
}
