package terma

import (
	"fmt"
	"html"
	"net/url"
	"strings"
	"unicode"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/darrenburns/terma/layout"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gmtext "github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// MarkdownState owns an immutable parsed document and the viewer's selection.
// Construct it with NewMarkdownState and update it outside Build.
type MarkdownState struct {
	document   Signal[*markdownDocument]
	activeLink Signal[int]
	selected   Signal[bool]
}

// NewMarkdownState parses source for a Markdown viewer.
func NewMarkdownState(source string) *MarkdownState {
	return &MarkdownState{document: NewSignal(parseMarkdown(source)), activeLink: NewSignal(-1), selected: NewSignal(false)}
}

// Source returns the source, subscribing when called in a tracked phase.
func (s *MarkdownState) Source() string { return s.document.Get().source }

// SetSource replaces the document and clears selection. Identical source is a no-op.
// Like other composite widget state, mutate it on the UI event loop.
func (s *MarkdownState) SetSource(source string) {
	if source == s.document.Peek().source {
		return
	}
	s.activeLink.Set(-1)
	s.selected.Set(false)
	s.document.Set(parseMarkdown(source))
}

// Append appends a fragment and reparses the document, including incomplete fences.
func (s *MarkdownState) Append(fragment string) { s.SetSource(s.document.Peek().source + fragment) }

// PlainText returns logical rendered text, without viewport wrapping or Markdown syntax.
// It subscribes when called in a tracked phase.
func (s *MarkdownState) PlainText() string { return s.document.Get().plainText() }

// Markdown displays CommonMark headings, paragraphs, lists, quotes, code and links.
// Compose with Scrollable for viewport scrolling. Absolute http, https and mailto
// links are written as OSC 8 terminal hyperlinks, which the terminal opens on its
// own link gesture. Keyboard and click activation never opens a URL itself; it
// calls OnLink, which can pass the destination to OpenURL. Ctrl+A selects all
// text; Y copies it, and Escape clears selection.
type Markdown struct {
	ID           string
	State        *MarkdownState
	Style        Style
	ScrollState  *ScrollState // Optional: share the containing Scrollable's state to reveal keyboard-selected links.
	CodeTheme    string       // Optional Chroma style name for fenced code syntax highlighting.
	DisableFocus bool
	OnLink       func(destination string)
	OnCopy       func(text string) // Nil uses the terminal system clipboard.
}

// Build creates a width-aware view of the already parsed document.
func (m Markdown) Build(ctx BuildContext) Widget {
	if m.State == nil {
		return EmptyWidget{}
	}
	id := m.ID
	if id == "" {
		id = ctx.AutoID()
	}
	view := &markdownView{owner: m, document: m.State.document.Get(), id: id, theme: ctx.Theme(), disabled: ctx.IsDisabled(), reveal: ctx.scrollToView}
	view.blocks = append([]markdownBlock(nil), view.document.blocks...)
	for i := range view.blocks {
		view.blocks[i].runs = markdownCodeRuns(view.blocks[i], m.CodeTheme)
	}
	view.focused = ctx.IsFocused(view)
	return view
}

type markdownRun struct {
	text  string
	style SpanStyle
	link  int // -1 for ordinary text
	code  bool
}

type markdownBlock struct {
	runs                    []markdownRun
	firstPrefix, nextPrefix string
	gap                     bool
	code                    bool
	rule                    bool
	language                string
}

type markdownDocument struct {
	source string
	blocks []markdownBlock
	links  []string
}

func (d *markdownDocument) plainText() string {
	var out strings.Builder
	for i, b := range d.blocks {
		if i > 0 {
			out.WriteByte('\n')
			if b.gap {
				out.WriteByte('\n')
			}
		}
		var content strings.Builder
		for _, r := range b.runs {
			content.WriteString(r.text)
		}
		for j, line := range strings.Split(content.String(), "\n") {
			if j > 0 {
				out.WriteByte('\n')
				out.WriteString(b.nextPrefix)
			} else {
				out.WriteString(b.firstPrefix)
			}
			out.WriteString(line)
		}
	}
	return out.String()
}

// markdownSafeText removes terminal control characters, including decoded entities.
func markdownSafeText(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return '\uFFFD'
		}
		return r
	}, value)
}

func markdownDecode(value []byte) string {
	return markdownSafeText(html.UnescapeString(string(util.UnescapePunctuations(value))))
}

