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

// ExecutionMode represents the algorithm execution mode
type ExecutionMode int

const (
	Sequential ExecutionMode = iota
	Parallel
	BatchSequential
	BatchParallel
)

func (m ExecutionMode) String() string {
	return [...]string{"Sequential", "Parallel", "BatchSequential", "BatchParallel"}[m]
}

func main() {
	// Define command-line flags
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile := flag.String("memprofile", "", "write memory profile to file")
	modeFlag := flag.String("mode", "all", "benchmark mode: sequential, parallel, batch-seq, batch-par, or all")
	seqLength := flag.Int("length", 1000, "sequence length")
	numWorkers := flag.Int("workers", runtime.GOMAXPROCS(0), "number of workers for parallel execution")
	batchSize := flag.Int("batch", 10, "batch size for batch mode")
	repetitions := flag.Int("reps", 3, "number of repetitions for more accurate timing")
	flag.Parse()

	// Determine which modes to benchmark
	var modesToRun []ExecutionMode
	switch *modeFlag {
	case "sequential":
		modesToRun = []ExecutionMode{Sequential}
	case "parallel":
		modesToRun = []ExecutionMode{Parallel}
	case "batch-seq":
		modesToRun = []ExecutionMode{BatchSequential}
	case "batch-par":
		modesToRun = []ExecutionMode{BatchParallel}
	case "all":
		modesToRun = []ExecutionMode{Sequential, Parallel, BatchSequential, BatchParallel}
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Invalid mode: %s\n", *modeFlag)
		os.Exit(1)
	}

	// Start CPU profiling if requested
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {

			}
		}(f)
		if err := pprof.StartCPUProfile(f); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	// Track execution times
	var sequentialTime, parallelTime time.Duration
	var batchSeqTime, batchParTime time.Duration

	// Generate test data only once for all benchmarks
	fmt.Printf("Generating test sequences (length: %d)...\n", *seqLength)
	query := data.GenerateDNASequence(*seqLength)
	reference := data.GenerateDNASequence(*seqLength)

	// Prepare batch data if needed
	var references []string
	if containsAny(modesToRun, BatchSequential, BatchParallel) {
		fmt.Printf("Generating %d reference sequences for batch processing...\n", *batchSize)
		references = make([]string, *batchSize)
		for i := range references {
			references[i] = data.GenerateDNASequence(*seqLength)
		}
	}

	// Run benchmarks for each requested mode
	for _, mode := range modesToRun {
		fmt.Printf("\n=== Benchmarking %s Mode ===\n", mode)

		switch mode {
		case Sequential:
			// Run sequential benchmark
			fmt.Printf("Running sequential Smith-Waterman (length: %d, repetitions: %d)...\n",
				*seqLength, *repetitions)
			sequentialTime = runSequentialBenchmark(query, reference, *repetitions)
			fmt.Printf("Sequential execution time: %v\n", sequentialTime)

		case Parallel:
			// Run parallel benchmark
			fmt.Printf("Running parallel Smith-Waterman (length: %d, workers: %d, repetitions: %d)...\n",
				*seqLength, *numWorkers, *repetitions)
			parallelTime = runParallelBenchmark(query, reference, *numWorkers, *repetitions)
			fmt.Printf("Parallel execution time: %v\n", parallelTime)

			// Report speedup if sequential was also run
			if sequentialTime > 0 {
				speedup := float64(sequentialTime) / float64(parallelTime)
				fmt.Printf("Speedup factor: %.2fx\n", speedup)
			}

		case BatchSequential:
			// Run batch sequential benchmark
			fmt.Printf("Running sequential batch processing (length: %d, batch size: %d, repetitions: %d)...\n",
				*seqLength, *batchSize, *repetitions)
			batchSeqTime = runBatchSequentialBenchmark(query, references, *repetitions)
			fmt.Printf("Sequential batch execution time: %v\n", batchSeqTime)

		case BatchParallel:
			// Run batch parallel benchmark
			fmt.Printf("Running parallel batch processing (length: %d, batch size: %d, workers: %d, repetitions: %d)...\n",
				*seqLength, *batchSize, *numWorkers, *repetitions)
			batchParTime = runBatchParallelBenchmark(query, references, *numWorkers, *repetitions)
			fmt.Printf("Parallel batch execution time: %v\n", batchParTime)

			// Report speedup if batch sequential was also run
			if batchSeqTime > 0 {
				speedup := float64(batchSeqTime) / float64(batchParTime)
				fmt.Printf("Batch speedup factor: %.2fx\n", speedup)
			}
		}
	}

	// Print overall comparison if multiple modes were run
	if len(modesToRun) > 1 {
		fmt.Printf("\n=== Performance Summary ===\n")

		if sequentialTime > 0 && parallelTime > 0 {
			fmt.Printf("Single alignment: Sequential = %v, Parallel = %v, Speedup = %.2fx\n",
				sequentialTime, parallelTime, float64(sequentialTime)/float64(parallelTime))
		}

		if batchSeqTime > 0 && batchParTime > 0 {
			fmt.Printf("Batch processing: Sequential = %v, Parallel = %v, Speedup = %.2fx\n",
				batchSeqTime, batchParTime, float64(batchSeqTime)/float64(batchParTime))
		}
	}

	// Memory profiling if requested
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not create memory profile: %v\n", err)
			os.Exit(1)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {

			}
		}(f)
		runtime.GC() // Run GC before taking memory profile
		if err := pprof.WriteHeapProfile(f); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not write memory profile: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Memory profile written to %s\n", *memprofile)
	}

	// Report memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("\nMemory usage:\n")
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

