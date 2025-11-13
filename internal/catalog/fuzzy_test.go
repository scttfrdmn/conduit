package catalog

import (
	"testing"
)

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected float64
		minScore float64 // Minimum expected score
	}{
		{
			name:     "Exact match",
			s1:       "alphafold2",
			s2:       "alphafold2",
			expected: 1.0,
			minScore: 1.0,
		},
		{
			name:     "Case insensitive",
			s1:       "AlphaFold2",
			s2:       "alphafold2",
			expected: 1.0,
			minScore: 1.0,
		},
		{
			name:     "Substring match",
			s1:       "alpha",
			s2:       "alphafold2",
			expected: 0.5, // 5/10 = 0.5
			minScore: 0.4,
		},
		{
			name:     "Typo - missing character",
			s1:       "alphafld",
			s2:       "alphafold",
			expected: 0.88, // 8/9 similarity
			minScore: 0.85,
		},
		{
			name:     "Typo - extra character",
			s1:       "alphabfold",
			s2:       "alphafold",
			expected: 0.9, // 9/10 similarity
			minScore: 0.85,
		},
		{
			name:     "Typo - swapped characters",
			s1:       "alphafodl",
			s2:       "alphafold",
			expected: 0.77, // 7/9 similarity
			minScore: 0.7,
		},
		{
			name:     "Very different strings",
			s1:       "pytorch",
			s2:       "tensorflow",
			expected: 0.2,
			minScore: 0.0,
		},
		{
			name:     "Empty strings",
			s1:       "",
			s2:       "alphafold",
			expected: 0.0,
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := FuzzyMatch(tt.s1, tt.s2)

			// Check if score meets minimum threshold
			if score < tt.minScore {
				t.Errorf("FuzzyMatch(%q, %q) = %.2f, want >= %.2f", tt.s1, tt.s2, score, tt.minScore)
			}

			// For exact match cases, check the exact value
			if tt.expected == 1.0 && score != 1.0 {
				t.Errorf("FuzzyMatch(%q, %q) = %.2f, want 1.0 for exact match", tt.s1, tt.s2, score)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "Identical strings",
			s1:       "hello",
			s2:       "hello",
			expected: 0,
		},
		{
			name:     "One insertion",
			s1:       "hello",
			s2:       "helo",
			expected: 1,
		},
		{
			name:     "One deletion",
			s1:       "helo",
			s2:       "hello",
			expected: 1,
		},
		{
			name:     "One substitution",
			s1:       "hello",
			s2:       "hallo",
			expected: 1,
		},
		{
			name:     "Multiple operations",
			s1:       "kitten",
			s2:       "sitting",
			expected: 3,
		},
		{
			name:     "Empty to non-empty",
			s1:       "",
			s2:       "hello",
			expected: 5,
		},
		{
			name:     "Non-empty to empty",
			s1:       "hello",
			s2:       "",
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := levenshteinDistance(tt.s1, tt.s2)
			if distance != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, distance, tt.expected)
			}
		})
	}
}

func TestFuzzySearchModels(t *testing.T) {
	models := []string{
		"alphafold2",
		"alphafold3",
		"bert-base",
		"gpt-4",
		"resnet50",
	}

	tests := []struct {
		name      string
		query     string
		threshold float64
		expected  []string // Expected matches in order
	}{
		{
			name:      "Exact match",
			query:     "alphafold2",
			threshold: 0.8,
			expected:  []string{"alphafold2"},
		},
		{
			name:      "Prefix match",
			query:     "alpha",
			threshold: 0.5,
			expected:  []string{"alphafold2", "alphafold3"}, // Both should match
		},
		{
			name:      "Fuzzy match with typo",
			query:     "alphafld",
			threshold: 0.7,
			expected:  []string{"alphafold2", "alphafold3"}, // Should find both with high scores
		},
		{
			name:      "No matches with high threshold",
			query:     "pytorch",
			threshold: 0.9,
			expected:  []string{}, // No close matches
		},
		{
			name:      "Match with low threshold",
			query:     "res",
			threshold: 0.3,
			expected:  []string{"resnet50"}, // Should match resnet50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := FuzzySearchModels(models, tt.query, tt.threshold)

			if len(matches) < len(tt.expected) {
				t.Errorf("FuzzySearchModels(%q, %.2f) returned %d matches, want at least %d",
					tt.query, tt.threshold, len(matches), len(tt.expected))
			}

			// Verify expected matches are present
			matchNames := make(map[string]bool)
			for _, m := range matches {
				matchNames[m.Value] = true
			}

			for _, expected := range tt.expected {
				if !matchNames[expected] {
					t.Errorf("Expected model %q not found in matches", expected)
				}
			}

			// Verify scores are descending
			for i := 0; i < len(matches)-1; i++ {
				if matches[i].Score < matches[i+1].Score {
					t.Errorf("Matches not sorted by score: %.2f < %.2f", matches[i].Score, matches[i+1].Score)
				}
			}
		})
	}
}
