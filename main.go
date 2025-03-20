package main

import (
	"fmt"
	"strings"

	"pgfp/align"
	"pgfp/data"
)

// printAlignment displays an alignment in a readable format
func printAlignment(query, reference string, score int) {
	fmt.Println("Alignment:")
	fmt.Printf("Score: %d\n", score)
	fmt.Printf("Query:     %s\n", query)

	// Generate the match line
	matchLine := make([]rune, len(query))
	for i := 0; i < len(query); i++ {
		if i < len(reference) && query[i] == reference[i] {
			matchLine[i] = '|' // Match
		} else if i < len(reference) && query[i] != '-' && reference[i] != '-' {
			matchLine[i] = '.' // Mismatch
		} else {
			matchLine[i] = ' ' // Gap
		}
	}

	fmt.Printf("           %s\n", string(matchLine))
	fmt.Printf("Reference: %s\n", reference)
	fmt.Println()
}

// demonstrateSNP shows how the algorithm handles a Single Nucleotide Polymorphism
func demonstrateSNP() {
	fmt.Println("===== DEMONSTRATION: Single Nucleotide Polymorphism (SNP) =====")

	// Generate a reference sequence
	reference := "GATTACAGATCAGATAGATACAGATAGACCA"
	fmt.Printf("Original Sequence: %s\n\n", reference)

	// Create a sequence with an SNP
	query := data.CreateSNP(reference, 15)
	fmt.Printf("Sequence with SNP: %s\n", query)

	// Find the position of the SNP
	for i := 0; i < len(reference); i++ {
		if reference[i] != query[i] {
			fmt.Printf("SNP at position %d: %c → %c\n\n", i, reference[i], query[i])
			break
		}
	}

	// Align using Smith-Waterman
	result := align.SmithWaterman(query, reference)
	printAlignment(result.AlignedQuery, result.AlignedRef, result.MaxScore)
}

// demonstrateInsertion shows how the algorithm handles an insertion
func demonstrateInsertion() {
	fmt.Println("===== DEMONSTRATION: Insertion =====")

	// Generate a reference sequence
	reference := "GATTACAGATCAGATAGATACAGATAGACCA"
	fmt.Printf("Original Sequence: %s\n\n", reference)

	// Create a sequence with an insertion
	insertion := "ACT"
	position := 10
	query := data.CreateInsertion(reference, position, insertion)
	fmt.Printf("Sequence with insertion: %s\n", query)
	fmt.Printf("Inserted '%s' at position %d\n\n", insertion, position)

	// Align using Smith-Waterman
	result := align.SmithWaterman(query, reference)
	printAlignment(result.AlignedQuery, result.AlignedRef, result.MaxScore)
}

// demonstrateDeletion shows how the algorithm handles a deletion
func demonstrateDeletion() {
	fmt.Println("===== DEMONSTRATION: Deletion =====")

	// Generate a reference sequence
	reference := "GATTACAGATCAGATAGATACAGATAGACCA"
	fmt.Printf("Original Sequence: %s\n\n", reference)

	// Create a sequence with a deletion
	position := 12
	length := 4
	query := data.CreateDeletion(reference, position, length)
	fmt.Printf("Sequence with deletion: %s\n", query)
	fmt.Printf("Deleted %d bases at position %d\n\n", length, position)

	// Align using Smith-Waterman
	result := align.SmithWaterman(query, reference)
	printAlignment(result.AlignedQuery, result.AlignedRef, result.MaxScore)
}

// demonstrateMultipleMutations shows how the algorithm handles multiple mutations
func demonstrateMultipleMutations() {
	fmt.Println("===== DEMONSTRATION: Multiple Mutations =====")

	// Generate a reference sequence
	reference := "GATTACAGATCAGATAGATACAGATAGACCA"
	fmt.Printf("Original Sequence: %s\n\n", reference)

	// Create a sequence with multiple mutations
	query := data.CreateMultipleMutations(reference, 3)
	fmt.Printf("Sequence with 3 random mutations: %s\n\n", query)

	// Find the mutations
	differences := 0
	fmt.Println("Mutations:")
	for i := 0; i < len(reference) && i < len(query); i++ {
		if reference[i] != query[i] {
			fmt.Printf("  Position %d: %c → %c\n", i, reference[i], query[i])
			differences++
		}
	}

	if len(reference) != len(query) {
		fmt.Printf("  Length difference: %d → %d\n", len(reference), len(query))
	}
	fmt.Println()

	// Align using Smith-Waterman
	result := align.SmithWaterman(query, reference)
	printAlignment(result.AlignedQuery, result.AlignedRef, result.MaxScore)
}

