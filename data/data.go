package data

import (
	"math/rand"
	"strings"
	"time"
)

// Initialize a global random source once
var globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// DNA bases used in sequence generation
var bases = []rune{'A', 'T', 'C', 'G'}

// GenerateDNASequence generates a random DNA sequence of a given length.
//
// Purpose:
//   - Creates a random DNA sequence composed of bases 'A', 'T', 'C', and 'G'.
//   - Useful for testing DNA sequence alignment algorithms.
//
// Parameters:
//   - length (int): The length of the DNA sequence to generate.
//
// Returns:
//   - (string): A randomly generated DNA sequence of the specified length.
//
// Example Usage:
//
//	seq := GenerateDNASequence(10)  // Returns something like "ATCGGCTTGA"
func GenerateDNASequence(length int) string {
	// Create a sequence slice of the specified length
	seq := make([]rune, length)

	// Populate the sequence with random DNA bases
	for i := range seq {
		seq[i] = bases[globalRand.Intn(len(bases))]
	}

	// Convert the slice to a string and return it
	return string(seq)
}

// CreateSNP creates a sequence with a single nucleotide polymorphism (SNP) at the specified position.
//
// Parameters:
//   - original (string): The original DNA sequence.
//   - position (int): The position where the SNP should be introduced (0-based).
//
// Returns:
//   - (string): A new DNA sequence with a single base changed at the specified position.
func CreateSNP(original string, position int) string {
	if position < 0 || position >= len(original) {
		return original // Return original if position is invalid
	}

	// Create a different base than the original at the specified position
	var newBase rune
	originalBase := rune(original[position])

	// Keep generating a random base until it's different from the original
	for {
		newBase = bases[globalRand.Intn(len(bases))]
		if newBase != originalBase {
			break
		}
	}

	// Convert original to rune slice for manipulation
	seq := []rune(original)
	seq[position] = newBase

	return string(seq)
}

// CreateInsertion inserts a specified sequence at the given position in the original sequence.
//
// Parameters:
//   - original (string): The original DNA sequence.
//   - position (int): The position where the insertion should occur (0-based).
//   - inserted (string): The DNA sequence to insert.
//
// Returns:
//   - (string): A new DNA sequence with the insertion.
func CreateInsertion(original string, position int, inserted string) string {
	if position < 0 || position > len(original) {
		return original // Return original if position is invalid
	}

	return original[:position] + inserted + original[position:]
}

// CreateDeletion creates a sequence with a deletion of specified length starting at the given position.
//
// Parameters:
//   - original (string): The original DNA sequence.
//   - position (int): The start position of the deletion (0-based).
//   - length (int): The number of bases to delete.
//
// Returns:
//   - (string): A new DNA sequence with the specified deletion.
func CreateDeletion(original string, position int, length int) string {
	if position < 0 || position >= len(original) {
		return original // Return original if position is invalid
	}

	// Ensure we don't try to delete past the end of the sequence
	if position+length > len(original) {
		length = len(original) - position
	}

	return original[:position] + original[position+length:]
}

// CreateMutatedSequence creates a sequence with random mutations at the specified rate.
//
// Parameters:
//   - original (string): The original DNA sequence.
//   - mutationRate (float64): The probability (0.0-1.0) of each base being mutated.
//
// Returns:
//   - (string): A new DNA sequence with random mutations.
func CreateMutatedSequence(original string, mutationRate float64) string {
	if mutationRate <= 0 || mutationRate > 1 {
		return original // Return original if mutation rate is invalid
	}

	// Create a local random source with a unique seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	seq := []rune(original)

	for i := range seq {
		// Determine if this position should be mutated
		if r.Float64() < mutationRate {
			// Select a different base
			originalBase := seq[i]
			for {
				newBase := bases[r.Intn(len(bases))]
				if newBase != originalBase {
					seq[i] = newBase
					break
				}
			}
		}
	}

	return string(seq)
}

// CreateMultipleMutations applies multiple random mutations to a sequence.
//
// Parameters:
//   - original (string): The original DNA sequence.
//   - numMutations (int): The number of mutations to introduce.
//
// Returns:
//   - (string): A new DNA sequence with the specified number of mutations.
func CreateMultipleMutations(original string, numMutations int) string {
	if numMutations <= 0 || numMutations > len(original) {
		return original // Return original if number of mutations is invalid
	}

	seq := []rune(original)

	// Track positions that have already been mutated
	mutatedPositions := make(map[int]bool)

	for i := 0; i < numMutations; i++ {
		// Find a position that hasn't been mutated yet
		var position int
		for {
			position = globalRand.Intn(len(seq))
			if !mutatedPositions[position] {
				break
			}
		}

		// Mark this position as mutated
		mutatedPositions[position] = true

		// Change the base
		originalBase := seq[position]
		for {
			newBase := bases[globalRand.Intn(len(bases))]
			if newBase != originalBase {
				seq[position] = newBase
				break
			}
		}
	}

	return string(seq)
}

// GenerateConsensusSequence creates a consensus sequence from multiple DNA sequences.
//
// Parameters:
//   - sequences ([]string): The DNA sequences to create a consensus from.
//
// Returns:
//   - (string): A consensus sequence where each position contains the most common base.
func GenerateConsensusSequence(sequences []string) string {
	if len(sequences) == 0 {
		return ""
	}

	// Find the length of the shortest sequence
	minLength := len(sequences[0])
	for _, seq := range sequences {
		if len(seq) < minLength {
			minLength = len(seq)
		}
	}

	// Build the consensus sequence
	consensus := strings.Builder{}

	for i := 0; i < minLength; i++ {
		// Count occurrences of each base at this position
		counts := make(map[rune]int)
		for _, seq := range sequences {
			base := rune(seq[i])
			counts[base]++
		}

		// Find the most common base
		var mostCommonBase rune
		maxCount := 0
		for base, count := range counts {
			if count > maxCount {
				maxCount = count
				mostCommonBase = base
			}
		}

		consensus.WriteRune(mostCommonBase)
	}

	return consensus.String()
}
