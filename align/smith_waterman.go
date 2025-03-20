package align

// Scoring parameters
const (
	MatchScore    = 2  // Score for a matching base
	MismatchScore = -1 // Penalty for a mismatched base
	GapPenalty    = -2 // Penalty for an insertion or deletion
)

// AlignmentResult holds the alignment matrix and results.
type AlignmentResult struct {
	ScoreMatrix  [][]int // The Smith-Waterman dynamic programming matrix
	MaxScore     int     // Maximum score in the matrix
	AlignedQuery string  // The aligned query sequence
	AlignedRef   string  // The aligned reference sequence
}

// SmithWaterman performs local sequence alignment using the Smith-Waterman algorithm.
//
// Parameters:
//   - query (string): The DNA query sequence.
//   - reference (string): The DNA reference sequence.
//
// Returns:
//   - (AlignmentResult): A struct containing the alignment score matrix, maximum score, and aligned sequences.
func SmithWaterman(query, reference string) AlignmentResult {
	m, n := len(query), len(reference)

	// Initialize score matrix
	matrix := make([][]int, m+1)
	for i := range matrix {
		matrix[i] = make([]int, n+1)
	}

	maxScore := 0
	maxRow, maxCol := 0, 0

	// Fill the score matrix
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			// Determine if this is a match or mismatch
			match := MismatchScore
			if query[i-1] == reference[j-1] {
				match = MatchScore
			}

			// Compute scores
			scoreDiag := matrix[i-1][j-1] + match
			scoreUp := matrix[i-1][j] + GapPenalty
			scoreLeft := matrix[i][j-1] + GapPenalty

			// Apply Smith-Waterman scoring rule (no negative scores)
			matrix[i][j] = smithMax(0, scoreDiag, scoreUp, scoreLeft)

			// Track maximum score for traceback
			if matrix[i][j] > maxScore {
				maxScore = matrix[i][j]
				maxRow, maxCol = i, j
			}
		}
	}

	// Traceback to reconstruct the alignment
	alignedQuery, alignedRef := traceback(matrix, query, reference, maxRow, maxCol)

	return AlignmentResult{
		ScoreMatrix:  matrix,
		MaxScore:     maxScore,
		AlignedQuery: alignedQuery,
		AlignedRef:   alignedRef,
	}
}

// traceback reconstructs the best local alignment from the score matrix.
//
// Parameters:
//   - matrix ([][]int): The alignment score matrix.
//   - query (string): The query DNA sequence.
//   - reference (string): The reference DNA sequence.
//   - row (int): The row index of the highest score.
//   - col (int): The column index of the highest score.
//
// Returns:
//   - (string, string): The aligned query and reference sequences.
func traceback(matrix [][]int, query, reference string, row, col int) (string, string) {
	var alignedQuery, alignedRef string

	// Perform traceback from the highest scoring cell
	for row > 0 && col > 0 && matrix[row][col] > 0 {
		currentScore := matrix[row][col]

		// Calculate match score for current position
		match := MismatchScore
		if query[row-1] == reference[col-1] {
			match = MatchScore
		}

		// Check diagonal move (match/mismatch)
		if currentScore == matrix[row-1][col-1]+match {
			alignedQuery = string(query[row-1]) + alignedQuery
			alignedRef = string(reference[col-1]) + alignedRef
			row--
			col--
		} else if currentScore == matrix[row-1][col]+GapPenalty {
			// Gap in reference
			alignedQuery = string(query[row-1]) + alignedQuery
			alignedRef = "-" + alignedRef
			row--
		} else if currentScore == matrix[row][col-1]+GapPenalty {
			// Gap in query
			alignedQuery = "-" + alignedQuery
			alignedRef = string(reference[col-1]) + alignedRef
			col--
		} else {
			// This shouldn't happen with correct scoring, but break as a safeguard
			break
		}
	}

	return alignedQuery, alignedRef
}

// smithMax returns the maximum of the provided integer values.
func smithMax(values ...int) int {
	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}
