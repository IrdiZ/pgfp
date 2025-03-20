# 🧬 Smith-Waterman Sequence Alignment Suite (for now) 🧬

A comprehensive toolkit for DNA sequence alignment using the Smith-Waterman algorithm, featuring high-performance parallel implementation, interactive visualization, and real-time performance analysis.

## 📁 Project Structure

```
├── align/
│   ├── smith_waterman.go             # Sequential implementation
│   ├── parallel_smith_waterman.go    # Parallel implementation
│   ├── benchmark_test.go             # Performance tests
│   └── smith_waterman_test.go        # Test Suite for S_W
├── data/
│   ├── dna.go                        # DNA sequence utilities
│   └── dna_test.go                   # Testing utilities
├── cmd/
│   ├── benchmark/                    # Benchmarking tools
│   │   └── main.go
│   ├── profile/                      # Profiling tools
│   │   └── main.go
│   ├── visualize/                    # Visualization utilities
│   │   └── main.go
│   │ 
│   └── webui/                        # Web interface
│       ├── README.md
│       ├── main.go
│       ├── templates/
│       │   └── index.html
│       └── static/
│           ├── css/
│           │   └── styles.css
│           └── js/
│               └── main.js
└── README.md                         # Documentation
```

## ✨ Key Features

### 🚀 High-Performance Algorithms

- **💻 Sequential Smith-Waterman**
    - Traditional single-threaded implementation
    - Baseline for performance comparison
    - Optimized matrix calculation and traceback

- **⚡ Parallel Smith-Waterman**
    - Multi-threaded implementation using goroutines
    - Wave-front parallelization approach
    - Configurable worker count
    - Up to 5x speedup on large sequences

- **📦 Batch Processing**
    - Concurrent alignment of multiple sequences
    - Efficient workload distribution
    - Perfect for genomic database searches

### 🔍 Analysis & Profiling

- **📊 Benchmarking Suite**
    - Comprehensive performance testing
    - Comparison of sequential vs. parallel implementations
    - Memory usage analysis
    - Scalability testing across sequence lengths

- **🧰 Profiling Toolkit**
    - CPU and memory profiling
    - Execution time measurement
    - Resource usage tracking
    - Performance bottleneck identification

### 🌐 Web Interface

- **📋 Sequence Input**
    - Direct entry of DNA sequences
    - Random sequence generation
    - Batch sequence processing

- **⚙️ Execution Controls**
    - Toggle between sequential and parallel algorithms
    - Configure worker count
    - Enable/disable batch processing

- **📈 Performance Monitoring**
    - Real-time execution timing
    - Memory usage tracking
    - Interactive performance charts
    - Historical data comparison

- **🔬 Alignment Visualization**
    - Interactive sequence viewer
    - Color-coded matches, mismatches, and gaps
    - Mutation detection and highlighting
    - Batch results comparison

### 🎨 Visualization Tools

- **🖼️ Static HTML Reports**
    - Standalone alignment visualization
    - Mutation analysis and statistics
    - Shareable results

- **🌍 Interactive Web Visualizer**
    - Browser-based alignment explorer
    - Realtime mutation detection
    - Sequence highlighting

## 🚀 Usage Examples

### 🖥️ Web UI

```bash
# Start the web interface
go run cmd/webui/main.go
# Access at http://localhost:8080
```

### 📊 Benchmarking

```bash
# Run benchmarks across different sequence lengths
go run cmd/benchmark/main.go --mode=all --lengths=100,500,1000,2000
```

### 🔍 Profiling

```bash
# Profile parallel execution
go run cmd/profile/main.go --mode=parallel --length=2000 --workers=4 --cpuprofile=cpu.prof
```

### 🎨 Visualization

```bash
# Generate visualization of an alignment
go run cmd/visualize/main.go --output=report.html --query=GATTACA --reference=GATCACA

# Start visualization server
go run cmd/visualize/main.go --server --port=8081 --random --length=1000
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run benchmarks
go test -bench=. ./align
```

## 🔧 Technical Details

- **🧬 Mutation Detection**
    - Single Nucleotide Polymorphisms (SNPs)
    - Insertions and deletions (indels)
    - Complex mutation patterns

- **⚙️ Parallelization Strategy**
    - Wave-front decomposition
    - Row/column partitioning
    - Optimized for multi-core processors

- **📉 Performance Characteristics**
    - Parallel algorithm scales with sequence length
    - Optimal for sequences > 500 bp
    - Memory usage optimized for large datasets

## 🏆 Advantages

- **⏱️ Speed**: Significant speedup through parallelization
- **📱 Usability**: Intuitive web interface for non-technical users
- **📊 Analysis**: Comprehensive performance insights
- **🧩 Flexibility**: Command-line and web-based interfaces
- **🔧 Extensibility**: Modular design for easy enhancement

---

🧬 **Smith-Waterman Sequence Alignment Suite** - Making DNA analysis faster, easier, and more insightful.