# Measuring reactivity performance

The reactivity benchmarks measure a state mutation followed by `Renderer.Update`
against an in-memory terminal buffer. This includes signal notification,
intrinsic-size checks, building, layout, and painting. It excludes terminal I/O,
the application's event loop and frame throttling, initial rendering, fixture
construction, and benchmark validation.

## Run the benchmarks

```sh
# Quick validation: every scenario must produce a frame and change visible output.
go test . -run '^$' -bench '^BenchmarkReactivity' -benchtime=1x

# Collect repeated timing and allocation measurements.
go test . -run '^$' -bench '^BenchmarkReactivity' -benchmem -benchtime=200ms -count=5

# Focus on how one changing leaf scales with tree size.
go test . -run '^$' -bench '^BenchmarkReactivityLeaf$' -benchmem -count=10
```

Run comparisons on the same machine, Go version, power settings, and otherwise
idle system. Save before/after output and compare repeated samples (for example
with `benchstat`, if installed). Do not run CPU profiling while collecting the
timing baseline.

## Scenarios

| Benchmark | Question |
| --- | --- |
| `ReactivityLeaf` | How does changing one leaf scale across 10, 100, and 1,000 leaves, with the signal read during build versus paint? |
| `ReactivityTextSize` | How do fixed-width text, stable auto-width text, and growing/shrinking auto-width text compare in a 100-leaf tree? |
| `ReactivityListCursor` | What does moving a focused list's cursor between the first two items cost with default versus custom item rendering, at 10, 100, and 1,000 items? |
| `ReactivityBatch` | What does 1, 10, or 100 signal writes followed by one update cost, targeting the same leaf or distinct leaves in a 1,000-leaf tree? |

All scenarios use a fixed 120 × 40 viewport. Leaf trees have ten fixed-size leaves
per row; leaves beyond the first 400 are offscreen. Lists have one item per row;
items beyond the first 40 are offscreen. Changing content remains visible. These
are retained-tree scaling measurements, not virtualization benchmarks.

The leaf phase comparison uses identical text and geometry. The list comparison
uses the built-in renderer versus a simple custom text renderer; their styling
and implementation costs differ, so it measures the practical cost of those two
paths rather than isolating read phase alone.

Each scenario alternates between two states, avoiding equal-value `Set` no-ops.
Batch scenarios deliberately perform all writes before calling `Update` once;
they do not measure the event loop's scheduling or channel coalescing.

## Metrics

Go reports `ns/op`, `B/op`, and `allocs/op` for the timed mutation/update loop.
Additional metrics come from two untimed validation updates, one in each
direction, so collecting them does not distort the timed loop:

| Metric | Meaning |
| --- | --- |
| `builds/op` | Widget build calls per update |
| `layouts/op` | Layout computations per update, counting cache misses only; a node measured under several constraints counts once per computation. Before the layout cache this counted nodes receiving a layout. |
| `paints/op` | Retained-node paint visits, including ancestors and repeated visits across damage rectangles |
| `damage-cells/op` | Sum of viewport-clipped damage rectangle areas; full frames count the whole viewport |
| `full/op`, `partial/op` | Full and partial frames per update |
| `sets/op` | Signal writes per update |

Damage area counts overlapping rectangles repeatedly. It is not a count of
unique changed cells, buffer writes, or bytes sent to the terminal. Paint visits
also do not imply that each visited widget changed visibly. Work counters show
the current execution path without asserting that it must remain unchanged as
the renderer improves.

## Initial baseline

Collected on 2026-09-25, Apple M4 Pro, darwin/arm64, Go 1.25.5, default
GOMAXPROCS=14. Renderer source: commit `82c01bc`, with this benchmark suite added.
The command used `-benchtime=200ms -count=5`; the table shows medians of the five
samples. [Raw benchmark output](benchmarks/reactivity-baseline.txt) includes all
21 scenarios, allocations, and work counters. This is a local reference, not a
portable performance threshold.

| Scenario | Time/update | Bytes/update | Allocations/update |
| --- | ---: | ---: | ---: |
| One build-read leaf, 10 leaves | 26.43 µs | 38,053 | 480 |
| One paint-read leaf, 10 leaves | 1.94 µs | 2,445 | 34 |
| One build-read leaf, 100 leaves | 266.00 µs | 367,119 | 4,606 |
| One paint-read leaf, 100 leaves | 2.47 µs | 2,445 | 34 |
| One build-read leaf, 1,000 leaves | 2,111.11 µs | 2,823,254 | 31,427 |
| One paint-read leaf, 1,000 leaves | 7.62 µs | 2,445 | 34 |
| Auto-width text changes size, 100 leaves | 264.83 µs | 366,499 | 4,596 |
| Default list cursor, 1,000 items | 688.74 µs | 715,409 | 10,259 |
| Custom list cursor, 1,000 items | 2,028.68 µs | 3,258,195 | 33,862 |
| 100 writes to one leaf, one update | 12.41 µs | 4,029 | 133 |
| 100 distinct leaves changed, one update | 276.46 µs | 249,495 | 3,308 |