func markdownSafeLink(destination string) bool {
	if destination == "" || strings.HasPrefix(destination, "//") || strings.Contains(destination, "\\") {
		return false
	}
	if strings.IndexFunc(destination, func(r rune) bool { return unicode.IsControl(r) || unicode.IsSpace(r) || r == '\uFFFD' }) >= 0 {
		return false
	}
	parsed, err := url.Parse(destination)
	if err != nil {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "":
		return parsed.Host == ""
	case "http", "https":
		return parsed.Host != ""
	case "mailto":
		return parsed.Opaque != "" || parsed.Path != ""
	default:
		return false
	}
}

// markdownAbsoluteLink reports whether a destination accepted by
// markdownSafeLink is absolute, so a terminal can open it as an OSC 8 link.
func markdownAbsoluteLink(destination string) bool {
	parsed, err := url.Parse(destination)
	return err == nil && parsed.Scheme != ""
}

func parseMarkdown(source string) *markdownDocument {
	d := &markdownDocument{source: source}
	normalized := strings.ReplaceAll(strings.ReplaceAll(source, "\r\n", "\n"), "\r", "\n")
	data := []byte(markdownSafeText(normalized))
	root := goldmark.New().Parser().Parse(gmtext.NewReader(data))
	var inline func(ast.Node, SpanStyle, int, bool) []markdownRun
	inline = func(node ast.Node, style SpanStyle, link int, code bool) []markdownRun {
		var runs []markdownRun
		appendText := func(value string) {
			if value != "" {
				runs = append(runs, markdownRun{text: value, style: style, link: link, code: code})
			}
		}
		switch n := node.(type) {
		case *ast.Text:
			value := string(n.Segment.Value(data))
			if !code && !n.IsRaw() {
				value = markdownDecode(n.Segment.Value(data))
			}
			appendText(value)
			if n.HardLineBreak() {
				appendText("\n")
			} else if n.SoftLineBreak() {
				appendText(" ")
			}
			return runs
		case *ast.String:
			value := string(n.Value)
			if !n.IsRaw() && !code {
				value = markdownDecode(n.Value)
			}
			appendText(value)
			return runs
		case *ast.Emphasis:
			if n.Level == 2 {
				style.Bold = true
			} else {
				style.Italic = true
			}
		case *ast.CodeSpan:
			code = true
			// CommonMark code spans normalize newlines to spaces and trim one matching
			// outer space when the content is not entirely spaces.
			var raw strings.Builder
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				if t, ok := child.(*ast.Text); ok {
					raw.Write(t.Segment.Value(data))
				}
			}
			value := strings.ReplaceAll(raw.String(), "\n", " ")
			if strings.HasPrefix(value, " ") && strings.HasSuffix(value, " ") && strings.TrimSpace(value) != "" {
				value = value[1 : len(value)-1]
			}
			appendText(value)
			return runs
		case *ast.Link:
			destination := markdownDecode(n.Destination)
			if markdownSafeLink(destination) {
				link = len(d.links)
				d.links = append(d.links, destination)
			}
		case *ast.AutoLink:
			destination := string(n.URL(data))
			if n.AutoLinkType == ast.AutoLinkEmail {
				destination = "mailto:" + destination
			}
			if markdownSafeLink(destination) {
				link = len(d.links)
				d.links = append(d.links, destination)
			}
			appendText(markdownDecode(n.Label(data)))
			return runs
		case *ast.Image:
			appendText("[image: ")
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				runs = append(runs, inline(child, style, -1, false)...)
			}
			appendText("]")
			return runs
		case *ast.RawHTML:
			appendText(string(n.Segments.Value(data)))
			return runs
		}
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			runs = append(runs, inline(child, style, link, code)...)
		}
		return runs
	}
	var blocks func(ast.Node, string, string, bool)
	blocks = func(node ast.Node, first, next string, gap bool) {
		switch n := node.(type) {
		case *ast.Document, *ast.ListItem:
			_, isItem := node.(*ast.ListItem)
			if isItem && node.FirstChild() == nil {
				d.blocks = append(d.blocks, markdownBlock{firstPrefix: first, nextPrefix: next, gap: gap})
				return
			}
			initial := true
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				prefix := next
				if initial {
					prefix = first
				}
				if _, nested := child.(*ast.List); nested && initial && isItem {
					d.blocks = append(d.blocks, markdownBlock{firstPrefix: first, nextPrefix: next, gap: gap})
					prefix = next
					gap = false
				}
				childGap := !initial
				if _, item := node.(*ast.ListItem); item {
					if list, ok := node.Parent().(*ast.List); ok && list.IsTight {
						childGap = false
					}
				}
				blocks(child, prefix, next, gap && initial || childGap)
				initial = false
			}
		case *ast.List:
			index := n.Start
			if !n.IsOrdered() {
				index = 1
			}
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				marker := "• "
				if n.IsOrdered() {
					marker = fmt.Sprintf("%d. ", index)
				}
				base := next
				if child == n.FirstChild() {
					base = first
				}
				// A list nested directly in an item follows the parent's hanging indent.
				blocks(child, base+marker, next+strings.Repeat(" ", ansi.StringWidth(marker)), gap && child == n.FirstChild() || !n.IsTight && child != n.FirstChild())
				index++
			}
		case *ast.Blockquote:
			initial := true
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				prefix := next
				if initial {
					prefix = first
				}
				blocks(child, prefix+"│ ", next+"│ ", gap && initial || !initial)
				initial = false
			}
		case *ast.Paragraph, *ast.TextBlock, *ast.Heading:
			style := SpanStyle{}
			if _, ok := node.(*ast.Heading); ok {
				style.Bold = true
			}
			d.blocks = append(d.blocks, markdownBlock{runs: inline(node, style, -1, false), firstPrefix: first, nextPrefix: next, gap: gap})
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			value := strings.TrimSuffix(string(node.Lines().Value(data)), "\n")
			language := ""
			if f, ok := node.(*ast.FencedCodeBlock); ok {
				language = string(f.Language(data))
			}
			d.blocks = append(d.blocks, markdownBlock{runs: []markdownRun{{text: value, link: -1, code: true}}, firstPrefix: first, nextPrefix: next, gap: gap, code: true, language: language})
		case *ast.ThematicBreak:
			d.blocks = append(d.blocks, markdownBlock{runs: []markdownRun{{text: "────────", link: -1}}, firstPrefix: first, nextPrefix: next, gap: gap, rule: true})
		case *ast.HTMLBlock:
			value := string(n.Lines().Value(data))
			if n.HasClosure() {
				value += string(n.ClosureLine.Value(data))
			}
			d.blocks = append(d.blocks, markdownBlock{runs: []markdownRun{{text: strings.TrimSuffix(value, "\n"), link: -1}}, firstPrefix: first, nextPrefix: next, gap: gap})
		}
	}
	blocks(root, "", "", false)
	return d
}

