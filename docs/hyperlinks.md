# Terminal hyperlinks

Terma writes linked text as OSC 8 hyperlinks. A terminal that supports OSC 8
opens the URL on its own link gesture, usually Cmd-click or Ctrl-click. Other
terminals ignore the sequence and show the text unchanged.

## Linking text

Set `SpanStyle.Link` to the URL. `LinkSpan` builds an underlined span with a
link:

```go
theme := ctx.Theme()
terma.Text{Spans: []terma.Span{
	terma.PlainSpan("Read "),
	terma.LinkSpan("the guide", "https://example.com/guide", theme.Link),
	terma.PlainSpan(" first."),
}}
```

`Link` is a field of `SpanStyle`, so it survives everything that carries a
span's style. A link that wraps onto several lines is a link on every line.
`TextInput` and `TextArea` highlights ignore `Link`, because input text is not
linked.

## Markup

The `link=` token links the text in a markup tag:

```go
terma.ParseMarkupToText("See [link=https://example.com/guide]the guide[/].", theme)
```

A link tag underlines its text in the theme's `Link` color. Name a foreground
color in the same tag to use another color. An empty `[link=]` is ignored.

```go
"[link=https://example.com/status $Error]Status page[/]"
```

The URL ends at the first space or `]`. Percent-encode those characters in the
URL. Tags nested inside a link tag keep the link. `$Link` is also a markup
color.

## Markdown

`Markdown` links absolute `http`, `https` and `mailto` destinations and shows
them in the theme's `Link` color. The terminal cannot resolve relative
destinations and fragments, so those are not written as OSC 8. They are only
active through `OnLink`. A disabled `Markdown` writes no terminal links.

## Opening links from the app

Terma doesn't open a URL when the user clicks linked text. The terminal owns the
link gesture. An app-level click handler would open the link a second time in
terminals that both open the link and report the click. To open a link from a
key binding or a button, pass the URL to [`OpenURL`](open-url.md):

```go
{Key: "o", Name: "Open", Action: func() {
	go func() { _ = terma.OpenURL(url) }()
}}
```

Some terminals need Shift held to click a link while the app captures the mouse.

## URL handling

Terma doesn't validate the URL scheme. The URL goes to the terminal as written,
with these changes:

- Spaces and non-ASCII bytes are percent-encoded, because OSC 8 allows only
  printable ASCII.
- A URL that contains a control character is dropped and the text is shown
  without a link. A control character could end the escape sequence early.

## Rendering

Each cell stores its link. The renderer opens a link before the first cell of a
run and closes it before the next cell that has another link or no link.
Repainting part of a link reopens it, so incremental frames never let a link
spill into neighboring cells. A modal backdrop removes links from the content it
covers.

`BufferToANSI` and `RenderToString` write OSC 8 around linked runs. Snapshot SVGs
wrap linked text in `<a href="...">`, so a golden file records which text links
where. Snapshot cell comparisons include the link.

## Example

`go run ./cmd/hyperlink-demo` shows links in markup, Markdown and a list. It is
also in `go run ./cmd/terma-demos`.
