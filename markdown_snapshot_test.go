package terma

import "testing"

const markdownSnapshotSource = "# Reading room\n\nA **bold** introduction with *emphasis*, `inline code`, and a [helpful guide](https://example.com/guide).\n\n3. First ordered item\n4. Second item with a longer description that wraps.\n   - Nested bullet with Unicode: café é 👩‍💻 中文\n\n> A quote with **weight** and words that wrap.\n>\n> > Nested thought.\n\n```go\nfunc main() {\n\tprintln(\"Hello, terminal\")\n}\n```\n\n---\n\n![Fox illustration](fox.png) and <b>literal HTML</b>."

func TestSnapshot_Markdown_Blocks(t *testing.T) {
	AssertSnapshot(t, Markdown{State: NewMarkdownState(markdownSnapshotSource), OnLink: func(string) {}, CodeTheme: "monokai"}, 60, 32, "CommonMark blocks, inline styles, nested list and quote hanging indents, syntax-highlighted code, literal HTML and image alt text")
}

func TestSnapshot_Markdown_Narrow(t *testing.T) {
	AssertSnapshot(t, Markdown{State: NewMarkdownState(markdownSnapshotSource), OnLink: func(string) {}}, 23, 52, "Narrow reflow preserves list/quote indentation and whole Unicode graphemes")
}

func TestSnapshot_Markdown_Incomplete(t *testing.T) {
	state := NewMarkdownState("# Stream\n\n*unfinished and [broken](\n\n```go\nlong_identifier_without_spaces := 123456789\n")
	AssertSnapshotNamed(t, "Markdown_open", Markdown{State: state, CodeTheme: "monokai"}, 25, 12, "Unclosed fence remains readable code and long words hard-wrap")
	state.Append("```\n\n**Complete** [guide](guide.md)")
	AssertSnapshotNamed(t, "Markdown_closed", Markdown{State: state, CodeTheme: "monokai", OnLink: func(string) {}}, 25, 15, "Appending a fence terminator reparses subsequent content")
}

func TestSnapshot_Markdown_LinkAndSelection(t *testing.T) {
	state := NewMarkdownState("# Links\n\n[first long link label across lines](https://first.test) and [second](https://second.test)\n\n[unsafe](javascript:alert) stays inert.")
	widget := Markdown{ID: "md", State: state, OnLink: func(string) {}}
	state.activeLink.Set(0)
	AssertSnapshotNamed(t, "Markdown_link", widget, 24, 10, "The active wrapped link is highlighted; unsafe destination is plain text")
	state.selected.Set(true)
	AssertSnapshotNamed(t, "Markdown_selected", widget, 24, 10, "Whole-document selection uses selection colors across visible text")
}

func TestSnapshot_Markdown_SafeAndEmpty(t *testing.T) {
	AssertSnapshotNamed(t, "Markdown_controls", Markdown{State: NewMarkdownState("# Controls\n\nESC: \x1b[31m Bell: \x07 Entity: &#27;\n\n<script>alert('literal')</script>")}, 45, 9, "Terminal controls become replacement characters; HTML is literal text")
	AssertSnapshotNamed(t, "Markdown_empty", Markdown{State: NewMarkdownState("")}, 20, 3, "Empty source remains empty")
	AssertSnapshotNamed(t, "Markdown_one-cell", Markdown{State: NewMarkdownState("- abc\n  - xyz\n\n中文")}, 1, 12, "Decoration yields to content at one cell; unrepresentable wide graphemes clip safely")
}