At 1,000 leaves, a build-dependent change rebuilds and lays out 1,101 nodes
(including containers). A paint-dependent change performs zero builds/layouts,
visits three paint nodes, and damages 12 cells, but its runtime still grows with
the retained tree. The partial path scans for damage and clears dirty flags
across the tree.

A separate five-second CPU profile of the 1,000-leaf paint scenario attributed
about 28% of samples cumulatively to `clearDirtyRecursive` and 23% to damage
collection. The allocation profile attributed about 94% of allocated bytes
directly to `DrawStyledText` and `SubContext` combined. These are concrete starting
points for optimization; the profile was collected separately from the timing
baseline above.

The default list cursor benchmark damages every visible row, covering all 4,800
viewport cells at 100 or 1,000 items. Each row reads the shared cursor signal, so
partial mode alone does not mean only the old and new cursor rows were painted.
At 10 items, the simpler custom renderer is slightly faster despite rebuilding;
use measured time and allocations alongside the frame mode.

## Build isolation

Build-dependent updates now rebuild only the dirty widget and its descendants.
Collected on the same machine and settings as the baseline;
[raw output](benchmarks/reactivity-build-isolation.txt).

| Scenario | Baseline | Build isolation | Builds/update |
| --- | ---: | ---: | ---: |
| One build-read leaf, 10 leaves | 26.43 µs | 19.38 µs | 12 → 1 |
| One build-read leaf, 100 leaves | 266.00 µs | 194.03 µs | 111 → 1 |
| One build-read leaf, 1,000 leaves | 2,111.11 µs | 1,400.10 µs | 1,101 → 1 |
| Auto-width text changes size, 100 leaves | 264.83 µs | 193.85 µs | 111 → 0 |
| Custom list cursor, 1,000 items | 2,028.68 µs | 1,774.49 µs | 1,001 → 1,001 |

Skipping builds removes about a third of the time and allocations. The rest of a
build-dependent update is full-tree layout and full-viewport painting, which still
run after any build or layout change. A CPU profile of the 1,000-leaf build
scenario splits `Update` time roughly 40% paint, 32% layout, and 27% walking the
retained tree. Custom list renderers still rebuild every row because the list
itself reads the cursor in `Build()` and passes `active` to each row.

## Partial repaint after relayout

A build or layout change used to clear and repaint the whole screen. These frames
(`reflow` mode) now measure the tree without drawing, recording every node's
position and the hit-test registry, and collect damage:

- the old and new area of every invalidated or rebuilt node's subtree;
- the old and new area of any node that moved, resized, or whose layout box
  changed (for example a `Scrollable` whose content grew, which changes its
  scrollbar).

Only the damaged areas are cleared and repainted. Frames with floating overlays,
forced full renders, and resizes still repaint everything.

## Layout cache

Each retained node caches its last few layout results by constraints. A node
whose subtree has no build or layout changes reuses a cached result for the same
constraints without building its layout node or visiting its children.
Rebuilt nodes are marked changed, since a parent's rebuild can hand them new
properties without any signal of their own changing. On a cache hit a node keeps
its layout-phase signal subscriptions, because the reads that created them don't
run. Forced full renders bypass the cache.

Collected on the same machine and settings as the baseline; medians of five
samples. Raw output: [partial repaint](benchmarks/reactivity-reflow.txt),
[layout cache](benchmarks/reactivity-layout-cache.txt).

| Scenario | Baseline | Build isolation | + partial repaint | + layout cache |
| --- | ---: | ---: | ---: | ---: |
| One build-read leaf, 10 leaves | 26.43 µs | 19.38 µs | 10.3 µs | 7.2 µs |
| One build-read leaf, 100 leaves | 266.00 µs | 194.03 µs | 79.5 µs | 26.3 µs |
| One build-read leaf, 1,000 leaves | 2,111.11 µs | 1,400.10 µs | 757.3 µs | 212.6 µs |
| Auto-width text changes size, 100 leaves | 264.83 µs | 193.85 µs | 91.9 µs | 38.9 µs |
| Custom list cursor, 1,000 items | 2,028.68 µs | 1,774.49 µs | 1,452.4 µs | 1,543.6 µs |

A one-leaf change in a 1,000-leaf tree now builds 1 widget, computes 3 layouts
and repaints 12 cells. What remains is proportional to the tree size: the build
walk (which also collects focusables and floats), assigning layouts, and the
measuring pass. Paint-only updates are unchanged.

## List rows and `Select`