// demonstrateComplexMutationPattern shows combining multiple mutation operations
func demonstrateComplexMutationPattern() {
	fmt.Println("===== DEMONSTRATION: Complex Mutation Pattern =====")

	// Generate a reference sequence
	reference := "GATTACAGATCAGATAGATACAGATAGACCA"
	fmt.Printf("Original Sequence: %s\n\n", reference)

	// Apply a series of mutations
	// 1. First apply an SNP
	afterSNP := data.CreateSNP(reference, 5)

	// 2. Then apply an insertion
	afterInsertion := data.CreateInsertion(afterSNP, 15, "ACGT")

	// 3. Finally apply a deletion
	query := data.CreateDeletion(afterInsertion, 20, 3)

	fmt.Printf("Sequence after multiple mutations: %s\n", query)

	// Describe the mutations
	fmt.Println("Applied mutations:")
	fmt.Println("  1. SNP at position 5")
	fmt.Println("  2. Insertion of 'ACGT' at position 15")
	fmt.Println("  3. Deletion of 3 bases at position 20")
	fmt.Println()

	// Align using Smith-Waterman
	result := align.SmithWaterman(query, reference)
	printAlignment(result.AlignedQuery, result.AlignedRef, result.MaxScore)
}

// demonstrateLocalAlignment shows how the algorithm handles partial matches
func demonstrateLocalAlignment() {
	fmt.Println("===== DEMONSTRATION: Local Alignment Capability =====")

	// Create a reference sequence with a known pattern in the middle
	knownPattern := "GATTACA"
	prefix := "XXXXXX"
	suffix := "YYYYYY"
	reference := prefix + knownPattern + suffix
	fmt.Printf("Reference with pattern in middle: %s\n", reference)
	fmt.Printf("Known pattern: %s (positions %d-%d)\n\n", knownPattern, len(prefix), len(prefix)+len(knownPattern)-1)

	// Create a query with just the pattern
	query := knownPattern
	fmt.Printf("Query (just the pattern): %s\n\n", query)

	// Align using Smith-Waterman
	result := align.SmithWaterman(query, reference)
	printAlignment(result.AlignedQuery, result.AlignedRef, result.MaxScore)

	// Check if the alignment correctly identified the pattern
	alignedRefStripped := strings.ReplaceAll(result.AlignedRef, "-", "")
	if alignedRefStripped == knownPattern {
		fmt.Println("SUCCESS: Smith-Waterman correctly identified the local pattern!")
	} else {
		fmt.Println("FAIL: Smith-Waterman did not correctly identify the local pattern.")
	}
}

