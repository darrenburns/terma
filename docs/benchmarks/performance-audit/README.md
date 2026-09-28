# Performance audit — 28 September 2026

The [HTML report](../../performance-audit.html) ranks six fixes by expected impact
in a typical mixed-widget TUI. That ranking is engineering judgment, not a survey
of applications. Large relative gains in a microbenchmark do not imply the same
gain in a complete app.

## Source and environment

- Baseline: `ad4bdcb` (freshly fast-forwarded `origin/main`).
- After: the working-tree fixes delivered with this report.
- Apple M4 Pro, darwin/arm64, Go 1.25.5, default GOMAXPROCS=14.
- Five samples per workload, 200 ms per sample, with allocation reporting.
- Both versions were compiled before timing. Baseline and after suites alternated
  for each sample; agent test, build, browser, and other benchmark jobs were
  paused while timing. This is a local comparison, not a portable threshold.
- Reported values are medians. `results.json` also retains every sample and its
  observed minimum and maximum; these ranges are not confidence intervals.

The baseline was exported from Git and given the same new benchmark harnesses,
without any production-code fixes. Correctness regressions were run separately,
since several intentionally fail (or time out) against the baseline.

## Reproduce

Run the same harness on the baseline and changed source, using the same machine
and Go version. These commands express the measured workloads; the captured run
used precompiled test binaries and alternated the two versions five times with
`-test.count=1` to reduce ordering effects.

```sh
go test . -run '^$' -bench '^(BenchmarkTextPerf|BenchmarkRetainedNestedLayout$|BenchmarkRetainedPartialDamage$|BenchmarkSignalSelectorFanout$|BenchmarkTreeFlattenVisible$|BenchmarkReactivityLeaf$|BenchmarkReactivityTextSize$|BenchmarkReactivityListCursor$)' -benchmem -benchtime=200ms -count=5
go test ./layout -run '^$' -bench '^BenchmarkTextPerf' -benchmem -benchtime=200ms -count=5
python3 docs/benchmarks/performance-audit/summarize.py \
  docs/benchmarks/performance-audit/before.txt \
  docs/benchmarks/performance-audit/after.txt \
  > docs/benchmarks/performance-audit/results.json
```

To recreate the baseline, export `ad4bdcb` into a temporary directory with
`git archive`, then copy these harnesses from the changed source into it:
`text_perf_test.go`, `retained_performance_test.go`,
`signal_notification_perf_test.go`, `tree_perf_test.go`, and
`layout/text_perf_test.go`. The reactivity helper suite is already in the baseline.

## What each measurement includes

- `TextPerfCollectSpans`: rich-text segmentation and wrapping, without painting
  or signal delivery. Status text is 77 bytes; paragraph text is 2,112 bytes;
  the unbroken-word stress case is 4,096 bytes.
- `TextPerfReactiveSpans`: a signal mutation and renderer update, alternating a
  visible prefix on the same status row or paragraph. The extra prefix adds one
  byte. Text occupies 80×1 or 80×24 cells inside a 120×40 in-memory terminal.
- `TextPerfHardWrap` / `TextPerfMeasureHardWrap`: painting-side line splitting
  and layout measurement respectively, at width 80. An 80-byte control fits on
  one line; 4 KiB and 16 KiB inputs exercise long unbroken lines.
- `RetainedNestedLayout`: mutation and update of alternating stretched/flexible
  rows and columns, with one unchanged label per level. Depth 3 represents a
  modest nested shell; depth 6 is a stress case, not a measured average app.
- `SignalSelectorFanout`: changing a signal watched by 100 or 1,000 selectors
  whose projected answers remain unchanged. It isolates notification overhead.
  Existing list-cursor benchmarks show the combined changes in an actual update.
- `RetainedPartialDamage/overlap`: a parent and its child both read the changing
  signal while painting. Other cases use disjoint leaves as controls; their
  damage and paint counts should remain unchanged.
- `TreeFlattenVisible`: flattening an expanded, unfiltered tree with ten siblings
  per level, at 110 and 1,110 nodes. It excludes widget creation/layout/painting.
- `Reactivity*`: existing mutation→update workloads and controls. See
  [performance documentation](../../performance.md) for their detailed scope.

Renderer timings exclude terminal I/O, the app event loop and frame throttling,
initial rendering, fixture setup, and benchmark validation. They include signal
notification, invalidation, and the renderer work required by the change.

## Interpretation

The fixes remove repeated work; they do not virtualize collections or change
public APIs. Selector notifications still evaluate all subscribers. Their
allocation **count** falls sharply, but bytes can rise slightly because the Go
allocator rounds the shared buffers to different size classes. Tree traversal
still allocates independent paths. Rich text still materializes graphemes.

The narrow-wide-character hang is a correctness finding, separate from the timing
charts: the baseline width-one CJK regression timed out at 500 ms, while the fixed
painting and layout paths complete. This does not claim to fix every Unicode
width issue in the underlying ANSI library.