List rows used to subscribe to the whole cursor signal, so every visible row
repainted on each cursor move, and with a custom `RenderItem` every row rebuilt.
Rows now use `Select` (and `SelectAny` for the selection), which notifies a
subscriber only when its projected answer ("am I the active row?") changes. With
a custom `RenderItem`, each row is its own widget that selects its state in its
own `Build`, so only the rows the cursor leaves and enters rebuild. Rows cache
their `SizePreserver` answers alongside cached layouts, so recomputing the list's
layout doesn't rebuild every row's layout node.

[Raw output](benchmarks/reactivity-select.txt), same machine and settings.

| Scenario | Baseline | Before `Select` | Now | Builds | Damage |
| --- | ---: | ---: | ---: | ---: | ---: |
| Default list cursor, 100 items | 434.5 µs | 431.9 µs | 30.4 µs | 0 | 2 rows |
| Default list cursor, 1,000 items | 688.7 µs | 712.9 µs | 76.9 µs | 0 | 2 rows |
| Custom list cursor, 100 items | 531.5 µs | 496.0 µs | 91.9 µs | 4 | 2 rows |
| Custom list cursor, 1,000 items | 2,028.7 µs | 1,543.6 µs | 654.8 µs | 4 | 2 rows |

The remaining cost grows with list length: every row's selector runs on each
cursor change, and with a custom renderer the list's own layout is recomputed,
walking every row (each a cache hit). A virtualized list would avoid both.

## Skipping clean subtrees in every pass

The per-frame passes other than building now follow the dirty path:

- **Dirty flag clearing and damage scanning** descend only into subtrees whose
  dirty flag is set. A node can only be dirty beneath ancestors with the flag.
- **Layout assignment** stops at a clean node handed back exactly the result it
  was last assigned (same box and the same cached child layout slice).
- **The measuring pass** stops at a clean node whose layout was reused and whose
  position is unchanged. It replays the subtree's hit-test entries from a view
  kept on the node; the registry allocates a fresh slice each frame so those
  views stay valid.

The build walk still visits every node, since it also collects focusables,
focus traps and floats in tree order.

[Raw output](benchmarks/reactivity-skip-clean.txt), same machine and settings.

| Scenario | Baseline | Before | Now |
| --- | ---: | ---: | ---: |
| One build-read leaf, 1,000 leaves | 2,111.1 µs | 213.2 µs | 135.5 µs |
| One paint-read leaf, 1,000 leaves | 7.6 µs | 7.3 µs | 3.0 µs |
| Custom list cursor, 1,000 items | 2,028.7 µs | 654.8 µs | 499.1 µs |
| One write, one update (1,000 leaves) | 7.6 µs | 7.5 µs | 3.2 µs |

Paint-only updates no longer depend on tree size.

## Correctness checks

`reactivity_sequence_test.go` drives state changes through a persistent renderer
and compares every incremental frame against an independent forced full render:
cells, focus, focus order, and hit targets. Set `TERMA_REACTIVITY_OUTPUT=<dir>`
to write expected, actual, and diff SVGs plus an HTML gallery for each frame;
failing tests write them to `snapshot-output/reactivity/`.

## CPU and memory profiles

Profile one scenario at a time. The test binary and profiles below go outside the
repository:

```sh
go test . -run '^$' \
  -bench '^BenchmarkReactivityLeaf$/^leaves=1000$/^paint$' \
  -benchtime=5s -count=1 \
  -cpuprofile=/tmp/terma-reactivity.cpu \
  -memprofile=/tmp/terma-reactivity.mem \
  -o /tmp/terma-reactivity.test

go tool pprof -top /tmp/terma-reactivity.test /tmp/terma-reactivity.cpu
go tool pprof -alloc_space -top /tmp/terma-reactivity.test /tmp/terma-reactivity.mem
```

Profiles can include setup and runtime work even though benchmark timings exclude
setup. Use a sufficiently long run to make repeated updates dominate.

## Interactive inspection

Run this in Ghostty or another terminal:

```sh
TERMA_DEBUG_OVERLAY=1 go run ./cmd/invalidation-demo
```

The overlay shows frame duration, frame-budget overruns, full/partial mode, work
counts, damage area, and the last render cause. Try the fixed-width and auto-width
inputs, or use Ctrl+B to toggle a build-dependent banner. Ctrl+G toggles the
spinner and Ctrl+Q exits. The display-cycle duration includes terminal output and
focus/hover reconciliation; it is not input-to-visible-pixel latency.

`Renderer.Stats()` and `CurrentRenderStats()` expose work counts for the most
recent rendered frame. A clean update does not reset those counts, and one
display cycle can contain multiple renderer updates. Use the headless suite for
repeatable comparisons and interactive inspection for behavior in a real
terminal.
