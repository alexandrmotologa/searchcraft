package typo

import (
	"testing"
)

func TestDamerauLevenshtein(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "b", 1},
		{"same", "same", 0},
		// Transposition (adjacent characters swapped)
		{"teh", "the", 1},
		{"keybaord", "keyboard", 1},
		// Insertion
		{"search", "serch", 1},
		// Deletion
		{"mouse", "mouses", 1},
		// Substitution
		{"cat", "bat", 1},
		// Complex multiple edits
		{"rosettacode", "raisethysword", 8},
	}

	for _, tc := range tests {
		dist := DamerauLevenshtein(tc.a, tc.b)
		if dist != tc.expected {
			t.Errorf("DamerauLevenshtein(%q, %q) = %d; expected %d", tc.a, tc.b, dist, tc.expected)
		}
	}
}

func TestTypoToleranceThresholds(t *testing.T) {
	// Length <= 3: 0 typos allowed
	if MaxAllowedDistance(3) != 0 {
		t.Errorf("expected 0 for len 3, got %d", MaxAllowedDistance(3))
	}
	// Length 4-7: 1 typo allowed
	if MaxAllowedDistance(4) != 1 || MaxAllowedDistance(7) != 1 {
		t.Errorf("expected 1 for len 4 and 7")
	}
	// Length >= 8: 2 typos allowed
	if MaxAllowedDistance(8) != 2 || MaxAllowedDistance(12) != 2 {
		t.Errorf("expected 2 for len 8+")
	}

	// Within tolerance checks
	if !IsWithinTypoTolerance("keyboard", "keybaord") {
		t.Errorf("expected 'keybaord' to be within tolerance for 'keyboard'")
	}
	if IsWithinTypoTolerance("cat", "bat") {
		t.Errorf("expected 'bat' NOT to be within tolerance for 3-letter word 'cat'")
	}
}

func TestFTSQueryBuilder(t *testing.T) {
	query := `mechanical  "keyboard" --drop table;`
	cleaned := CleanQuery(query)
	if cleaned != "mechanical keyboard drop table" {
		t.Errorf("unexpected cleaned query: %s", cleaned)
	}

	fts := BuildFTSQuery("mechanical keyb", true)
	if fts != `"mechanical" AND "keyb"*` {
		t.Errorf("unexpected fts query: %s", fts)
	}

	orQuery := BuildFTSOrQuery("fast search")
	if orQuery != `"fast"* OR "search"*` {
		t.Errorf("unexpected fts or query: %s", orQuery)
	}
}
