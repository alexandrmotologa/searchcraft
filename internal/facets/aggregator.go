package facets

// AggregateFacets counts occurrences of values for the specified facet fields across matching documents.
func AggregateFacets(documents []map[string]any, facetFields []string) map[string]map[string]int {
	if len(facetFields) == 0 || len(documents) == 0 {
		return nil
	}

	result := make(map[string]map[string]int)
	for _, field := range facetFields {
		result[field] = make(map[string]int)
	}

	for _, doc := range documents {
		for _, field := range facetFields {
			val, exists := doc[field]
			if !exists || val == nil {
				continue
			}

			switch v := val.(type) {
			case []string:
				for _, item := range v {
					result[field][item]++
				}
			case []any:
				for _, item := range v {
					str := Stringify(item)
					result[field][str]++
				}
			default:
				str := Stringify(v)
				result[field][str]++
			}
		}
	}

	return result
}
