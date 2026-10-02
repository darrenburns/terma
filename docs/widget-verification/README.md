# Widget documentation verification

The reviewed documentation targets the widget implementations on `main` at `cf3635e8c56843a425048aaa4b708d47564e8416`.
It includes the draft from the `widget-docs` worktree and corrections from three independent section reviewers.

The checker inventories documentation for 41 exported widget types.
Of the 40 widget and layout pages, 36 have per-line evidence records covering 593 prose, navigation, and image-description entries.
The four additional guides, SelectBox, Form and Field, FilePicker, and Markdown, were source-reviewed separately. The checker verifies their presence but does not claim per-line evidence coverage for them.
The home page and local documentation instructions were also reviewed.

Each evidence record names implementation source and records its SHA-256 hash.
A changed source file requires another source review before its hash is updated.
A matching hash and evidence record do not establish that a claim is true. The reviewers checked the meaning against source and behavioral tests.

## Repeat the checks

From the repository root:

```sh
python3 scripts/verify-widget-docs.py
go test ./...
uv run --project docs mkdocs build --strict
```

The checker validates the widget inventory, evidence records, source hashes, local links, and snippet regions for the 36 evidence-backed pages.
It compiles the shared widget examples and both complete introductory programs, then compares all 36 rendered SVGs byte for byte.
MkDocs includes those Go sources directly, so the displayed examples and compiled examples share one source.

To regenerate images after an intentional example change:

```sh
go run ./docs/widget-examples -render
```

To preview the documentation:

```sh
uv run --project docs mkdocs serve
```

Open http://127.0.0.1:8000. Run MkDocs from the repository root so snippet paths resolve.

## Review results

| Section | Result | Main corrections |
| --- | --- | --- |
| Inputs and focus, 10 pages | PASS | State lifetime, Autocomplete child requirements, callback behavior, palette focus and search, text editing modes |
| Collections and navigation, 8 pages | PASS | Table sorting and column controls, stable row identity, source indices, tree lazy loading, selection anchors, shortcuts |
| Display and layout, 17 pages | PASS | Explicit Auto sizing for reactive text, computed-text measurement, layout dimensions, modal focus, state lifetime |
| Four additional guides | PASS | SelectBox Space behavior, form setup, FilePicker qualifiers and confirmation button, Markdown no-op updates |
| Overview, home page, navigation, and run instructions | PASS | Links to all widgets, compiled homepage example, accurate verification scope |

The review corrected the homepage's `*Signal[int]` field to `Signal[int]` and moved the complete program into compiled source.
The reactive counter example now sets an explicit Auto width so two-digit values remain visible.
The table controls example is available as `go run ./docs/widget-examples -widget table-controls`.

## Verification in this review

- The full Go suite passes, including behavioral and snapshot tests.
- The documentation checker passes for 593 claims and 36 SVGs.
- MkDocs passes a strict build.
- Browser inspection confirms a widget page shows its SVG, navigation, and expanded Go snippet at the default viewport.
- In the reactive counter example, ten Enter presses display `Count: 10`, and a mouse click displays `Count: 11` without clipping.
- In the table controls example, clicking Name sorts ascending, Ctrl+S sorts descending, and Ctrl+Right increases the active column width.

Live examples used `terma-browser` at 60 columns and 15 rows.
These checks do not certify every interaction in every terminal emulator. Native Kitty and Sixel image protocols were not exercised.

## Earlier draft evidence

The other reports, logs, and PNG captures in this directory record the earlier draft at `5728b68d01669f8e99cc9621b6d79463d8e16f93`.
Their counts and screenshots describe that earlier pass. The current scope and results are listed above.
