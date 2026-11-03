package facets

import (
	"testing"
)

func TestFilterEvaluation(t *testing.T) {
	doc := map[string]any{
		"title":    "Mechanical Keyboard",
		"category": "Keyboards",
		"price":    99.50,
		"rating":   4.8,
		"in_stock": true,
		"tags":     []string{"wireless", "mechanical", "rgb"},
	}

	// 1. Exact equality
	if !MatchesFilter(doc, map[string]any{"category": "Keyboards"}) {
		t.Errorf("expected category Keyboards to match")
	}
	if MatchesFilter(doc, map[string]any{"category": "Mice"}) {
		t.Errorf("expected category Mice NOT to match")
	}

	// 2. Numeric range ($gte and $lte)
	rangeFilter := map[string]any{
		"price": map[string]any{
			"$gte": 50.0,
			"$lte": 150.0,
		},
	}
	if !MatchesFilter(doc, rangeFilter) {
		t.Errorf("expected price 99.50 to match range [50, 150]")
	}

	outOfRangeFilter := map[string]any{
		"price": map[string]any{
			"$lt": 50.0,
		},
	}
	if MatchesFilter(doc, outOfRangeFilter) {
		t.Errorf("expected price 99.50 NOT to match < 50")
	}

	// 3. Boolean check
	if !MatchesFilter(doc, map[string]any{"in_stock": true}) {
		t.Errorf("expected in_stock true to match")
	}

	// 4. Array inclusion
	if !MatchesFilter(doc, map[string]any{"tags": "wireless"}) {
		t.Errorf("expected tags containing wireless to match")
	}
}

func TestAggregateFacets(t *testing.T) {
	docs := []map[string]any{
		{"category": "Keyboards", "brand": "Keychron", "tags": []string{"rgb", "wireless"}},
		{"category": "Keyboards", "brand": "Logitech", "tags": []string{"wireless"}},
		{"category": "Mice", "brand": "Logitech", "tags": []string{"ergonomic", "wireless"}},
	}

	facets := AggregateFacets(docs, []string{"category", "brand", "tags"})
	if facets == nil {
		t.Fatalf("expected facets map, got nil")
	}

	if facets["category"]["Keyboards"] != 2 {
		t.Errorf("expected 2 Keyboards, got %d", facets["category"]["Keyboards"])
	}
	if facets["category"]["Mice"] != 1 {
		t.Errorf("expected 1 Mice, got %d", facets["category"]["Mice"])
	}

	if facets["brand"]["Logitech"] != 2 {
		t.Errorf("expected 2 Logitech, got %d", facets["brand"]["Logitech"])
	}

	if facets["tags"]["wireless"] != 3 {
		t.Errorf("expected 3 wireless tags, got %d", facets["tags"]["wireless"])
	}
	if facets["tags"]["rgb"] != 1 {
		t.Errorf("expected 1 rgb tag, got %d", facets["tags"]["rgb"])
	}
}
