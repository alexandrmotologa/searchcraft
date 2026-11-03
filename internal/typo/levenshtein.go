package typo

// DamerauLevenshtein computes the edit distance between strings a and b,
// including insertions, deletions, substitutions, and adjacent character transpositions.
func DamerauLevenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la := len(ra)
	lb := len(rb)

	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// 2D distance matrix (la+1) x (lb+1)
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}

			// Deletion, insertion, substitution
			del := d[i-1][j] + 1
			ins := d[i][j-1] + 1
			sub := d[i-1][j-1] + cost

			minVal := del
			if ins < minVal {
				minVal = ins
			}
			if sub < minVal {
				minVal = sub
			}
			d[i][j] = minVal

			// Transposition
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				trans := d[i-2][j-2] + 1
				if trans < d[i][j] {
					d[i][j] = trans
				}
			}
		}
	}

	return d[la][lb]
}

// MaxAllowedDistance returns the maximum number of typos permitted for a word of length n.
// Rule:
// - Length 1-3: 0 typos (exact match only)
// - Length 4-7: 1 typo
// - Length >= 8: 2 typos
func MaxAllowedDistance(wordLen int) int {
	if wordLen <= 3 {
		return 0
	}
	if wordLen <= 7 {
		return 1
	}
	return 2
}

// IsWithinTypoTolerance checks if two words match within the adaptive typo threshold.
func IsWithinTypoTolerance(target, candidate string) bool {
	l := len([]rune(target))
	maxDist := MaxAllowedDistance(l)
	dist := DamerauLevenshtein(target, candidate)
	return dist <= maxDist
}
