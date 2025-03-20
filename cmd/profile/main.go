package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"pgfp/align"
	"pgfp/data"
)

// ProfileConfig holds profiling configuration
type ProfileConfig struct {
	CPUProfile  string
	MemProfile  string
	Mode        string
	SequenceLen int
	NumWorkers  int
	BatchSize   int
	Repetitions int
}

func main() {
	// Define command-line flags
	config := ProfileConfig{}

	flag.StringVar(&config.CPUProfile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&config.MemProfile, "memprofile", "", "write memory profile to file")
	flag.StringVar(&config.Mode, "mode", "sequential", "alignment mode: sequential, parallel, or batch")
	flag.IntVar(&config.SequenceLen, "length", 1000, "sequence length")
	flag.IntVar(&config.NumWorkers, "workers", 0, "number of workers (0 = auto)")
	flag.IntVar(&config.BatchSize, "batch", 10, "batch size for batch mode")
	flag.IntVar(&config.Repetitions, "reps", 1, "number of repetitions")
	flag.Parse()

	// Start CPU profiling if requested
	if config.CPUProfile != "" {
		f, err := os.Create(config.CPUProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	// Generate test data
	fmt.Printf("Generating test sequences (length: %d)...\n", config.SequenceLen)
	query := data.GenerateDNASequence(config.SequenceLen)
	reference := data.GenerateDNASequence(config.SequenceLen)

	// Prepare batch data if needed
	var references []string
	if config.Mode == "batch" {
		fmt.Printf("Generating %d reference sequences for batch processing...\n", config.BatchSize)
		references = make([]string, config.BatchSize)
		for i := range references {
			references[i] = data.GenerateDNASequence(config.SequenceLen)
		}
	}

	// Set number of workers
	if config.NumWorkers <= 0 {
		config.NumWorkers = runtime.GOMAXPROCS(0)
		fmt.Printf("Using auto worker count: %d\n", config.NumWorkers)
	}

	// Variables for tracking results and performance
	var result interface{}
	totalTime := time.Duration(0)

	// Run the selected alignment mode
	fmt.Printf("Running %s Smith-Waterman alignment (%d repetitions)...\n",
		config.Mode, config.Repetitions)

	for i := 0; i < config.Repetitions; i++ {
		// Garbage collect before each run
		runtime.GC()

		// Run and time the appropriate algorithm
		start := time.Now()

		switch config.Mode {
		case "sequential":
			result = align.SmithWaterman(query, reference)

		case "parallel":
			result = align.ParallelSmithWaterman(query, reference, config.NumWorkers)

		case "batch":
			result = align.ConcurrentSmithWatermanBatch(query, references, config.NumWorkers)

		default:
			fmt.Fprintf(os.Stderr, "Invalid mode: %s\n", config.Mode)
			os.Exit(1)
		}

		// Record execution time
		elapsed := time.Since(start)
		totalTime += elapsed

		// Report progress
		fmt.Printf("Run %d/%d: %v\n", i+1, config.Repetitions, elapsed)
	}

	// Report execution statistics
	avgTime := totalTime / time.Duration(config.Repetitions)
	fmt.Printf("\nExecution statistics:\n")
	fmt.Printf("- Total time: %v\n", totalTime)
	fmt.Printf("- Average time: %v per run\n", avgTime)

	// Print alignment results based on mode
	switch config.Mode {
	case "sequential":
		res := result.(align.AlignmentResult)
		fmt.Printf("Alignment score: %d\n", res.MaxScore)
		printShortAlignment(res.AlignedQuery, res.AlignedRef, res.MaxScore)

	case "parallel":
		res := result.(align.ParallelAlignmentResult)
		fmt.Printf("Alignment score: %d (at position [%d,%d])\n", res.MaxScore, res.MaxRow, res.MaxCol)
		printShortAlignment(res.AlignedQuery, res.AlignedRef, res.MaxScore)

	case "batch":
		results := result.([]align.AlignmentResult)
		fmt.Printf("Completed %d alignments\n", len(results))
		totalScore := 0
		for _, res := range results {
			totalScore += res.MaxScore
		}
		fmt.Printf("Average alignment score: %.1f\n", float64(totalScore)/float64(len(results)))
		fmt.Printf("First alignment score: %d\n", results[0].MaxScore)
		printShortAlignment(results[0].AlignedQuery, results[0].AlignedRef, results[0].MaxScore)
	}

	// Memory profiling if requested
	if config.MemProfile != "" {
		f, err := os.Create(config.MemProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create memory profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		runtime.GC() // Run GC before taking memory profile
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Could not write memory profile: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Memory profile written to %s\n", config.MemProfile)
	}

	// Report memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("\nMemory usage:\n")
	fmt.Printf("- Allocated: %v MiB\n", bToMb(m.Alloc))
	fmt.Printf("- Total allocated: %v MiB\n", bToMb(m.TotalAlloc))
	fmt.Printf("- System memory: %v MiB\n", bToMb(m.Sys))
	fmt.Printf("- Garbage collections: %v\n", m.NumGC)

	// Additional profiling insights
	fmt.Printf("\nProfiling insights:\n")
	fmt.Printf("- CPU cores available: %d\n", runtime.NumCPU())
	fmt.Printf("- Goroutines used: %d\n", runtime.NumGoroutine())

	// Calculate memory per base pair
	bytesPerBase := float64(m.TotalAlloc) / float64(config.SequenceLen)
	fmt.Printf("- Memory efficiency: %.2f bytes/base\n", bytesPerBase)

	// Print recommended best practices
	fmt.Printf("\nRecommendations:\n")
	if config.SequenceLen < 500 && config.Mode == "parallel" {
		fmt.Println("- For short sequences (<500 bp), sequential algorithm may be more efficient")
	}
	if config.NumWorkers > runtime.NumCPU() {
		fmt.Println("- Worker count exceeds available CPU cores, which may reduce performance")
	}
	fmt.Println("- For maximum performance, tune worker count based on your specific hardware")
	fmt.Println("- Batch processing is recommended for aligning many sequences against a single query")
}

// printShortAlignment displays the first part of an alignment
func printShortAlignment(query, reference string, score int) {
	maxLen := 50
	if len(query) > maxLen {
		query = query[:maxLen] + "..."
		reference = reference[:maxLen] + "..."
	}

	fmt.Println("\nAlignment (truncated):")
	fmt.Printf("Query:     %s\n", query)

	// Generate match line
	matchLine := make([]rune, len(query))
	for i := 0; i < len(query) && i < len(reference); i++ {
		if query[i] == reference[i] && query[i] != '-' && reference[i] != '-' {
			matchLine[i] = '|' // Match
		} else if query[i] != '-' && reference[i] != '-' {
			matchLine[i] = '.' // Mismatch
		} else {
			matchLine[i] = ' ' // Gap
		}
	}

	fmt.Printf("           %s\n", string(matchLine))
	fmt.Printf("Reference: %s\n", reference)
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
