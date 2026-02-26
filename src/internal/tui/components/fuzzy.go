package components

import "strings"

// FuzzyResult holds the result of a fuzzy match.
type FuzzyResult struct {
	Match   bool
	Score   int
	Indices []int // character positions that matched in the text
}

// FuzzyMatch performs fuzzy matching of query against text.
// Scoring: +1 per match, +3 consecutive, +2 word boundary.
func FuzzyMatch(text, query string) FuzzyResult {
	if query == "" {
		return FuzzyResult{Match: true, Score: 0}
	}

	lower := strings.ToLower(text)
	q := strings.ToLower(query)
	qi := 0
	score := 0
	indices := make([]int, 0, len(q))
	lastMatchIdx := -1

	for i := 0; i < len(lower) && qi < len(q); i++ {
		if lower[i] == q[qi] {
			indices = append(indices, i)
			// Consecutive match bonus
			if lastMatchIdx == i-1 {
				score += 3
			}
			// Word boundary bonus
			if i == 0 || text[i-1] == ' ' || text[i-1] == '_' || text[i-1] == '-' {
				score += 2
			}
			score += 1
			lastMatchIdx = i
			qi++
		}
	}

	return FuzzyResult{
		Match:   qi == len(q),
		Score:   score,
		Indices: indices,
	}
}
