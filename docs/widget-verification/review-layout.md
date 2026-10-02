# Independent layout documentation review

Verdict: ISSUES.

I reviewed every prose sentence, image description, related-link claim, and included Go snippet in the 11 requested pages against implementation source.
The 142 records in layout.json served as pointers rather than authority.
All page prose and image descriptions have matching evidence records.
The examples compile with `GOCACHE=/private/tmp/terma-docs-collections-go-cache go test ./docs/widget-examples/layout`.

## Findings

1. `docs/widgets/focustrap.md:18` says an inactive trap restores global focus order.
   An inactive inner trap still inherits an enclosing active trap.
   `render_tree.go:57` pushes a scope only for active traps, so an inactive wrapper leaves the current collector scope unchanged.
   `FocusCollector.CurrentTrapID` returns the innermost remaining scope, and `FocusManager.focusablesInScope` constrains cycling to that scope.
   Suggested replacement: "When `Active` is false, the wrapper adds no focus restriction of its own."
   Keep the next sentence about nested active traps.

2. `docs/layout/splitpane.md:20` and `:21` reverse the physical divider orientation.
   `SplitHorizontal` lays out panes side by side and renders a vertical `│` divider in `SplitPane.dividerChar`.
   `SplitVertical` renders the horizontal `─` divider.
   `SplitPane.Keybinds` maps Left and Right to SplitHorizontal and Up and Down to SplitVertical.
   Suggested replacements: "With `SplitHorizontal`, Left and Right, or `h` and `l`, change the divider position in steps of 0.05." and "With `SplitVertical`, Up and Down change the position by the same amount."

3. `docs/floating.md:17` through `:22` describe requested coordinates as the final placement without mentioning screen clamping.
   `Renderer.renderFloats` calls `clampToScreen` after either position calculation, so an AnchorBottomLeft popup near the bottom can overlap its anchor rather than remain below it.
   Suggested adjustment: use "requests placement" for the precise AnchorBottomLeft statement and add "The renderer clamps the final position to keep the overlay on screen."
   Cite `render.go:1514` and `floating.go:438` for the added qualification, using exact current line numbers during editing.

## Coverage

| Page | Records reviewed | Result |
| --- | ---: | --- |
| docs/layout/row-column.md | 12 | Source supports the prose and example. |
| docs/layout/dock.md | 10 | Source supports the prose and example. |
| docs/layout/scrollable.md | 15 | Source supports the prose and example. |
| docs/layout/spacer.md | 9 | Source supports the prose and example. |
| docs/layout/splitpane.md | 14 | Divider wording needs correction. |
| docs/layout/stack.md | 13 | Source supports the prose and example. |
| docs/widgets/focustrap.md | 10 | Inactive nested-trap wording needs correction. |
| docs/widgets/switcher.md | 10 | Source supports the prose and example. |
| docs/widgets/tooltip.md | 12 | Source supports the prose and example. |
| docs/widgets/empty.md | 8 | Source supports the prose and example. |
| docs/floating.md | 29 | Placement needs the screen-clamping qualification. |

The source review included layout.go, layout/linear_node.go, dock.go, layout/dock_node.go, spacer.go, split_pane.go, stack.go, layout/stack_node.go, scroll.go, floating.go, render.go, render_tree.go, focus.go, focus_trap.go, switcher.go, tooltip.go, conditional.go, and the complete layout examples file.
I also inspected the relevant focus-trap, scrolling, and floating-geometry tests.
The existing SVG text and coordinates agree with the image descriptions, including the four visible log lines, tooltip text below Save, and adjacent lines in the EmptyWidget example.
This was source and SVG-content review, not browser interaction verification.

## Tone and snippets

The pages use a consistent factual reference tone with short examples.
I found no unsupported marketing claims, stale callback signatures, placeholder code, or signal writes inside the included Build methods.
The Switcher snippet creates its signal once in its factory and updates it in button callbacks.
The FocusTrap snippet creates its input states once in its factory.
The geometry callback handles a missing or vertically clipped anchor and returns nil as documented.

No documentation or source files were edited during this review.
