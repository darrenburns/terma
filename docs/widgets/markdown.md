# Markdown

`Markdown` is a rendered CommonMark document viewer. Keep its state outside `Build`.

```go
state := terma.NewMarkdownState("# Welcome\n\nRead the [guide](https://example.com).")
scroll := terma.NewScrollState()
view := terma.Scrollable{
    State: scroll,
    Height: terma.Flex(1),
    Child: terma.Markdown{
        ID: "readme", State: state, ScrollState: scroll,
        OnLink: func(destination string) { /* application decides what to open */ },
        OnCopy: func(text string) { terma.SetClipboard(terma.SystemClipboard, text) },
    },
}
```

## State and interaction

- `NewMarkdownState(source)`, `Source()`, `SetSource(source)`, and `Append(fragment)` own the source and parsed document. `Source` subscribes when read in a tracked phase. Set and append reparse the complete document; parsing never happens during painting. Identical replacements are no-ops. Changes clear link and document selection.
- `PlainText()` returns the displayed text with list/quote markers, code whitespace, and logical paragraph breaks, without wrapping or formatting syntax. It subscribes like `Source`.
- `Markdown{ID, State, Style, OnLink, OnCopy, CodeTheme, ScrollState, DisableFocus}` is embeddable. `Style` controls dimensions and outer decoration. Nil state displays nothing. `CodeTheme` optionally names a Chroma style; unknown names use the default code colors. Unknown fence languages remain plain code.
- Left/Right choose the previous/next active link, wrapping at either end. Enter activates the selected link. Mouse clicks activate only link text, including wrapped continuation text. Tab leaves the viewer; scrolling keys bubble to its `Scrollable`. Supplying that container's `ScrollState` also reveals links selected by keyboard, including when Markdown follows siblings or sits inside padded containers. Links can be revealed even when the document starts outside the viewport.
- Link callbacks are the only link side effect; the widget never launches a browser or shell. HTTP(S), mailto, fragment and relative destinations are active when `OnLink` is set. Unsupported schemes, protocol-relative URLs and control-containing destinations display their labels without activation.
- Ctrl+A selects the complete displayed document; Y copies that selection through `OnCopy`, or the terminal clipboard when no callback is set. Escape clears selection. Partial pointer/keyboard text selection is not implemented. Alt+C is also accepted when the terminal sends Alt as Meta; some browsers intercept it. Ctrl+C remains the application-wide quit shortcut.

## Supported Markdown

`Markdown` uses Goldmark to parse CommonMark and supports the following terminal rendering behaviour.

| Content or condition | Rendering |
| --- | --- |
| Empty source or nil state | Empty content, no crash or link activation. |
| Headings, emphasis, inline code | Visible text with theme-aware emphasis; Markdown delimiters disappear. |
| Paragraph soft/hard breaks | Soft breaks become spaces; hard breaks remain newlines. |
| Ordered and nested lists | Preserve ordered start numbers; bullets and hanging indentation distinguish nested items. |
| Quotes and nested quotes | Prefix each wrapped line with a quote marker. |
| Malformed/unclosed inline markup | CommonMark literal fallback; content remains readable. |
| Incomplete fenced code | Remainder is code until closing fence arrives; appending reparses and can change later structure. |
| Long words, URLs and code | Wrap by terminal cell width; code preserves internal whitespace, wrapping long lines. |
| Narrow width including 0/1 cells | Safe clipping; omit excessive indentation to leave space for content. Wide graphemes that cannot fit are clipped. |
| Unicode combining marks / emoji / CJK | Wrap whole grapheme clusters and count terminal cells. |
| HTML and images | Show literal HTML and image alt text; never execute HTML or fetch images. |
| Links without callback / unsafe URL | Readable label, no activation. |
| Link click on blank space, padding, marker | No callback. |
| Duplicate links / wrapped links | Select and activate the correct document occurrence. |
| Source changes while selected | Clear selection, recompute layout and link targets from new content. |
| Disabled subtree | No keyboard/mouse/copy actions; normal Terma disabled handling applies. |
| Source contains terminal controls | Render replacement characters for controls except newline/tab; never emit escape sequences. |

## Limits

This is a document viewer, not an editor or HTML browser. GFM tables, task-list checkboxes, footnotes, embedded media and partial text selection are not implemented. Copy uses logical document text, not the currently wrapped viewport. Source updates reparse the complete source and layout considers the complete document; this is not a virtualized or incremental streaming parser. Construct and mutate state on the UI event loop, as for other composite widget state.

## Demo

Run `go run ./cmd/markdown-demo` to try document rendering, links and copying.
