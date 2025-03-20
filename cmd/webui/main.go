package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"runtime"
	_ "strconv"
	"strings"
	"time"

	"pgfp/align"
	"pgfp/data"
)

// AlignmentRequest represents a request for sequence alignment
type AlignmentRequest struct {
	Query          string `json:"query"`
	Reference      string `json:"reference"`
	UseParallel    bool   `json:"useParallel"`
	Workers        int    `json:"workers"`
	GenerateRandom bool   `json:"generateRandom"`
	RandomLength   int    `json:"randomLength"`
	BatchSize      int    `json:"batchSize"`
	UseBatch       bool   `json:"useBatch"`
}

// AlignmentResponse represents the response to an alignment request
type AlignmentResponse struct {
	QuerySequence   string          `json:"querySequence"`
	RefSequence     string          `json:"refSequence"`
	AlignedQuery    string          `json:"alignedQuery"`
	AlignedRef      string          `json:"alignedRef"`
	Score           int             `json:"score"`
	ExecutionTime   string          `json:"executionTime"`
	ExecutionTimeMs float64         `json:"executionTimeMs"`
	MemoryUsageMB   uint64          `json:"memoryUsageMB"`
	IsParallel      bool            `json:"isParallel"`
	Workers         int             `json:"workers"`
	BatchResults    []BatchResult   `json:"batchResults,omitempty"`
	PerformanceData PerformanceData `json:"performanceData"`
}

// BatchResult represents the result of a batch alignment
type BatchResult struct {
	Index        int    `json:"index"`
	Score        int    `json:"score"`
	AlignedQuery string `json:"alignedQuery"`
	AlignedRef   string `json:"alignedRef"`
}

// PerformanceData represents performance metrics
type PerformanceData struct {
	CpuCores       int     `json:"cpuCores"`
	Goroutines     int     `json:"goroutines"`
	AllocatedMB    uint64  `json:"allocatedMB"`
	SystemMemoryMB uint64  `json:"systemMemoryMB"`
	BytesPerBase   float64 `json:"bytesPerBase"`
	GcRuns         uint32  `json:"gcRuns"`
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port int
}

func main() {
	// Set up server config
	config := ServerConfig{
		Port: 8080,
	}

	// Set up the HTTP server
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("./cmd/webui/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Set up routes
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/align", handleAlign)
	mux.HandleFunc("/system-info", handleSystemInfo)

	// Start the server
	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("Starting server on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// handleIndex serves the main HTML page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("./cmd/webui/templates/index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
		return
	}

	// Get system information for the template
	cpuCores := runtime.NumCPU()

	d := struct {
		CPUCores int
	}{
		CPUCores: cpuCores,
	}

	err = tmpl.Execute(w, d)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleAlign processes alignment requests
func handleAlign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request
	var req AlignmentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request: %v", err), http.StatusBadRequest)
		return
	}

	// Prepare sequences
	query := req.Query
	reference := req.Reference

	// Generate random sequences if requested
	if req.GenerateRandom {
		length := req.RandomLength
		if length <= 0 {
			length = 100 // Default length
		}

		query = data.GenerateDNASequence(length)
		reference = data.GenerateDNASequence(length)
	}

	// Validate sequences
	if !isValidDNA(query) || !isValidDNA(reference) {
		http.Error(w, "Invalid DNA sequence. Use only A, C, G, T characters.", http.StatusBadRequest)
		return
	}

	// Set default worker count if needed
	if req.Workers <= 0 {
		req.Workers = runtime.GOMAXPROCS(0)
	}

	// Prepare response
	resp := AlignmentResponse{
		QuerySequence: query,
		RefSequence:   reference,
		IsParallel:    req.UseParallel,
		Workers:       req.Workers,
	}

	// Clear memory before alignment
	runtime.GC()

	// Get initial memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Start timing
	startTime := time.Now()

	// Perform the alignment
	if req.UseBatch {
		// Create batch of references
		batchSize := req.BatchSize
		if batchSize <= 0 {
			batchSize = 10 // Default batch size
		}

		references := make([]string, batchSize)
		for i := range references {
			if i == 0 {
				references[i] = reference // Use the original reference as first
			} else {
				// Create slightly modified references
				references[i] = data.CreateMultipleMutations(reference, 3)
			}
		}

		// Process batch
		var results []align.AlignmentResult
		if req.UseParallel {
			results = align.ConcurrentSmithWatermanBatch(query, references, req.Workers)
		} else {
			results = make([]align.AlignmentResult, len(references))
			for i, ref := range references {
				results[i] = align.SmithWaterman(query, ref)
			}
		}

		// Save batch results
		resp.BatchResults = make([]BatchResult, len(results))
		totalScore := 0
		for i, result := range results {
			totalScore += result.MaxScore
			resp.BatchResults[i] = BatchResult{
				Index:        i,
				Score:        result.MaxScore,
				AlignedQuery: result.AlignedQuery,
				AlignedRef:   result.AlignedRef,
			}
		}

		// Use the first result for the main display
		resp.AlignedQuery = results[0].AlignedQuery
		resp.AlignedRef = results[0].AlignedRef
		resp.Score = results[0].MaxScore
	} else {
		// Single alignment
		var result interface{}
		if req.UseParallel {
			result = align.ParallelSmithWaterman(query, reference, req.Workers)
			parallelResult := result.(align.ParallelAlignmentResult)
			resp.AlignedQuery = parallelResult.AlignedQuery
			resp.AlignedRef = parallelResult.AlignedRef
			resp.Score = parallelResult.MaxScore
		} else {
			result = align.SmithWaterman(query, reference)
			seqResult := result.(align.AlignmentResult)
			resp.AlignedQuery = seqResult.AlignedQuery
			resp.AlignedRef = seqResult.AlignedRef
			resp.Score = seqResult.MaxScore
		}
	}

	// Stop timing
	executionTime := time.Since(startTime)
	resp.ExecutionTime = executionTime.String()
	resp.ExecutionTimeMs = float64(executionTime) / float64(time.Millisecond)

	// Get final memory stats
	runtime.ReadMemStats(&m)
	resp.MemoryUsageMB = m.Alloc / (1024 * 1024)

	// Add performance data
	bytesPerBase := float64(m.TotalAlloc) / float64(len(query)+len(reference))
	resp.PerformanceData = PerformanceData{
		CpuCores:       runtime.NumCPU(),
		Goroutines:     runtime.NumGoroutine(),
		AllocatedMB:    m.Alloc / (1024 * 1024),
		SystemMemoryMB: m.Sys / (1024 * 1024),
		BytesPerBase:   bytesPerBase,
		GcRuns:         m.NumGC,
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleSystemInfo returns information about the system
func handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	// Gather system information
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Create response
	info := struct {
		CPUCores       int    `json:"cpuCores"`
		GoVersion      string `json:"goVersion"`
		NumGoroutines  int    `json:"numGoroutines"`
		AllocatedMemMB uint64 `json:"allocatedMemMB"`
		SystemMemMB    uint64 `json:"systemMemMB"`
	}{
		CPUCores:       runtime.NumCPU(),
		GoVersion:      runtime.Version(),
		NumGoroutines:  runtime.NumGoroutine(),
		AllocatedMemMB: m.Alloc / (1024 * 1024),
		SystemMemMB:    m.Sys / (1024 * 1024),
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(info)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// isValidDNA checks if a string is a valid DNA sequence
func isValidDNA(s string) bool {
	if s == "" {
		return false
	}

	s = strings.ToUpper(s)
	for _, c := range s {
		if c != 'A' && c != 'C' && c != 'G' && c != 'T' {
			return false
		}
	}

	return true
}
