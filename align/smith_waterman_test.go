package align

import (
	"strings"
	"testing"
)

// TestCase defines the structure for test inputs and expected results.
type TestCase struct {
	Query         string
	Reference     string
	ExpectedScore int
	// Instead of expecting exact alignments, we'll check for correct matching regions
	RequiredMatches []string
}

// TestSmithWaterman runs multiple test cases to verify correctness.
func TestSmithWaterman(t *testing.T) {
	testCases := []TestCase{
		// Test Case 1: Perfect Match (Score = 7*2 = 14)
		{
			Query:           "GATTACA",
			Reference:       "GATTACA",
			ExpectedScore:   14,
			RequiredMatches: []string{"GATTACA"},
		},
		// Test Case 2: One Mismatch (Score = (6*2) - 1 = 11)
		{
			Query:           "GATTACA",
			Reference:       "GATTTCA",
			ExpectedScore:   11,
			RequiredMatches: []string{"GAT", "CA"}, // Before and after the mismatch
		},
		// Test Case 3: Insertion/Deletion (Score = (6*2) - 2 = 10)
		{
			Query:           "GATTACA",
			Reference:       "GATACA",
			ExpectedScore:   10,
			RequiredMatches: []string{"GAT", "ACA"}, // Before and after the indel
		},
		// Test Case 4: Insertion/Deletion (Score = (6*2) - 2 = 10)
		{
			Query:           "GATTACA",
			Reference:       "GATTCA",
			ExpectedScore:   10,
			RequiredMatches: []string{"GATT", "CA"}, // Before and after the indel
		},
		// Test Case 5: Partial match at beginning
		{
			Query:           "GATTACA",
			Reference:       "GATXXXX",
			ExpectedScore:   6, // 3 matches * 2
			RequiredMatches: []string{"GAT"},
		},
		// Test Case 6: Partial match at end
		{
			Query:           "GATTACA",
			Reference:       "XXXXACA",
			ExpectedScore:   6, // 3 matches * 2
			RequiredMatches: []string{"ACA"},
		},
		// Test Case 7: Local alignment should find best subsection
		{
			Query:           "XXGATTACAXX",
			Reference:       "YYGATTACAYY",
			ExpectedScore:   14, // Should find the GATTACA substring
			RequiredMatches: []string{"GATTACA"},
		},
	}

	// Run all test cases
	for i, tc := range testCases {
		result := SmithWaterman(tc.Query, tc.Reference)

		// Validate alignment score
		if result.MaxScore != tc.ExpectedScore {
			t.Errorf("Test case %d - FAIL: Query: %s, Ref: %s\nExpected Score: %d, Got: %d",
				i+1, tc.Query, tc.Reference, tc.ExpectedScore, result.MaxScore)
		}

		// Validate that required matching sections are present in the alignment
		strippedQuery := stripGaps(result.AlignedQuery)
		strippedRef := stripGaps(result.AlignedRef)

		for _, match := range tc.RequiredMatches {
			if !containsSubsequence(strippedQuery, match) {
				t.Errorf("Test case %d - FAIL: Query alignment %s does not contain required match: %s",
					i+1, strippedQuery, match)
			}
			if !containsSubsequence(strippedRef, match) {
				t.Errorf("Test case %d - FAIL: Reference alignment %s does not contain required match: %s",
					i+1, strippedRef, match)
			}
		}

		// Check that the aligned sequences make biological sense
		if !isValidAlignment(result.AlignedQuery, result.AlignedRef) {
			t.Errorf("Test case %d - FAIL: Invalid alignment: \nQuery: %s\nRef: %s",
				i+1, result.AlignedQuery, result.AlignedRef)
		}
	}
}

// stripGaps removes all gap characters from a sequence
func stripGaps(seq string) string {
	return strings.ReplaceAll(seq, "-", "")
}

// containsSubsequence checks if a sequence contains a specific subsequence
func containsSubsequence(seq, subseq string) bool {
	return strings.Contains(seq, subseq)
}

// isValidAlignment checks if an alignment makes biological sense
func isValidAlignment(query, reference string) bool {
	// Alignments must be the same length
	if len(query) != len(reference) {
		return false
	}

	// Check that we don't have gaps in both sequences at the same position
	for i := 0; i < len(query); i++ {
		if query[i] == '-' && reference[i] == '-' {
			return false
		}
	}

	return true
}
