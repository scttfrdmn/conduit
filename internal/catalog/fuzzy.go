package catalog

import "strings"

// FuzzyMatch calculates a similarity score between two strings (0.0 to 1.0)
// Uses Levenshtein distance for fuzzy matching
func FuzzyMatch(s1, s2 string) float64 {
	// Normalize strings
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	// Exact match
	if s1 == s2 {
		return 1.0
	}

	// Empty strings
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Check if one is substring of the other
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		longer := maxInt(len(s1), len(s2))
		shorter := minInt(len(s1), len(s2))
		return float64(shorter) / float64(longer)
	}

	// Calculate Levenshtein distance
	distance := levenshteinDistance(s1, s2)
	maxLen := maxInt(len(s1), len(s2))

	// Convert distance to similarity score (0.0 to 1.0)
	if maxLen == 0 {
		return 0.0
	}

	similarity := 1.0 - (float64(distance) / float64(maxLen))
	return similarity
}

// FuzzyResult represents a fuzzy match result
type FuzzyResult struct {
	Value string
	Score float64
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	// Create matrix
	m := len(s1)
	n := len(s2)

	// Handle edge cases
	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}

	// Create distance matrix
	d := make([][]int, m+1)
	for i := range d {
		d[i] = make([]int, n+1)
	}

	// Initialize first row and column
	for i := 0; i <= m; i++ {
		d[i][0] = i
	}
	for j := 0; j <= n; j++ {
		d[0][j] = j
	}

	// Calculate distances
	for j := 1; j <= n; j++ {
		for i := 1; i <= m; i++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			d[i][j] = min3(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}

	return d[m][n]
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// min3 returns the minimum of three integers
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// maxFloat64 returns the maximum of two float64 values
func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// FuzzySearchModels performs fuzzy search on model names
func FuzzySearchModels(models []string, query string, threshold float64) []FuzzyResult {
	var matches []FuzzyResult

	for _, model := range models {
		score := FuzzyMatch(query, model)
		if score >= threshold {
			matches = append(matches, FuzzyResult{
				Value: model,
				Score: score,
			})
		}
	}

	// Sort by score (highest first)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Score > matches[i].Score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	return matches
}
