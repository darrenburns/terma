# Layout documentation verification

Status: PASS for source review, snippet compilation, and headless rendering.

The partition contains 11 pages, 12 registered examples, and 142 evidence records.
Every prose line and image description has a record in layout.json.
The examples compile against the requested worktree in a temporary Go module.
All 12 examples pass a headless t.SaveSnapshot rendering check.
The rendered SVG text contains the expected labels, including the focused tooltip's help text and the four visible scrollable log lines.
Browser interaction and final image inspection remain with the parent agent.

The existing Row and Column documentation incorrectly identified CrossAxisStretch as the default.
The zero-valued CrossAxisStart constant and toLayoutCrossAlign both select start alignment.
The new page describes that behavior.

FocusTrap's field comment says an active trap requires an ID, but both BuildRenderTree and the retained renderer generate an ID when omitted.
The new page describes the implementation.

No core library files changed.
The first isolated compile used a Go 1.25 directive and failed because the repository requires Go 1.25.5.
The corrected isolated module compiled successfully with Go 1.25.5.
The snapshot check used a temporary GOCACHE after the default cache denied access.

Verification commands:

```sh
go test -mod=mod ./...
GOCACHE=/private/tmp/terma-docs-work/layout/check/cache go test -mod=mod -run TestRenderExamples -v
```

The temporary render test calls t.SaveSnapshot for every record from Examples.
The temporary module, test, and render outputs were removed after verification because the parent owns the committed renderer and image assets.
