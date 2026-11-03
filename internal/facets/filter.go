package facets

import (
	"fmt"
	"strconv"
)

// MatchesFilter checks if a document's JSON data satisfies all filter conditions.
func MatchesFilter(docData map[string]any, filters map[string]any) bool {
	if len(filters) == 0 {
		return true
	}

	for field, cond := range filters {
		val, exists := docData[field]
		if !exists {
			return false
		}

		if !evaluateCondition(val, cond) {
			return false
		}
	}

	return true
}

func evaluateCondition(docVal, condition any) bool {
	switch cond := condition.(type) {
	case map[string]any:
		// Range or operator condition: {"$gte": 10, "$lte": 100}
		for op, target := range cond {
			if !evaluateOperator(docVal, op, target) {
				return false
			}
		}
		return true

	case []any:
		// Implicit IN condition: ["Keyboards", "Mice"]
		return evaluateIn(docVal, cond)

	default:
		// Direct equality comparison
		return evaluateEquality(docVal, cond)
	}
}

func evaluateOperator(docVal any, op string, target any) bool {
	switch op {
	case "$eq":
		return evaluateEquality(docVal, target)
	case "$ne":
		return !evaluateEquality(docVal, target)
	case "$in":
		if targetSlice, ok := target.([]any); ok {
			return evaluateIn(docVal, targetSlice)
		}
		return false
	case "$nin":
		if targetSlice, ok := target.([]any); ok {
			return !evaluateIn(docVal, targetSlice)
		}
		return false
	case "$gt":
		numDoc, okDoc := toFloat64(docVal)
		numTarget, okTarget := toFloat64(target)
		return okDoc && okTarget && numDoc > numTarget
	case "$gte":
		numDoc, okDoc := toFloat64(docVal)
		numTarget, okTarget := toFloat64(target)
		return okDoc && okTarget && numDoc >= numTarget
	case "$lt":
		numDoc, okDoc := toFloat64(docVal)
		numTarget, okTarget := toFloat64(target)
		return okDoc && okTarget && numDoc < numTarget
	case "$lte":
		numDoc, okDoc := toFloat64(docVal)
		numTarget, okTarget := toFloat64(target)
		return okDoc && okTarget && numDoc <= numTarget
	default:
		return false
	}
}

func evaluateEquality(docVal, target any) bool {
	if docVal == target {
		return true
	}

	// Case insensitive string comparison or numeric equivalence
	s1, ok1 := docVal.(string)
	s2, ok2 := target.(string)
	if ok1 && ok2 {
		return s1 == s2
	}

	// Number comparison (e.g. int vs float64)
	n1, okN1 := toFloat64(docVal)
	n2, okN2 := toFloat64(target)
	if okN1 && okN2 {
		return n1 == n2
	}

	// Array field containing value (e.g. doc.tags contains "wireless")
	if slice, ok := docVal.([]any); ok {
		for _, item := range slice {
			if evaluateEquality(item, target) {
				return true
			}
		}
	}
	if slice, ok := docVal.([]string); ok {
		for _, item := range slice {
			if evaluateEquality(item, target) {
				return true
			}
		}
	}

	return false
}

func evaluateIn(docVal any, targetList []any) bool {
	for _, target := range targetList {
		if evaluateEquality(docVal, target) {
			return true
		}
	}
	return false
}

func toFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// Stringify turns any primitive into a string for facet keys.
func Stringify(val any) string {
	if s, ok := val.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", val)
}
