# Independent review of display and index documentation

Status: ISSUES, with full source coverage completed.

I reviewed every non-heading prose line, image description, behavior statement, and link description in the nine assigned pages.
The initial reviewed versions contained 140 such units, including navigation links.
The writer's evidence records were pointers only; I read the corresponding implementations independently.
I also checked every local link target and the generated SVG text for the display examples.
No local link target was missing.
Final browser image inspection remains with the parent agent.

## Required corrections

1. `docs/widgets/sparkline.md:20` says that downsampling averages groups of values without qualifying the one-cell case.
   `sparklineResample` in `sparkline.go:274` returns the final value when width is one, before reaching `sparklineDownsample`.
   Replace the sentence with “When the width exceeds one cell, downsampling averages groups of values, and upsampling interpolates between values.”
   Add “At a width of one cell, the sparkline uses the final value.”
   Update the evidence record to cover this branch.

2. `docs/widgets/index.md:66` includes Spacer in the list of widgets that arrange child widgets.
   `Spacer` in `spacer.go` is a leaf with no Child or Children field.
   Remove Spacer from that sentence and describe it separately as reserving empty layout space.

3. `docs/layout/index.md:3` says “Layout widgets arrange, size, or position their children.”
   The page lists Spacer, which has no children.
   Use “Layout containers arrange, size, or position their children.” and identify Spacer as a gap widget, or use “Layout widgets control arrangement and spacing.”
   The Containers heading also includes Spacer, so remove it from that group or rename the heading to “Widgets”.

4. The initial `docs/widgets/index.md` said that the getting-started guide covers module setup.
   `docs/getting-started.md` has no module-setup instructions.
   Its first application and later examples support “walks through a complete application”.
   The parent reports that this correction is already applied.

## Complete coverage

| Page | Initial prose and link units | Source checked |
| --- | ---: | --- |
| docs/widgets/text.md | 11 | text.go Text.Render, textContent, wrapText, alignLine, renderPlain, renderSpans; render.go drawSpan; style.go Style; presented_text.go helpers |
| docs/widgets/presentedtext.md | 12 | presented_text.go all displayed helper implementations and PresentedText layout/paint methods; signal.go Get; retained_renderer.go paint subscriptions; widget.go intrinsicSizeChanged |
| docs/widgets/progressbar.md | 11 | progressbar.go GetContentDimensions and Render; docs/animation.md transitions and animation types |
| docs/widgets/sparkline.md | 12 | sparkline.go GetContentDimensions, Render, sparklineResample, sparklineDownsample, sparklineUpsample |
| docs/widgets/spinner.md | 13 | spinner.go built-in styles, NewSpinnerState, Start, Stop, widestFrame, Render; frame_animation.go constructor, Start, Stop, Value; docs/animation.md |
| docs/widgets/image.md | 16 | image.go NewImageResource, Image layout/render, mapImage, DrawImage, halfBlockCell; image_detection.go protocol selection and kittyColorEncodingSafe; cmd/image-demo/main.go |
| docs/widgets/custom-widgets.md | 12 | widget.go Widget, Renderable, ContainerLayoutBuilder; context.go Theme; signal.go Get; retained_renderer.go Build subscriptions; repository AGENTS.md invalidation guidance; display Greeting example |
| docs/widgets/index.md | 36 | widget.go Widget; app.go Run; example runner flags and registries; complete widget-start program; declarations for every named exported widget; getting-started guide; layout implementations |
| docs/layout/index.md | 17 | dimension.go Auto, Cells, Flex, Percent, DimensionSet; style.go Style; layout/box_model.go content/padding/margin definitions; layout.go Row/Column; all linked layout declarations |

Every example inclusion was checked against `docs/widget-examples/display/examples.go`, including the import alias, constructors, initial state, and returned widget.
The generated SVG text matches the stated deployment message, counter, upload label, requests label, spinner label, and greeting.
The image example constructs an NRGBA gradient and Image.DrawImage draws half-block cells, matching the image description.

All statements not identified above are supported by the inspected source or the linked guide's actual contents.
No additional behavioral correction is required in Text, PresentedText, ProgressBar, Spinner, Image, or Custom widgets.