type markdownGlyph struct {
	text  string
	width int
	style SpanStyle
	link  int
}
type markdownLine struct {
	glyphs []markdownGlyph
	width  int
	code   bool
}
type markdownView struct {
	owner             Markdown
	document          *markdownDocument
	id                string
	theme             ThemeData
	focused, disabled bool
	lines             []markdownLine
	blocks            []markdownBlock
	reveal            func(*ScrollState, int, int) bool
}

func (v *markdownView) WidgetID() string  { return v.id }
func (v *markdownView) IsFocusable() bool { return !v.owner.DisableFocus && !v.disabled }
func (v *markdownView) Build(ctx BuildContext) Widget {
	v.reveal = ctx.scrollToView
	return v
}
func (v *markdownView) GetStyle() Style { return v.owner.Style }
func (v *markdownView) GetContentDimensions() (Dimension, Dimension) {
	return v.owner.Style.Width, v.owner.Style.Height
}

func (v *markdownView) BuildLayoutNode(BuildContext) layout.LayoutNode {
	style := v.owner.Style
	padding, border := toLayoutEdgeInsets(style.Padding), borderToEdgeInsets(style.Border)
	dims := GetWidgetDimensionSet(v)
	minW, maxW, minH, maxH := dimensionSetToMinMax(dims, padding, border)
	node := layout.LayoutNode(&layout.BoxNode{Padding: padding, Border: border, Margin: toLayoutEdgeInsets(style.Margin), MinWidth: minW, MaxWidth: maxW, MinHeight: minH, MaxHeight: maxH,
		MeasureFunc: func(c layout.Constraints) (int, int) {
			v.lines = v.wrap(c.MaxWidth)
			width := 0
			for _, line := range v.lines {
				width = max(width, line.width)
			}
			return width, len(v.lines)
		},
	})
	if hasPercentMinMax(dims) {
		node = &percentConstraintWrapper{child: node, minWidth: dims.MinWidth, maxWidth: dims.MaxWidth, minHeight: dims.MinHeight, maxHeight: dims.MaxHeight, padding: padding, border: border}
	}
	return node
}

