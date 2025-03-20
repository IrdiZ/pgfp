package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"pgfp/align"
	"pgfp/data"
)

// VisualizationData represents alignment data for visualization
type VisualizationData struct {
	AlignedQuery string     `json:"alignedQuery"`
	AlignedRef   string     `json:"alignedRef"`
	Score        int        `json:"score"`
	Mutations    []Mutation `json:"mutations"`
}

// Mutation represents a detected mutation
type Mutation struct {
	Type     string `json:"type"`     // "snp", "insertion", "deletion"
	Position int    `json:"position"` // Position in the original sequence
	Length   int    `json:"length"`   // Length of the mutation (for insertions/deletions)
	Original string `json:"original"` // Original bases
	Mutated  string `json:"mutated"`  // Mutated bases
}

func main() {
	// Define flags
	outputPath := flag.String("output", "", "Path to output HTML file")
	querySeq := flag.String("query", "", "Query DNA sequence")
	refSeq := flag.String("reference", "", "Reference DNA sequence")
	generateRandom := flag.Bool("random", false, "Generate random sequences")
	seqLength := flag.Int("length", 1000, "Length for random sequences")
	useParallel := flag.Bool("parallel", false, "Use parallel Smith-Waterman")
	workers := flag.Int("workers", 0, "Number of workers for parallel execution (0 = auto)")
	runServer := flag.Bool("server", false, "Run as web server")
	serverPort := flag.Int("port", 8081, "Port for web server")

	flag.Parse()

	// Validate flags
	if !*runServer && *outputPath == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Error: must specify either -server or -output")
		flag.Usage()
		os.Exit(1)
	}

	// Get sequences
	var query, reference string
	if *generateRandom {
		log.Println("Generating random sequences of length", *seqLength)
		query = data.GenerateDNASequence(*seqLength)
		reference = data.GenerateDNASequence(*seqLength)
	} else {
		query = *querySeq
		reference = *refSeq

		if query == "" || reference == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Error: must provide both query and reference sequences, or use -random flag")
			flag.Usage()
			os.Exit(1)
		}
	}

	// Perform alignment
	var alignResult align.AlignmentResult
	startTime := time.Now()

	if *useParallel {
		log.Println("Running parallel Smith-Waterman alignment...")
		if *workers <= 0 {
			*workers = runtime.GOMAXPROCS(0)
			log.Printf("Using %d workers (auto)", *workers)
		} else {
			log.Printf("Using %d workers", *workers)
		}
		parallelResult := align.ParallelSmithWaterman(query, reference, *workers)
		alignResult = align.AlignmentResult{
			ScoreMatrix:  parallelResult.ScoreMatrix,
			MaxScore:     parallelResult.MaxScore,
			AlignedQuery: parallelResult.AlignedQuery,
			AlignedRef:   parallelResult.AlignedRef,
		}
	} else {
		log.Println("Running sequential Smith-Waterman alignment...")
		alignResult = align.SmithWaterman(query, reference)
	}

	elapsedTime := time.Since(startTime)
	log.Printf("Alignment completed in %v", elapsedTime)
	log.Printf("Alignment score: %d", alignResult.MaxScore)

	// Handle the result based on mode
	if *runServer {
		// Run as web server
		log.Printf("Starting visualization server on port %d...", *serverPort)
		err := serveVisualization(alignResult, *serverPort)
		if err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	} else {
		// Generate HTML file
		outPath := *outputPath
		if !strings.HasSuffix(outPath, ".html") {
			outPath += ".html"
		}

		// Ensure the output directory exists
		dir := filepath.Dir(outPath)
		if dir != "." && dir != "" {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				log.Fatalf("Error creating output directory: %v", err)
			}
		}

		log.Printf("Generating visualization to %s...", outPath)
		err := generateVisualization(alignResult, outPath)
		if err != nil {
			log.Fatalf("Error generating visualization: %v", err)
		}

		log.Println("Visualization generated successfully")
	}
}