// runSequentialBenchmark runs the sequential algorithm and returns execution time
func runSequentialBenchmark(query, reference string, repetitions int) time.Duration {
	totalTime := time.Duration(0)

	for i := 0; i < repetitions; i++ {
		start := time.Now()
		result := align.SmithWaterman(query, reference)
		totalTime += time.Since(start)

		// Report score from first run
		if i == 0 {
			fmt.Printf("Alignment score: %d\n", result.MaxScore)
		}
	}

	return totalTime / time.Duration(repetitions)
}

// runParallelBenchmark runs the parallel algorithm and returns execution time
func runParallelBenchmark(query, reference string, workers, repetitions int) time.Duration {
	totalTime := time.Duration(0)

	for i := 0; i < repetitions; i++ {
		start := time.Now()
		result := align.ParallelSmithWaterman(query, reference, workers)
		totalTime += time.Since(start)

		// Report score from first run
		if i == 0 {
			fmt.Printf("Alignment score: %d\n", result.MaxScore)
		}
	}

	return totalTime / time.Duration(repetitions)
}

// runBatchSequentialBenchmark runs sequential batch processing and returns execution time
func runBatchSequentialBenchmark(query string, references []string, repetitions int) time.Duration {
	totalTime := time.Duration(0)

	for i := 0; i < repetitions; i++ {
		start := time.Now()

		// Process each reference sequentially
		results := make([]align.AlignmentResult, len(references))
		for j, ref := range references {
			results[j] = align.SmithWaterman(query, ref)
		}

		totalTime += time.Since(start)

		// Report average score from first run
		if i == 0 {
			totalScore := 0
			for _, result := range results {
				totalScore += result.MaxScore
			}
			fmt.Printf("Average alignment score: %.1f\n", float64(totalScore)/float64(len(results)))
		}
	}

	return totalTime / time.Duration(repetitions)
}

// runBatchParallelBenchmark runs parallel batch processing and returns execution time
func runBatchParallelBenchmark(query string, references []string, workers, repetitions int) time.Duration {
	totalTime := time.Duration(0)

	for i := 0; i < repetitions; i++ {
		start := time.Now()
		results := align.ConcurrentSmithWatermanBatch(query, references, workers)
		totalTime += time.Since(start)

		// Report average score from first run
		if i == 0 {
			totalScore := 0
			for _, result := range results {
				totalScore += result.MaxScore
			}
			fmt.Printf("Average alignment score: %.1f\n", float64(totalScore)/float64(len(results)))
		}
	}

	return totalTime / time.Duration(repetitions)
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// containsAny checks if the slice contains any of the given values
func containsAny(slice []ExecutionMode, values ...ExecutionMode) bool {
	for _, v := range values {
		for _, s := range slice {
			if s == v {
				return true
			}
		}
	}
	return false
}