// demonstrateRealWorldExample shows a realistic use case with longer sequences
func demonstrateRealWorldExample() {
	fmt.Println("===== DEMONSTRATION: Realistic Use Case with Longer Sequences =====")

	// Create a longer reference sequence (e.g., a gene fragment)
	reference := data.GenerateDNASequence(200)
	fmt.Printf("Reference sequence (200 bp): %s...\n", reference[:50])

	// Create a query with a combination of mutations
	// 1. Start with the reference
	query := reference
	// 2. Apply multiple SNPs
	query = data.CreateMultipleMutations(query, 5)
	// 3. Add an insertion
	query = data.CreateInsertion(query, 75, "ACGTACGT")
	// 4. Add a deletion
	query = data.CreateDeletion(query, 120, 6)

	fmt.Printf("Mutated query sequence: %s...\n\n", query[:50])
	fmt.Println("Mutations applied:")
	fmt.Println("  - 5 random SNPs")
	fmt.Println("  - 8 bp insertion at position 75")
	fmt.Println("  - 6 bp deletion at position 120")
	fmt.Println()

	// Align using Smith-Waterman
	result := align.SmithWaterman(query, reference)

	// For long sequences, just print the alignment score and statistics
	fmt.Printf("Alignment Score: %d\n", result.MaxScore)

	// Count matches, mismatches, and gaps
	matches, mismatches, queryGaps, refGaps := 0, 0, 0, 0
	for i := 0; i < len(result.AlignedQuery); i++ {
		if i < len(result.AlignedRef) {
			if result.AlignedQuery[i] == '-' {
				queryGaps++
			} else if result.AlignedRef[i] == '-' {
				refGaps++
			} else if result.AlignedQuery[i] == result.AlignedRef[i] {
				matches++
			} else {
				mismatches++
			}
		}
	}

	fmt.Printf("Alignment Statistics:\n")
	fmt.Printf("  - Matches: %d\n", matches)
	fmt.Printf("  - Mismatches: %d\n", mismatches)
	fmt.Printf("  - Gaps in Query: %d\n", queryGaps)
	fmt.Printf("  - Gaps in Reference: %d\n", refGaps)
	fmt.Printf("  - Alignment Length: %d\n", len(result.AlignedQuery))

	// Print a sample of the alignment (first 50 characters)
	fmt.Println("\nSample of the alignment (first 50 characters):")
	if len(result.AlignedQuery) > 50 {
		printAlignment(result.AlignedQuery[:50]+"...", result.AlignedRef[:50]+"...", result.MaxScore)
	} else {
		printAlignment(result.AlignedQuery, result.AlignedRef, result.MaxScore)
	}
}

// demonstrateConsensusSequence shows how to generate a consensus sequence from multiple related sequences
func demonstrateConsensusSequence() {
	fmt.Println("===== DEMONSTRATION: Consensus Sequence Generation =====")

	// Generate a reference sequence
	reference := "GATTACAGATCAGATAGATACAGATAGACCA"
	fmt.Printf("Original Sequence: %s\n\n", reference)

	// Create multiple variants of the sequence
	variants := make([]string, 5)
	variants[0] = reference

	// Add mutations to generate variants
	variants[1] = data.CreateSNP(reference, 3)
	variants[2] = data.CreateSNP(reference, 10)
	variants[3] = data.CreateSNP(reference, 17)
	variants[4] = data.CreateSNP(reference, 25)

	fmt.Println("Sequence variants:")
	for i, variant := range variants {
		fmt.Printf("  Variant %d: %s\n", i+1, variant)
	}
	fmt.Println()

	// Generate consensus sequence
	consensus := data.GenerateConsensusSequence(variants)
	fmt.Printf("Consensus Sequence: %s\n\n", consensus)

	// Compare consensus to reference
	differences := 0
	for i := 0; i < len(reference) && i < len(consensus); i++ {
		if reference[i] != consensus[i] {
			differences++
		}
	}

	fmt.Printf("Differences between consensus and reference: %d\n", differences)
	fmt.Println("Note: The consensus sequence should match the reference because most variants agree with the reference at each position.")
}

func main() {
	fmt.Println("DNA MUTATION DETECTION WITH SMITH-WATERMAN ALGORITHM")
	fmt.Println("===================================================")
	fmt.Println()

	demonstrateSNP()
	fmt.Println(strings.Repeat("-", 80))

	demonstrateInsertion()
	fmt.Println(strings.Repeat("-", 80))

	demonstrateDeletion()
	fmt.Println(strings.Repeat("-", 80))

	demonstrateMultipleMutations()
	fmt.Println(strings.Repeat("-", 80))

	demonstrateComplexMutationPattern()
	fmt.Println(strings.Repeat("-", 80))

	demonstrateLocalAlignment()
	fmt.Println(strings.Repeat("-", 80))

	demonstrateConsensusSequence()
	fmt.Println(strings.Repeat("-", 80))

	demonstrateRealWorldExample()
}
