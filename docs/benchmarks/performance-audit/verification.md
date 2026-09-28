# Verification record

Completed on 28 September 2026. Final automated checks and measurements use
`ad4bdcb` plus the delivered performance fixes.

## Automated checks

- `go test ./...`: passed, including all snapshot comparisons and browser-bridge
  loopback tests. No golden snapshots were changed.
- `go test -race ./...`: passed across all packages.
- Selector snapshot regression: baseline captured 100 subscribers with 101
  allocations; fixed implementation passes the bounded-allocation test. Separate
  coverage verifies snapshots survive subscription removal and compaction.
- Nested-layout regression: baseline repeatedly measured identical constraints;
  fixed implementation computes each once in the tested initial, changed, and
  forced frames, while updating the leaf's measured size.
- Parent/child damage regression: baseline painted twice; fixed incremental output
  matches a forced full render with one visit per node. A scrollbar-plus-content
  regression verifies content is still updated alongside scrollbar-only damage.
- Rich-text differential check: 2,000 seeded cases matched the original output
  across wrapping modes, styles, Unicode, heights, and widths. This used a
  temporary copy of the original functions, removed after comparison.
- Existing text snapshots: 42 passed unchanged.
- `go test ./internal/textutil -run '^$' -fuzz '^FuzzHardWrap$' -fuzztime=3s -parallel=2`:
  65,158 fuzz executions passed, checking termination and byte preservation.
- Baseline `TestTextWrappingNarrowWideGrapheme` timed out at 500 ms on width-one
  CJK input. The fixed root and layout paths pass the same regression.
- Tree coverage checks preorder, collapsed branches, expansion state, and
  independent path slices.
- Benchmark validation: all 39 scenarios completed with five before and five
  after samples. Renderer workloads assert visible output changes per update.
- `git diff --check`: passed.

## Real terminal apps in the browser

Used `go run ./cmd/terma-browser -- go run <demo>`, inspecting screenshots and
accessible terminal text. Tabs were closed and launchers stopped after use.

- Final gallery smoke test on `ad4bdcb`: list cursor movement, range selection,
  appending ten rows, mouse selection, tree lazy loading, navigation, and collapse.
  Inspected screenshots and terminal text at the default viewport.

The following focused checks ran with the same performance fixes before the final
upstream refresh (on `ef63bc1`); the full test and race suites were rerun afterward.

- Todo demo: inspected the default viewport and fixed 100×30 terminal; entered
  a task, navigated the cursor, toggled completion styling, and opened the help
  dialog. Persistence used isolated `XDG_STATE_HOME` under `/private/tmp`.
- Markup demo, 100×30: inspected bold, italic, underline, foreground/background
  colors and combined styles; changed the theme and scrolled with the mouse.
- Directory tree fixture, default 1280×720 browser: mouse focus, nested expansion
  and collapse, navigation skipping collapsed descendants, range selection,
  filtering to a descendant with its ancestor chain, and restoring the view.

## HTML artifact

- Inspected desktop (1280 px), mobile (390 px), and narrow (320 px) layouts.
- At 320 px, the document had no horizontal overflow; the wide evidence table
  scrolls within its own labeled region.
- Both workload settings render six comparisons with the correct measurements.
- Searching `HardWrap` returns six of 39 scenarios.
- Browser reported no console errors or warnings during the check.
- Embedded measurements exactly match `results.json`; all source/evidence links
  resolve to the repository or bundled data. Published source and Markdown links
  point to the audited commit on GitHub. No external assets or runtime dependencies.
