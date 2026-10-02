# Collection widget verification notes

Eight reference pages cover List, Table, Tree, DirectoryTree, TabBar, TabView, Breadcrumbs, KeybindBar, and Jumper.
The 124 records in collections.json document source verification for the page prose, related links, and image descriptions.
Every included Go snippet comes from the compiled examples.go file.
The parent owns SVG generation, visual inspection, integration, and independent claim review.

## Checks

- `GOCACHE=/private/tmp/terma-docs-collections-go-cache go test /private/tmp/terma-docs-work/collections/docs/widget-examples/collections/examples.go` passed.
- `go test . -run 'Test(TreeState|TreeOnMouse|ListState|ListClick|TableState|TabState|NewTabState|TabBar|ReactivityKeybindBar|KeybindBar|Jump)'` passed in the source worktree.
- All evidence source paths exist, referenced symbol tokens occur in the source, and exact claim text occurs in its page.
- The symbol and text check supplements source reading and does not prove claim semantics.

## Implementation concerns

The old List examples supplied an index to RenderItem, but the current signature has no index.
The old List reference named ItemSpacing, which is absent from the current struct.
The old Tree reference omitted the selected-node slice from OnSelect.
The old List and Table pages described Space selection, but their current Keybinds methods contain no Space binding and their OnKey methods return false.
The new pages describe Shift and pointer range selection instead.

DirectoryTree uses deterministic preloaded entries for the image and registered demo.
The page includes both the filesystem constructor snippet and the preview source, with prose explaining the difference.
Jumper uses a preview function that activates jump state before the first render.
The page includes that function and describes the active initial state.

The targeted tests modified generated testdata/snapshot_gallery.html in the worktree.
That generated file is unrelated to this docs partition and should be excluded during integration.
The first example compile passed with the default Go cache, but a later compile met a cache sandbox denial.
The final compile passed with the writable temporary cache shown above.