// generateVisualization creates an HTML visualization of an alignment and saves it to a file
func generateVisualization(alignResult align.AlignmentResult, outputPath string) error {
	// Create a visualization d object
	visualData := VisualizationData{
		AlignedQuery: alignResult.AlignedQuery,
		AlignedRef:   alignResult.AlignedRef,
		Score:        alignResult.MaxScore,
		Mutations:    detectMutations(alignResult.AlignedQuery, alignResult.AlignedRef),
	}

	// Convert to JSON for use in the template
	jsonData, err := json.Marshal(visualData)
	if err != nil {
		return fmt.Errorf("error marshaling visualization d: %v", err)
	}

	// Create template d
	d := struct {
		AlignedQuery string
		AlignedRef   string
		Score        int
		Timestamp    string
		MatchLine    string
		JSONData     template.JS
	}{
		AlignedQuery: alignResult.AlignedQuery,
		AlignedRef:   alignResult.AlignedRef,
		Score:        alignResult.MaxScore,
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		MatchLine:    generateMatchLine(alignResult.AlignedQuery, alignResult.AlignedRef),
		JSONData:     template.JS(jsonData),
	}

	// Parse and execute the template
	tmpl, err := template.New("visualization").Parse(visualizationTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	// Create the output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Error closing output file: %v", err)
		}
	}(file)

	// Execute the template
	err = tmpl.Execute(file, d)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
}

// serveVisualization starts a web server to visualize alignments
func serveVisualization(alignResult align.AlignmentResult, port int) error {
	// Create a visualization data object
	visualData := VisualizationData{
		AlignedQuery: alignResult.AlignedQuery,
		AlignedRef:   alignResult.AlignedRef,
		Score:        alignResult.MaxScore,
		Mutations:    detectMutations(alignResult.AlignedQuery, alignResult.AlignedRef),
	}

	// Create a handler for serving the visualization
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Convert to JSON for use in the template
		jsonData, err := json.Marshal(visualData)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error marshaling d: %v", err), http.StatusInternalServerError)
			return
		}

		// Create template d
		d := struct {
			AlignedQuery string
			AlignedRef   string
			Score        int
			Timestamp    string
			MatchLine    string
			JSONData     template.JS
		}{
			AlignedQuery: alignResult.AlignedQuery,
			AlignedRef:   alignResult.AlignedRef,
			Score:        alignResult.MaxScore,
			Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
			MatchLine:    generateMatchLine(alignResult.AlignedQuery, alignResult.AlignedRef),
			JSONData:     template.JS(jsonData),
		}

		// Parse and execute the template
		tmpl, err := template.New("visualization").Parse(visualizationTemplate)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, d)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
			return
		}
	})

	// Start the server
	addr := ":" + strconv.Itoa(port)
	log.Printf("Starting visualization server at http://localhost%s", addr)
	return http.ListenAndServe(addr, nil)
}

// detectMutations analyzes aligned sequences to find mutations
func detectMutations(alignedQuery, alignedRef string) []Mutation {
	mutations := []Mutation{}

	// Keep track of positions in the original sequences
	queryPos, refPos := 0, 0

	// Variables to track potential insertions/deletions
	var currentMutation *Mutation

	for i := 0; i < len(alignedQuery) && i < len(alignedRef); i++ {
		if alignedQuery[i] == '-' {
			// Gap in query = deletion
			if currentMutation == nil || currentMutation.Type != "deletion" {
				// Start a new deletion
				currentMutation = &Mutation{
					Type:     "deletion",
					Position: refPos,
					Original: string(alignedRef[i]),
					Mutated:  "-",
					Length:   1,
				}
				mutations = append(mutations, *currentMutation)
			} else {
				// Continue the current deletion
				lastIdx := len(mutations) - 1
				mutations[lastIdx].Original += string(alignedRef[i])
				mutations[lastIdx].Length++
			}
			refPos++
		} else if alignedRef[i] == '-' {
			// Gap in reference = insertion
			if currentMutation == nil || currentMutation.Type != "insertion" {
				// Start a new insertion
				currentMutation = &Mutation{
					Type:     "insertion",
					Position: queryPos,
					Original: "-",
					Mutated:  string(alignedQuery[i]),
					Length:   1,
				}
				mutations = append(mutations, *currentMutation)
			} else {
				// Continue the current insertion
				lastIdx := len(mutations) - 1
				mutations[lastIdx].Mutated += string(alignedQuery[i])
				mutations[lastIdx].Length++
			}
			queryPos++
		} else if alignedQuery[i] != alignedRef[i] {
			// Mismatch = SNP
			mutations = append(mutations, Mutation{
				Type:     "snp",
				Position: queryPos,
				Original: string(alignedRef[i]),
				Mutated:  string(alignedQuery[i]),
				Length:   1,
			})
			queryPos++
			refPos++
			currentMutation = nil
		} else {
			// Match = no mutation
			queryPos++
			refPos++
			currentMutation = nil
		}
	}

	return mutations
}