// Keybinds exposes the viewer's available actions to KeybindBar.
func (v *markdownView) Keybinds() []Keybind {
	if v.disabled {
		return nil
	}
	s := v.owner.State
	document := s.document.Get() // KeybindBar must refresh when the document gains or loses links.
	bindings := []Keybind{{Key: "ctrl+a", Name: "Select all", Action: func() { s.selected.Set(true) }}}
	if s.selected.Get() {
		copySelection := func() {
			content := s.document.Peek().plainText()
			if v.owner.OnCopy != nil {
				v.owner.OnCopy(content)
			} else {
				SetClipboard(SystemClipboard, content)
			}
		}
		bindings = append(bindings,
			Keybind{Key: "escape", Name: "Clear selection", Action: func() { s.selected.Set(false) }},
			Keybind{Key: "y", Name: "Copy", Action: copySelection},
			Keybind{Key: "alt+c", Name: "Copy", Hidden: true, Action: copySelection})
	}
	if v.owner.OnLink != nil && len(document.links) > 0 {
		bindings = append(bindings,
			Keybind{Key: "right", Name: "Next link", Action: func() { v.selectLink(1) }},
			Keybind{Key: "left", Name: "Previous link", Action: func() { v.selectLink(-1) }})
		if s.activeLink.Get() >= 0 {
			bindings = append(bindings, Keybind{Key: "enter", Name: "Open link", Action: func() { v.activate(s.activeLink.Peek()) }})
		}
	}
	return bindings
}

func (v *markdownView) OnKey(KeyEvent) bool { return false }

func (v *markdownView) selectLink(direction int) {
	s := v.owner.State
	count := len(v.document.links)
	next := s.activeLink.Peek()
	if next < 0 && direction < 0 {
		next = 0
	}
	next = (next + direction + count) % count
	s.activeLink.Set(next)
	s.selected.Set(false)
	if v.owner.ScrollState != nil {
		inset := v.owner.Style.Padding.Top + borderToEdgeInsets(v.owner.Style.Border).Top
		for y, line := range v.lines {
			for _, g := range line.glyphs {
				if g.link == next {
					if v.reveal != nil && v.reveal(v.owner.ScrollState, y, 1) {
						return
					}
					v.owner.ScrollState.ScrollToView(y+inset, 1)
					return
				}
			}
		}
	}
}

func (v *markdownView) activate(link int) bool {
	if v.disabled || v.owner.OnLink == nil || link < 0 || link >= len(v.document.links) {
		return false
	}
	v.owner.OnLink(v.document.links[link])
	return true
}

func (v *markdownView) OnClick(event MouseEvent) {
	if v.disabled || event.Button != uv.MouseLeft {
		return
	}
	if !v.owner.DisableFocus {
		RequestFocus(v.id)
	}
	// Mouse coordinates include border and padding; layout lines do not.
	border := borderToEdgeInsets(v.owner.Style.Border)
	x := event.LocalX - v.owner.Style.Padding.Left - border.Left
	y := event.LocalY - v.owner.Style.Padding.Top - border.Top
	if x < 0 || y < 0 || y >= len(v.lines) {
		return
	}
	for _, g := range v.lines[y].glyphs {
		if x < g.width {
			if g.link >= 0 && v.owner.OnLink != nil {
				v.owner.State.activeLink.Set(g.link)
				v.owner.State.selected.Set(false)
				v.activate(g.link)
			}
			return
		}
		x -= g.width
	}
}

func (v *markdownView) Render(ctx *RenderContext) {
	// Layout can shrink an Auto-sized view after measuring. Reflow at actual width.
	v.lines = v.wrap(ctx.Width)
	selected := v.owner.State.selected.Get()
	active := v.owner.State.activeLink.Get()
	for y, line := range v.lines {
		if y >= ctx.Height {
			break
		}
		if !ctx.IsVisible(ctx.X, ctx.Y+y) && (ctx.Y+y < ctx.clip.Y || ctx.Y+y >= ctx.clip.Y+ctx.clip.Height) {
			continue
		}
		base := v.owner.Style
		if base.ForegroundColor == nil {
			base.ForegroundColor = v.theme.Text
		}
		if line.code {
			base.BackgroundColor = v.theme.Surface
		}
		if selected {
			base.BackgroundColor = v.theme.Selection
			base.ForegroundColor = v.theme.SelectionText
		}
		ctx.DrawStyledText(0, y, strings.Repeat(" ", ctx.Width), base)
		x := 0
		for _, g := range line.glyphs {
			style := g.style
			if selected || v.focused && active >= 0 && g.link == active && v.owner.OnLink != nil {
				style.Background = v.theme.Selection
				style.Foreground = v.theme.SelectionText
				if !selected {
					style.Background = v.theme.ActiveCursor
				}
			}
			if v.disabled {
				style.Foreground = v.theme.TextDisabled
				style.Link = ""
			}
			ctx.DrawSpan(x, y, Span{Text: g.text, Style: style}, base)
			x += g.width
		}
	}
}
