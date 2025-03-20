package data

import (
	"strings"
	"testing"
)

// TestGenerateDNASequence ensures that DNA sequences are generated correctly
func TestGenerateDNASequence(t *testing.T) {
	// Test cases with different lengths
	testCases := []int{0, 1, 10, 100, 1000}

	for _, length := range testCases {
		// Generate a sequence of the specified length
		seq := GenerateDNASequence(length)

		// Check that the length is as expected
		if len(seq) != length {
			t.Errorf("Generated sequence length is %d, expected %d", len(seq), length)
		}

		// Check that the sequence only contains valid DNA bases
		for i, base := range seq {
			if base != 'A' && base != 'T' && base != 'C' && base != 'G' {
				t.Errorf("Invalid character %c at position %d in sequence", base, i)
			}
		}
	}

	// Test that sequences are random (different sequences for different calls)
	seq1 := GenerateDNASequence(100)
	seq2 := GenerateDNASequence(100)

	if seq1 == seq2 {
		t.Error("Two separately generated sequences are identical, suggesting randomness issues")
	}
}

// TestCreateSNP tests single nucleotide polymorphism creation
func TestCreateSNP(t *testing.T) {
	// Test with a known sequence
	original := "GATTACA"

	// Test SNP at each position
	for pos := 0; pos < len(original); pos++ {
		// Create SNP
		mutated := CreateSNP(original, pos)

		// Check sequence length is unchanged
		if len(mutated) != len(original) {
			t.Errorf("SNP changed sequence length from %d to %d", len(original), len(mutated))
		}

		// Check exactly one base is different
		differences := 0
		for i := 0; i < len(original); i++ {
			if original[i] != mutated[i] {
				differences++

				// Check the different base is at the expected position
				if i != pos {
					t.Errorf("SNP occurred at position %d, expected at position %d", i, pos)
				}

				// Check the new base is valid
				if mutated[i] != 'A' && mutated[i] != 'T' && mutated[i] != 'C' && mutated[i] != 'G' {
					t.Errorf("Invalid base %c after SNP", mutated[i])
				}

				// Check the new base is different from the original
				if mutated[i] == original[i] {
					t.Errorf("SNP didn't change the base at position %d", i)
				}
			}
		}

		if differences != 1 {
			t.Errorf("Expected 1 difference, found %d", differences)
		}
	}

	// Test with invalid positions
	invalid := []int{-1, len(original), len(original) + 10}
	for _, pos := range invalid {
		mutated := CreateSNP(original, pos)

		// Invalid positions should return the original sequence
		if mutated != original {
			t.Errorf("SNP with invalid position %d changed the sequence", pos)
		}
	}
}

// TestCreateInsertion tests insertion of sequences
func TestCreateInsertion(t *testing.T) {
	// Test with a known sequence
	original := "GATTACA"

	// Test insertions at different positions
	insertions := []struct {
		position int
		inserted string
	}{
		{0, "CG"},   // Insert at beginning
		{3, "TAG"},  // Insert in middle
		{7, "ACGT"}, // Insert at end
	}

	for _, tc := range insertions {
		// Create insertion
		mutated := CreateInsertion(original, tc.position, tc.inserted)

		// Check sequence length increased by expected amount
		expectedLength := len(original) + len(tc.inserted)
		if len(mutated) != expectedLength {
			t.Errorf("Insertion resulted in length %d, expected %d", len(mutated), expectedLength)
		}

		// Check the insertion is at the correct position
		prefix := original[:tc.position]
		suffix := original[tc.position:]
		expected := prefix + tc.inserted + suffix

		if mutated != expected {
			t.Errorf("Insertion result was %s, expected %s", mutated, expected)
		}
	}

	// Test with invalid positions
	invalid := []int{-1, len(original) + 10}
	for _, pos := range invalid {
		mutated := CreateInsertion(original, pos, "ACG")

		// Invalid positions should return the original sequence
		if mutated != original {
			t.Errorf("Insertion with invalid position %d changed the sequence", pos)
		}
	}
}

// TestCreateDeletion tests deletion of sequence sections
func TestCreateDeletion(t *testing.T) {
	// Test with a known sequence
	original := "GATTACA"

	// Test deletions at different positions and lengths
	deletions := []struct {
		position int
		length   int
		expected string
	}{
		{0, 2, "TTACA"}, // Delete from beginning
		{2, 3, "GACA"},  // Delete from middle
		{5, 2, "GATTA"}, // Delete to the end
		{3, 10, "GAT"},  // Delete more than available
	}

	for _, tc := range deletions {
		// Create deletion
		mutated := CreateDeletion(original, tc.position, tc.length)

		// Check result is as expected
		if mutated != tc.expected {
			t.Errorf("Deletion result was %s, expected %s", mutated, tc.expected)
		}
	}

	// Test with invalid positions
	invalid := []int{-1, len(original) + 10}
	for _, pos := range invalid {
		mutated := CreateDeletion(original, pos, 2)

		// Invalid positions should return the original sequence
		if mutated != original {
			t.Errorf("Deletion with invalid position %d changed the sequence", pos)
		}
	}
}