// generateMatchLine creates a string representing matches/mismatches/gaps
func generateMatchLine(seq1, seq2 string) string {
	matchLine := make([]byte, len(seq1))

	for i := 0; i < len(seq1) && i < len(seq2); i++ {
		if seq1[i] == '-' || seq2[i] == '-' {
			matchLine[i] = ' ' // Gap
		} else if seq1[i] == seq2[i] {
			matchLine[i] = '|' // Match
		} else {
			matchLine[i] = '.' // Mismatch
		}
	}

	return string(matchLine)
}

// HTML template for visualization
const visualizationTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Smith-Waterman Alignment Visualization</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .alignment-container { 
            font-family: monospace; 
            white-space: pre;
            overflow-x: auto;
            background-color: #f5f5f5;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .alignment-row { margin: 0; }
        .match { color: green; }
        .mismatch { color: red; }
        .gap { color: gray; }
        .mutation { 
            margin: 10px 0;
            padding: 10px;
            border-radius: 5px;
        }
        .snp { background-color: #fff3cd; }
        .insertion { background-color: #d1e7dd; }
        .deletion { background-color: #f8d7da; }
        .mutation-highlight { 
            font-weight: bold;
            text-decoration: underline;
        }
        h1, h2 { color: #333; }
        .info { color: #666; margin-bottom: 5px; }
        pre { margin: 0; }
    </style>
</head>
<body>
    <h1>Smith-Waterman Alignment Visualization</h1>
    <div class="info">
        <strong>Alignment Score:</strong> {{.Score}}
    </div>
    <div class="info">
        <strong>Generated:</strong> {{.Timestamp}}
    </div>
    
    <h2>Alignment</h2>
    <div class="alignment-container">
        <pre class="alignment-row">Query:  {{.AlignedQuery}}</pre>
        <pre class="alignment-row">Match:  {{.MatchLine}}</pre>
        <pre class="alignment-row">Ref:    {{.AlignedRef}}</pre>
    </div>
    
    <h2>Detected Mutations</h2>
    <div id="mutations-container">
        <!-- Mutations will be inserted here -->
    </div>
    
    <h2>Statistics</h2>
    <div id="statistics">
        <div>Total Mutations: <span id="total-mutations">0</span></div>
        <div>SNPs: <span id="snp-count">0</span></div>
        <div>Insertions: <span id="insertion-count">0</span></div>
        <div>Deletions: <span id="deletion-count">0</span></div>
    </div>
    
    <script>
        // Alignment data from Go template
        const alignmentData = {{.JSONData}};
        
        // Display mutations
        function displayMutations(mutations) {
            const container = document.getElementById('mutations-container');
            
            if (mutations.length === 0) {
                container.innerHTML = '<div>No mutations detected.</div>';
                return;
            }
            
            let snps = 0, insertions = 0, deletions = 0;
            
            mutations.forEach((mutation, index) => {
                const div = document.createElement('div');
                div.className = 'mutation ' + mutation.type;
                
                let description = '';
                if (mutation.type === 'snp') {
                    description = 'SNP at position ' + mutation.position + ': ' + mutation.original + ' → ' + mutation.mutated;
                    snps++;
                } else if (mutation.type === 'insertion') {
                    description = 'Insertion at position ' + mutation.position + ': ' + mutation.mutated + ' inserted';
                    insertions++;
                } else if (mutation.type === 'deletion') {
                    description = 'Deletion at position ' + mutation.position + ': ' + mutation.original + ' deleted';
                    deletions++;
                }
                
                div.innerHTML = '<div><strong>Mutation #' + (index + 1) + ':</strong> ' + description + '</div>';
                container.appendChild(div);
            });
            
            // Update statistics
            document.getElementById('total-mutations').textContent = mutations.length;
            document.getElementById('snp-count').textContent = snps;
            document.getElementById('insertion-count').textContent = insertions;
            document.getElementById('deletion-count').textContent = deletions;
        }
        
        // Initialize visualization
        window.onload = function() {
            displayMutations(alignmentData.mutations || []);
        };
    </script>
</body>
</html>`
