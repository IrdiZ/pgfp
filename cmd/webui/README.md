# Smith-Waterman Algorithm Web UI

This project provides a web-based interface for running the Smith-Waterman sequence alignment algorithm with both sequential and parallel implementations. The web UI allows users to visualize alignments, compare performance, and configure execution parameters.

## Features

### Sequence Alignment
- Load custom DNA sequences or generate random ones
- View sequence alignments with colorized matches, mismatches, and gaps
- Detect and visualize mutations (SNPs, insertions, deletions)

### Performance Analysis
- Toggle between sequential and parallel execution
- Configure worker count for parallel processing
- Support for batch alignment processing
- Real-time performance metrics (execution time, memory usage)
- Performance history tracking with interactive charts

### Visualization Tools
- Interactive alignment viewer
- Batch processing results viewer
- Performance comparison charts
- Mutation highlighting and analysis

## Getting Started

### Prerequisites
- Go 1.18 or higher
- Web browser with JavaScript enabled

### Running the Web UI

1. Clone the repository
2. Navigate to the project root directory
3. Run the web UI server:

```bash
go run cmd/webui/main.go
```

4. Open your browser and navigate to http://localhost:8080

### Usage Guide

1. **Input Sequences:**
    - Enter DNA sequences manually
    - Or generate random sequences of specified length

2. **Configure Alignment:**
    - Toggle parallel execution on/off
    - Set worker count (0 for auto)
    - Enable batch processing for multiple alignments

3. **View Results:**
    - Alignment visualization with color-coded matches
    - Performance metrics
    - History tracking charts

## Performance Benchmarking

For detailed performance analysis, use the profiling and benchmarking tools:

```bash
# Run benchmarks
go run cmd/benchmark/main.go --mode=all --length=1000

# Profile execution
go run cmd/profile/main.go --mode=parallel --length=2000 --cpuprofile=cpu.prof
```

## Implementation Details

The web UI integrates the following components:

- **Sequential Smith-Waterman** - Traditional single-threaded implementation
- **Parallel Smith-Waterman** - Multi-threaded implementation using Go goroutines
- **Batch Processing** - Support for aligning multiple sequences concurrently
- **Real-time Performance Tracking** - Charts and metrics for analysis

## License

This project is licensed under the MIT License - see the LICENSE file for details.