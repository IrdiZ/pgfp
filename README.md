# pgfp

DNA sequence alignment in Go: a sequential and a parallel version of the Smith-Waterman algorithm, with benchmarks and a small web UI.

## Context

Built as part of a university group project. This repository contains my implementation.

## What it does

- Aligns two DNA sequences with the Smith-Waterman local alignment algorithm.
- Has a sequential version (`align/smith_waterman.go`).
- Has a parallel version that uses goroutines (`align/parallel_smith_waterman.go`).
- Can align one query against many references at the same time (batch mode).
- Includes tools to benchmark, profile and visualize alignments.
- Includes a web UI to run alignments and compare run times.

## Quick start

You need Go 1.24 or newer.

```bash
# Run the tests
go test ./...

# Run the Go benchmarks in the align package
go test -bench=. ./align

# Run the demo in the root package
go run .

# Benchmark all modes (sequential, parallel, batch-seq, batch-par)
go run ./cmd/benchmark --mode=all --length=1000 --workers=4

# Profile one mode (writes a CPU profile)
go run ./cmd/profile --mode=parallel --length=2000 --workers=4 --cProfile=cpu.prof

# Write an HTML report for one alignment
go run ./cmd/visualize --output=report.html --query=GATTACA --reference=GATCACA

# Start the web UI, then open http://localhost:8080
go run ./cmd/webui
```

## Benchmarks

The numbers below come from `benchmark_results.html` in this repository. Each row compares the parallel run with 8 workers against the sequential run on the same input. The file does not record which machine produced them, so treat them as rough.

| Mode | Sequence length | Sequential (ms) | Parallel, 8 workers (ms) | Speedup |
|---|---|---|---|---|
| Batch | 100 | 7.33 | 3.67 | 2.00x |
| Batch | 500 | 150.66 | 44.33 | 3.40x |
| Batch | 1000 | 512.61 | 171.62 | 2.99x |
| Batch | 2000 | 2029.84 | 654.17 | 3.10x |
| Single pair | 1000 | 9.91 | 100.67 | 0.10x |
| Single pair | 2000 | 41.01 | 374.84 | 0.11x |

Batch mode reaches up to 3.40x. The parallel version for a single pair of sequences is slower than the sequential one in these runs (see "Known limits").

## How it works

Smith-Waterman fills a score matrix. Cell (i, j) depends on its three neighbors: (i-1, j-1), (i-1, j) and (i, j-1). A traceback from the highest score gives the best local alignment.

Sequential version: fill the matrix row by row, then trace back.

Parallel version (single pair): cells on the same anti-diagonal (i + j = constant) do not depend on each other. The code starts one goroutine per anti-diagonal and tracks the maximum score under a mutex.

```
        j ->
      0  1  2  3  4
  i   1  a  b  c  d      cells with the same letter are on one
  |   2  b  c  d  e      anti-diagonal and can be computed
  v   3  c  d  e  f      at the same time
```

Batch version: one query is aligned against many references. Each alignment runs the sequential algorithm in its own goroutine. A semaphore limits the number of goroutines to the worker count. The alignments are independent, so this part scales well.

## Known limits

- The single-pair parallel version starts all anti-diagonal goroutines at once and does not make a later diagonal wait for the earlier one. The scores in the benchmark file for this mode (for example 15 instead of 53 at length 100) also differ from the sequential scores. Do not rely on it for correct results until this is fixed.
- Creating one goroutine per diagonal costs more than the work it does, so this mode is slower than the sequential version.
- Sequences shorter than 50 bases fall back to the sequential code.
- The benchmark file has no machine information and is a single saved run.

## Web UI

`go run ./cmd/webui` starts a local server on port 8080. In the browser you can:

- enter DNA sequences or generate random ones;
- choose sequential or parallel mode and the worker count;
- run batch alignments;
- see the alignment with matches, mismatches and gaps colored;
- see run time and memory use, with a history chart.

More detail is in `cmd/webui/README.md`.

## Project layout

```
align/                         Smith-Waterman code and tests
  smith_waterman.go            sequential version
  parallel_smith_waterman.go   parallel and batch versions
  smith_waterman_test.go       tests
  benchmark_test.go            Go benchmarks
data/
  data.go, data_test.go        DNA sequence helpers
cmd/
  benchmark/  profile/  visualize/  webui/
main.go                        demo
benchmark_results.html         saved benchmark run
```

## License

MIT. See [LICENSE](LICENSE).
