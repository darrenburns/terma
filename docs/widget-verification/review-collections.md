# Independent collections documentation review

Verdict is ISSUES, with three small wording corrections below.
I reviewed every sentence, behavior bullet, related-link claim, image description, and displayed Go snippet in the eight assigned pages.
The review covered all 124 prose and image-alt claims recorded in collections.json.
The JSON records were pointers only; I independently read the implementation for each claim.

## Required corrections

1. `docs/widgets/directorytree.md:28` says EagerLoad "collapses each directory it loads."
   `DirectoryTree.eagerLoad` calls `state.Collapse` only inside `if len(children) > 0` at `directory_tree.go:225`.
   Replace the bullet with "`EagerLoad` recursively loads unloaded descendants in the background and collapses newly loaded directories that have children."
   Update the matching evidence record to name the nonempty-children condition.

2. `docs/widgets/table.md:17` says "`RenderHeader` overrides the headers in `Columns`."
   `Table.Build` uses the column header if `RenderHeader(colIdx)` returns nil at `table.go:974`.
   Replace the bullet with "`RenderHeader` supplies each header, with the column's `Header` as the fallback when the callback returns nil."
   Update the matching evidence record to include that fallback.

3. `docs/widgets/tree.md:21` says a double-click invokes OnSelect without distinguishing the expansion indicator.
   `Tree.OnMouseDown` toggles expansion and returns before activation when the pointer hits an expandable indicator at `tree.go:854`.
   Replace the bullet with "Enter and a double-click on a node's label invoke `OnSelect` with the cursor node and the selected nodes."
   The underlying `Tree.ActivateOnClick` option also changes activation to a plain single left click.
   A separate brief sentence describing that option would make Tree consistent with the List and Table pages, but adding it is optional.

## Coverage

| Page | Claims reviewed | Source read independently |
| --- | ---: | --- |
| List | 15 | `list.go`, `collection_pointer.go`, example source |
| Table | 16 | `table.go`, `collection_pointer.go`, example source |
| Tree | 19 | `tree.go`, `collection_pointer.go`, `directory_tree.go`, example source |
| DirectoryTree | 15 | `directory_tree.go`, shared `Tree` methods, example source |
| Tabs | 16 | `tab.go`, `switcher.go`, example source |
| Breadcrumbs | 11 | `breadcrumbs.go`, example source |
| KeybindBar | 12 | `keybindbar.go`, `context.go`, `focus.go`, `jump.go`, `keybindbar_test.go`, example source |
| Jumper | 20 | `jump.go`, built-in Jump implementations in List, Table, Tree, and Tab, example source |

The remaining claims agree with the implementation.
Source inspection covered constructors, state mutators, keyboard bindings, pointer handlers, render callbacks, filter construction, default styling, and scroll reveal paths as applicable.
The eight SVGs contain the labels named by the image descriptions.
I checked their XML text against the example source; the parent's visual inspection remains responsible for checking their appearance.

## Snippets and tone

All ten displayed snippet regions are present in `docs/widget-examples/collections/examples.go`.
The snippets use real exported APIs, allocate state outside Build, and contain no placeholders or uncompiled pseudocode.
DirectoryTree and Jumper explicitly distinguish their ordinary example from their deterministic image preview.
Their preview functions match the registered example factories.
The snippets are short, and the tone consistently describes behavior in direct present-tense statements.
The pages avoid claims of ease and show the setup through code instead.
No style rewrite is needed.

## Checks run

- PASS. `go test ./docs/widget-examples/collections` compiled the example package.
- PASS. `go test . -run 'Test(Jump|List|Table|Tree|DirectoryTree|Tab|Breadcrumb|KeybindBar|ReactivityKeybindBar|Snapshot_(List|Table|Tree|DirectoryTree|Tab|Breadcrumb|KeybindBar))' -count=1` passed existing behavioral and snapshot tests.
- PASS. A mechanical inventory found an exact evidence record for every one of the 124 prose and alt-text claims.
- No pages or evidence files were modified by this review.