// TestCreateMutatedSequence tests random mutations at a given rate
func TestCreateMutatedSequence(t *testing.T) {
	// Create a test sequence
	original := strings.Repeat("GATTACA", 100) // 700 bases

	// Test different mutation rates
	rates := []float64{0.0, 0.05, 0.1, 0.5, 0.9, 1.0, 1.5}

	for _, rate := range rates {
		// Create mutated sequence
		mutated := CreateMutatedSequence(original, rate)

		// Check lengths are the same (this function doesn't add/remove bases)
		if len(mutated) != len(original) {
			t.Errorf("Mutation changed sequence length from %d to %d", len(original), len(mutated))
		}

		// Count differences
		differences := 0
		for i := 0; i < len(original); i++ {
			if original[i] != mutated[i] {
				differences++

				// Check the new base is valid
				if mutated[i] != 'A' && mutated[i] != 'T' && mutated[i] != 'C' && mutated[i] != 'G' {
					t.Errorf("Invalid base %c after mutation", mutated[i])
				}
			}
		}

		// For rate 0 or > 1, check expected behavior
		if rate <= 0 {
			if differences != 0 {
				t.Errorf("Zero/negative mutation rate resulted in %d differences", differences)
			}
		} else if rate > 1.0 {
			if mutated != original {
				t.Errorf("Mutation rate > 1.0 changed the sequence")
			}
		}
		// Note: For rates between 0 and 1, due to randomness, we don't check exact counts
	}
}

// TestCreateMultipleMutations tests adding a specific number of mutations
func TestCreateMultipleMutations(t *testing.T) {
	// Create a test sequence
	original := "GATTACAGATTACAGATTACA" // 21 bases

	// Test different numbers of mutations
	counts := []int{0, 1, 3, 5, 10, 25}

	for _, count := range counts {
		// Create mutated sequence
		mutated := CreateMultipleMutations(original, count)

		// Check lengths are the same
		if len(mutated) != len(original) {
			t.Errorf("Multiple mutations changed sequence length from %d to %d", len(original), len(mutated))
		}

		// Count differences
		differences := 0
		for i := 0; i < len(original); i++ {
			if original[i] != mutated[i] {
				differences++

				// Check the new base is valid
				if mutated[i] != 'A' && mutated[i] != 'T' && mutated[i] != 'C' && mutated[i] != 'G' {
					t.Errorf("Invalid base %c after multiple mutations", mutated[i])
				}
			}
		}

		// Check number of mutations is as expected
		expectedDifferences := count
		if count > len(original) {
			// Function should return original if count > length
			expectedDifferences = 0
		} else if count <= 0 {
			// Function should return original if count <= 0
			expectedDifferences = 0
		}

		if differences != expectedDifferences {
			t.Errorf("Multiple mutations resulted in %d differences, expected %d",
				differences, expectedDifferences)
		}
	}
}

// TestGenerateConsensusSequence tests consensus sequence generation
func TestGenerateConsensusSequence(t *testing.T) {
	// Test with simple sequences
	sequences := []string{
		"GATTACA",
		"GATCACA",
		"GATTACA",
		"GATAACA",
	}

	// Expected consensus: "GATTACA" (majority rule)
	expected := "GATTACA"
	consensus := GenerateConsensusSequence(sequences)

	if consensus != expected {
		t.Errorf("Consensus was %s, expected %s", consensus, expected)
	}

	// Test with different length sequences
	unequalSequences := []string{
		"GATTACA",
		"GATCACAG",
		"GATTA",
	}

	// Should only consider positions up to the shortest sequence
	expected = "GATTA"
	consensus = GenerateConsensusSequence(unequalSequences)

	if consensus != expected {
		t.Errorf("Consensus for unequal lengths was %s, expected %s", consensus, expected)
	}

	// Test with empty sequences
	if GenerateConsensusSequence([]string{}) != "" {
		t.Errorf("Consensus for empty slice was not empty string")
	}
}

// BenchmarkGenerateDNASequence benchmarks sequence generation performance
func BenchmarkGenerateDNASequence(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateDNASequence(1000)
	}
}

// BenchmarkCreateMutatedSequence benchmarks mutation performance
func BenchmarkCreateMutatedSequence(b *testing.B) {
	// Create a test sequence
	original := strings.Repeat("GATTACA", 1000) // 7000 bases

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateMutatedSequence(original, 0.05)
	}
}
