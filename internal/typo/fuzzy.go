package typo

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// Matches non-alphanumeric characters except basic spaces
	specialCharsRegex = regexp.MustCompile(`[^\p{L}\p{N}\s]`)
	multipleSpaces    = regexp.MustCompile(`\s+`)
)

// CleanQuery strips malicious or syntax-breaking characters from raw search queries.
func CleanQuery(q string) string {
	cleaned := specialCharsRegex.ReplaceAllString(q, " ")
	cleaned = multipleSpaces.ReplaceAllString(cleaned, " ")
	return strings.TrimSpace(cleaned)
}

// Tokenize splits cleaned text into normalized lowercase word tokens.
func Tokenize(text string) []string {
	cleaned := CleanQuery(text)
	if cleaned == "" {
		return nil
	}
	words := strings.FieldsFunc(cleaned, func(r rune) bool {
		return unicode.IsSpace(r)
	})
	var tokens []string
	for _, w := range words {
		lw := strings.ToLower(w)
		if len(lw) > 0 {
			tokens = append(tokens, lw)
		}
	}
	return tokens
}

// BuildFTSQuery constructs a robust SQLite FTS5 MATCH expression.
// Supports prefix matching on the last token (instant search behavior)
// and handles multi-term phrase/conjunction.
func BuildFTSQuery(raw string, prefixLastToken bool) string {
	tokens := Tokenize(raw)
	if len(tokens) == 0 {
		return ""
	}

	var clauses []string
	for i, token := range tokens {
		// Escape token inside double quotes for exact term safety in FTS5
		if prefixLastToken && i == len(tokens)-1 && len(token) >= 2 {
			// Prefix match on last token: "token"*
			clauses = append(clauses, `"`+token+`"*`)
		} else {
			clauses = append(clauses, `"`+token+`"`)
		}
	}

	return strings.Join(clauses, " AND ")
}

// BuildFTSOrQuery builds a lenient OR expression across all tokens (for fuzzy/fallback passes).
func BuildFTSOrQuery(raw string) string {
	tokens := Tokenize(raw)
	if len(tokens) == 0 {
		return ""
	}

	var clauses []string
	for _, token := range tokens {
		if len(token) >= 2 {
			clauses = append(clauses, `"`+token+`"*`)
		} else {
			clauses = append(clauses, `"`+token+`"`)
		}
	}

	return strings.Join(clauses, " OR ")
}
