package index

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alexandrmotologa/searchcraft/internal/storage"
	"github.com/alexandrmotologa/searchcraft/internal/typo"
)

// HighlightText wraps case-insensitive occurrences of query tokens inside <mark> tags.
func HighlightText(text string, queryTokens []string) string {
	if text == "" || len(queryTokens) == 0 {
		return text
	}

	highlighted := text
	for _, token := range queryTokens {
		token = strings.TrimSpace(token)
		if len(token) < 2 {
			continue
		}
		// Match word boundaries or substring
		pattern := fmt.Sprintf(`(?i)(%s)`, regexp.QuoteMeta(token))
		re, err := regexp.Compile(pattern)
		if err == nil {
			highlighted = re.ReplaceAllString(highlighted, "<mark>$1</mark>")
		}
	}
	return highlighted
}

// GenerateHighlights creates snippets with <mark> tags for textual fields.
func GenerateHighlights(docData map[string]any, queryTokens []string) map[string]string {
	highlights := make(map[string]string)
	targetFields := []string{"title", "name", "description", "content", "summary"}

	for _, field := range targetFields {
		val, exists := docData[field]
		if !exists || val == nil {
			continue
		}
		strVal, ok := val.(string)
		if !ok || strVal == "" {
			continue
		}

		hl := HighlightText(strVal, queryTokens)
		if strings.Contains(hl, "<mark>") {
			highlights[field] = hl
		}
	}

	return highlights
}

// BuildFTSQueryWithSynonyms constructs an FTS5 MATCH expression expanding query terms with active synonyms.
func BuildFTSQueryWithSynonyms(rawQuery string, synMap map[string][]string, prefixLast bool) string {
	tokens := typo.Tokenize(rawQuery)
	if len(tokens) == 0 {
		return ""
	}

	var termClauses []string
	for i, token := range tokens {
		synonyms := synMap[token]
		if len(synonyms) == 0 {
			if prefixLast && i == len(tokens)-1 && len(token) >= 2 {
				termClauses = append(termClauses, `"`+token+`"*`)
			} else {
				termClauses = append(termClauses, `"`+token+`"`)
			}
			continue
		}

		// Expand term with synonyms: ("term" OR "synonym1" OR "synonym2")
		var orList []string
		orList = append(orList, `"`+token+`"`)
		for _, s := range synonyms {
			cleanSyn := typo.CleanQuery(s)
			if cleanSyn != "" {
				orList = append(orList, `"`+cleanSyn+`"`)
			}
		}
		termClauses = append(termClauses, "("+strings.Join(orList, " OR ")+")")
	}

	return strings.Join(termClauses, " AND ")
}

// ExecuteBM25 runs full-text search against the SQLite FTS5 table with fallback on typos.
func ExecuteBM25(db *storage.DB, indexName, rawQuery string, synMap map[string][]string, limit int) ([]storage.FTSResult, error) {
	tokens := typo.Tokenize(rawQuery)
	if len(tokens) == 0 {
		return nil, nil
	}

	// 1. Try exact & prefix query with synonyms:
	ftsQuery := BuildFTSQueryWithSynonyms(rawQuery, synMap, true)
	results, err := db.QueryFTS(indexName, ftsQuery, limit)
	if err == nil && len(results) > 0 {
		return normalizeBM25Scores(results), nil
	}

	// 2. Fallback to lenient OR query if few or no results found
	orQuery := typo.BuildFTSOrQuery(rawQuery)
	if orQuery != "" && orQuery != ftsQuery {
		results, err = db.QueryFTS(indexName, orQuery, limit)
		if err == nil && len(results) > 0 {
			return normalizeBM25Scores(results), nil
		}
	}

	return nil, nil
}

// normalizeBM25Scores scales raw positive scores to [0.0, 1.0] range relative to the top hit.
func normalizeBM25Scores(results []storage.FTSResult) []storage.FTSResult {
	if len(results) == 0 {
		return results
	}

	maxScore := results[0].BM25Score
	if maxScore <= 0 {
		for i := range results {
			results[i].BM25Score = 1.0
		}
		return results
	}

	for i := range results {
		results[i].BM25Score = results[i].BM25Score / maxScore
	}

	return results
}
