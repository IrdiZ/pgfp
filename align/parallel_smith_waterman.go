package align

import (
	"runtime"
	"sync"
)

// ParallelAlignmentResult holds the alignment matrix and results for parallel execution.
type ParallelAlignmentResult struct {
	ScoreMatrix  [][]int // The Smith-Waterman dynamic programming matrix
	MaxScore     int     // Maximum score in the matrix
	MaxRow       int     // Row index of the maximum score
	MaxCol       int     // Column index of the maximum score
	AlignedQuery string  // The aligned query sequence
	AlignedRef   string  // The aligned reference sequence
}

// ParallelSmithWaterman performs local sequence alignment using the Smith-Waterman
// algorithm with parallel matrix calculation using goroutines.
//
// Parameters:
//   - query (string): The DNA query sequence.
//   - reference (string): The DNA reference sequence.
//   - numWorkers (int): Number of goroutines to use (0 = use GOMAXPROCS)
//
// Returns:
//   - (ParallelAlignmentResult): A struct containing the alignment matrix and results.
func ParallelSmithWaterman(query, reference string, numWorkers int) ParallelAlignmentResult {
	m, n := len(query), len(reference)

	// If the number of workers is not specified, use the number of CPUs
	if numWorkers <= 0 {
		numWorkers = runtime.GOMAXPROCS(0)
	}

	// For very small sequences, just use sequential algorithm
	if m < 50 || n < 50 {
		result := SmithWaterman(query, reference)
		return ParallelAlignmentResult{
			ScoreMatrix:  result.ScoreMatrix,
			MaxScore:     result.MaxScore,
			MaxRow:       0, // Not tracked in sequential version
			MaxCol:       0, // Not tracked in sequential version
			AlignedQuery: result.AlignedQuery,
			AlignedRef:   result.AlignedRef,
		}
	}

	// Initialize score matrix
	matrix := make([][]int, m+1)
	for i := range matrix {
		matrix[i] = make([]int, n+1)
	}

	// Shared variables for maximum score tracking (protected by mutex)
	var mu sync.Mutex
	maxScore := 0
	maxRow, maxCol := 0, 0

	// Calculate work chunks - divide matrix into blocks
	// Using wave-front decomposition instead of block decomposition

	// Use wait group to synchronize workers
	var wg sync.WaitGroup

	// Process the matrix in diagonal waves to handle dependencies
	// Each cell (i,j) depends on (i-1,j-1), (i-1,j), and (i,j-1)
	for wave := 2; wave <= m+n; wave++ {
		wg.Add(1)
		go func(waveFront int) {
			defer wg.Done()

			// Process all cells where i+j = waveFront
			for i := 1; i <= m && i < waveFront; i++ {
				j := waveFront - i
				if j < 1 || j > n {
					continue // Skip invalid coordinates
				}

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
				if matrix[i][j] > 0 {
					mu.Lock()
					if matrix[i][j] > maxScore {
						maxScore = matrix[i][j]
						maxRow, maxCol = i, j
					}
					mu.Unlock()
				}
			}
		}(wave)
	}

	// Wait for all diagonal waves to complete
	wg.Wait()

	// Perform traceback to reconstruct the alignment
	alignedQuery, alignedRef := parallelTraceback(matrix, query, reference, maxRow, maxCol)

	return ParallelAlignmentResult{
		ScoreMatrix:  matrix,
		MaxScore:     maxScore,
		MaxRow:       maxRow,
		MaxCol:       maxCol,
		AlignedQuery: alignedQuery,
		AlignedRef:   alignedRef,
	}
}

// parallelTraceback reconstructs the best local alignment from the score matrix.
// This implementation doesn't actually run the traceback in parallel (which is complex),
// but is designed to be compatible with the parallel Smith-Waterman implementation.
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
func parallelTraceback(matrix [][]int, query, reference string, row, col int) (string, string) {
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

// ConcurrentSmithWatermanBatch processes multiple sequence alignments concurrently.
// This function is useful for aligning one query against multiple references.
//
// Parameters:
//   - query (string): The DNA query sequence.
//   - references ([]string): An array of reference DNA sequences.
//   - numWorkers (int): Maximum number of concurrent alignments (0 = use GOMAXPROCS).
//
// Returns:
//   - ([]AlignmentResult): Array of alignment results, one per reference.
func ConcurrentSmithWatermanBatch(query string, references []string, numWorkers int) []AlignmentResult {
	// If the number of workers is not specified, use the number of CPUs
	if numWorkers <= 0 {
		numWorkers = runtime.GOMAXPROCS(0)
	}

	// Limit workers to number of references
	if numWorkers > len(references) {
		numWorkers = len(references)
	}

	// Create a channel for results and a semaphore channel to limit concurrency
	results := make([]AlignmentResult, len(references))
	semaphore := make(chan struct{}, numWorkers)
	var wg sync.WaitGroup

	// Process each reference sequence
	for i, ref := range references {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(index int, reference string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			// Run the standard Smith-Waterman algorithm
			results[index] = SmithWaterman(query, reference)
		}(i, ref)
	}

	// Wait for all alignments to complete
	wg.Wait()
	close(semaphore)

	return results
}
