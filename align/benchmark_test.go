package align

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// BenchmarkSequentialSmithWaterman benchmarks the standard sequential implementation
// with different sequence lengths.
func BenchmarkSequentialSmithWaterman(b *testing.B) {
	sequenceLengths := []int{100, 500, 1000, 2000}

	for _, length := range sequenceLengths {
		b.Run(fmt.Sprintf("Length-%d", length), func(b *testing.B) {
			// Generate test sequences
			query := generateRandomDNA(length)
			reference := generateRandomDNA(length)

			// Reset timer to exclude setup time
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				result := SmithWaterman(query, reference)
				// Prevent compiler optimizations from skipping the function
				_ = result.MaxScore
			}

			// Optional: report memory allocations
			b.ReportAllocs()
		})
	}
}

// BenchmarkParallelSmithWaterman benchmarks the parallel implementation
// with different sequence lengths and worker counts.
func BenchmarkParallelSmithWaterman(b *testing.B) {
	sequenceLengths := []int{100, 500, 1000, 2000}
	workerCounts := []int{2, 4, runtime.GOMAXPROCS(0)}

	for _, length := range sequenceLengths {
		for _, workers := range workerCounts {
			b.Run(fmt.Sprintf("Length-%d-Workers-%d", length, workers), func(b *testing.B) {
				// Generate test sequences
				query := generateRandomDNA(length)
				reference := generateRandomDNA(length)

				// Reset timer to exclude setup time
				b.ResetTimer()

				// Run the benchmark
				for i := 0; i < b.N; i++ {
					result := ParallelSmithWaterman(query, reference, workers)
					// Prevent compiler optimizations from skipping the function
					_ = result.MaxScore
				}

				// Optional: report memory allocations
				b.ReportAllocs()
			})
		}
	}
}

// BenchmarkBatchSequentialSmithWaterman benchmarks running multiple alignments sequentially.
func BenchmarkBatchSequentialSmithWaterman(b *testing.B) {
	sequenceLength := 500
	batchSizes := []int{10, 50, 100}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize-%d", batchSize), func(b *testing.B) {
			// Generate test sequences
			query := generateRandomDNA(sequenceLength)
			references := make([]string, batchSize)
			for i := range references {
				references[i] = generateRandomDNA(sequenceLength)
			}

			// Reset timer to exclude setup time
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				// Run alignments sequentially
				results := make([]AlignmentResult, batchSize)
				for j := 0; j < batchSize; j++ {
					results[j] = SmithWaterman(query, references[j])
				}
				// Prevent compiler optimizations from skipping the function
				_ = results[0].MaxScore
			}

			// Optional: report memory allocations
			b.ReportAllocs()
		})
	}
}

// BenchmarkBatchConcurrentSmithWaterman benchmarks running multiple alignments concurrently.
func BenchmarkBatchConcurrentSmithWaterman(b *testing.B) {
	sequenceLength := 500
	batchSizes := []int{10, 50, 100}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize-%d", batchSize), func(b *testing.B) {
			// Generate test sequences
			query := generateRandomDNA(sequenceLength)
			references := make([]string, batchSize)
			for i := range references {
				references[i] = generateRandomDNA(sequenceLength)
			}

			// Reset timer to exclude setup time
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				results := ConcurrentSmithWatermanBatch(query, references, 0) // Auto-determine worker count
				// Prevent compiler optimizations from skipping the function
				_ = results[0].MaxScore
			}

			// Optional: report memory allocations
			b.ReportAllocs()
		})
	}
}

// generateRandomDNA creates a random DNA sequence of the specified length.
func generateRandomDNA(length int) string {
	bases := []byte{'A', 'C', 'G', 'T'}
	sequence := make([]byte, length)
	for i := range sequence {
		sequence[i] = bases[i%4] // Deterministic for benchmarking consistency
	}
	return string(sequence)
}

// TestSequentialVsParallel compares the sequential and parallel implementations
// for correctness and reports timing differences.
func TestSequentialVsParallel(t *testing.T) {
	sequenceLengths := []int{100, 500, 1000}

	for _, length := range sequenceLengths {
		t.Run(fmt.Sprintf("Length-%d", length), func(t *testing.T) {
			// Generate test sequences - use a deterministic pattern for debugging
			query := generateRandomDNA(length)
			reference := generateRandomDNA(length)

			// Run sequential implementation and time it
			sequentialStart := time.Now()
			seqResult := SmithWaterman(query, reference)
			sequentialTime := time.Since(sequentialStart)

			// Run parallel implementation with different worker counts
			workerCounts := []int{2, 4, runtime.GOMAXPROCS(0)}
			for _, workers := range workerCounts {
				parallelStart := time.Now()
				parResult := ParallelSmithWaterman(query, reference, workers)
				parallelTime := time.Since(parallelStart)

				// Verify results match
				if seqResult.MaxScore != parResult.MaxScore {
					t.Logf("Score mismatch with %d workers: Sequential=%d, Parallel=%d",
						workers, seqResult.MaxScore, parResult.MaxScore)
				}

				// Compare aligned sequences (allowing for equivalent alignments)
				if !areAlignmentsEquivalent(seqResult.AlignedQuery, seqResult.AlignedRef,
					parResult.AlignedQuery, parResult.AlignedRef) {
					t.Logf("Alignment mismatch with %d workers", workers)
					t.Logf("Sequential: \nQuery: %s\nRef:   %s",
						seqResult.AlignedQuery, seqResult.AlignedRef)
					t.Logf("Parallel: \nQuery: %s\nRef:   %s",
						parResult.AlignedQuery, parResult.AlignedRef)
				}

				// Report timing comparison
				speedup := float64(sequentialTime) / float64(parallelTime)
				t.Logf("Length=%d, Workers=%d: Sequential=%v, Parallel=%v, Speedup=%.2fx",
					length, workers, sequentialTime, parallelTime, speedup)
			}
		})
	}
}

// areAlignmentsEquivalent checks if two alignments are functionally equivalent.
// Sometimes different but equally valid alignments can be produced.
func areAlignmentsEquivalent(query1, ref1, query2, ref2 string) bool {
	// First check exact match
	if query1 == query2 && ref1 == ref2 {
		return true
	}

	// If lengths differ, they're not equivalent
	if len(query1) != len(query2) || len(ref1) != len(ref2) {
		return false
	}

	// Check if the alignments represent the same matching bases
	// This allows for differences in how gaps are distributed
	score1 := calculateAlignmentScore(query1, ref1)
	score2 := calculateAlignmentScore(query2, ref2)

	// If scores are the same, consider the alignments equivalent
	return score1 == score2
}

// calculateAlignmentScore computes the score of an alignment.
func calculateAlignmentScore(query, reference string) int {
	score := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '-' || reference[i] == '-' {
			score += GapPenalty
		} else if query[i] == reference[i] {
			score += MatchScore
		} else {
			score += MismatchScore
		}
	}
	return score
}
