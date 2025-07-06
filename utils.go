package plate

import (
	"strings"
	"unicode"
)

// toSnakeCase converts a CamelCase string to snake_case
func toSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				// Check if previous character is lowercase or next character is lowercase
				// This handles cases like "UserID" -> "user_id" (not "user_i_d")
				prevIsLower := i > 0 && unicode.IsLower(rune(s[i-1]))
				nextIsLower := i < len(s)-1 && unicode.IsLower(rune(s[i+1]))
				
				if prevIsLower || nextIsLower {
					result.WriteByte('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}