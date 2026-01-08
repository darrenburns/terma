package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeasureText(t *testing.T) {
	t.Run("EmptyContent", func(t *testing.T) {
		w, h := MeasureText("", WrapNone, 100)
		assert.Equal(t, 0, w)
		assert.Equal(t, 0, h)
	})

	t.Run("SingleLine_NoWrap", func(t *testing.T) {
		w, h := MeasureText("Hello", WrapNone, 100)
		assert.Equal(t, 5, w)
		assert.Equal(t, 1, h)
	})

	t.Run("SingleLine_ExceedsWidth_NoWrap", func(t *testing.T) {
		// Without wrapping, text can exceed maxWidth
		w, h := MeasureText("Hello World", WrapNone, 5)
		assert.Equal(t, 11, w, "no wrap means width can exceed maxWidth")
		assert.Equal(t, 1, h)
	})

	t.Run("MultiLine_ExplicitNewlines", func(t *testing.T) {
		w, h := MeasureText("Line 1\nLine 2\nLine 3", WrapNone, 100)
		assert.Equal(t, 6, w, "widest line is 'Line 1' = 6 chars")
		assert.Equal(t, 3, h)
	})

	t.Run("WrapWord_SingleWrap", func(t *testing.T) {
		// "Hello World" (11 chars) wrapping at width 6
		// Should wrap to: "Hello" and "World"
		w, h := MeasureText("Hello World", WrapWord, 6)
		assert.Equal(t, 5, w, "widest line is 'Hello' or 'World' = 5 chars")
		assert.Equal(t, 2, h)
	})

	t.Run("WrapWord_LongWord_FallsBackToChar", func(t *testing.T) {
		// WrapWord uses a hybrid approach: it wraps at word boundaries when possible,
		// but falls back to character-level breaks for words longer than maxWidth.
		// This matches CSS word-wrap: break-word behavior.
		w, h := MeasureText("Supercalifragilistic", WrapWord, 5)
		assert.LessOrEqual(t, w, 5, "no line should exceed maxWidth")
		assert.Equal(t, 4, h, "20-char word at width 5 = 4 lines")
	})

	t.Run("WrapChar_ExactBreaks", func(t *testing.T) {
		// "HelloWorld" (10 chars) wrapping at width 4
		// Should wrap to: "Hell", "oWor", "ld"
		w, h := MeasureText("HelloWorld", WrapChar, 4)
		assert.Equal(t, 4, w, "max width per line is 4")
		assert.Equal(t, 3, h, "10 chars / 4 = 3 lines")
	})

	t.Run("UnboundedWidth", func(t *testing.T) {
		// maxWidth <= 0 means unbounded
		w, h := MeasureText("This is a very long line that should not wrap", WrapWord, 0)
		assert.Equal(t, 45, w, "entire line on one row")
		assert.Equal(t, 1, h)
	})
}

func TestTextNode_BasicLayout(t *testing.T) {
	t.Run("SimpleText", func(t *testing.T) {
		node := &TextNode{
			Content: "Hello",
		}
		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 5, result.Box.Width)
		assert.Equal(t, 1, result.Box.Height)
		assert.Nil(t, result.Children, "TextNode is a leaf")
	})

	t.Run("MultiLineText", func(t *testing.T) {
		node := &TextNode{
			Content: "Line 1\nLine 2",
		}
		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 6, result.Box.Width, "widest line")
		assert.Equal(t, 2, result.Box.Height, "2 lines")
	})

	t.Run("TextWithWrapWord", func(t *testing.T) {
		node := &TextNode{
			Content: "Hello World",
			Wrap:    WrapWord,
		}
		// Constrain to width 6, should wrap to "Hello" and "World" (both 5 chars)
		result := node.ComputeLayout(Loose(6, 100))

		// Box should hug content (5), not expand to fill constraint (6)
		assert.Equal(t, 5, result.Box.Width, "should hug content, not fill constraint")
		assert.Equal(t, 2, result.Box.Height, "wrapped to 2 lines")
	})
}

func TestTextNode_WithInsets(t *testing.T) {
	t.Run("EmptyContentWithPadding", func(t *testing.T) {
		// Empty content with padding should produce a padding-only box,
		// not collapse to 0x0. This is important for placeholder elements.
		node := &TextNode{
			Content: "",
			Padding: EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		}
		result := node.ComputeLayout(Loose(100, 100))

		// Content: 0x0, Padding: 2x2
		assert.Equal(t, 2, result.Box.Width, "padding-only width")
		assert.Equal(t, 2, result.Box.Height, "padding-only height")
		assert.Equal(t, 0, result.Box.ContentWidth(), "no content")
		assert.Equal(t, 0, result.Box.ContentHeight(), "no content")
	})

	t.Run("Padding", func(t *testing.T) {
		node := &TextNode{
			Content: "Hi", // 2 chars wide, 1 line
			Padding: EdgeInsets{Top: 1, Right: 2, Bottom: 1, Left: 2},
		}
		result := node.ComputeLayout(Loose(100, 100))

		// Content: 2x1
		// + Padding: 2+4=6 width, 1+2=3 height
		assert.Equal(t, 6, result.Box.Width)
		assert.Equal(t, 3, result.Box.Height)
		assert.Equal(t, 2, result.Box.ContentWidth())
		assert.Equal(t, 1, result.Box.ContentHeight())
	})

	t.Run("PaddingAndBorder", func(t *testing.T) {
		node := &TextNode{
			Content: "Hi", // 2 chars wide, 1 line
			Padding: EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
			Border:  EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		}
		result := node.ComputeLayout(Loose(100, 100))

		// Content: 2x1
		// + Padding: 4x3
		// + Border: 6x5
		assert.Equal(t, 6, result.Box.Width)
		assert.Equal(t, 5, result.Box.Height)
	})

	t.Run("Margin", func(t *testing.T) {
		node := &TextNode{
			Content: "Hi",
			Margin:  EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		}
		result := node.ComputeLayout(Loose(100, 100))

		// Content: 2x1
		// Margin doesn't change box dimensions, just MarginBox
		assert.Equal(t, 2, result.Box.Width)
		assert.Equal(t, 1, result.Box.Height)
		assert.Equal(t, 12, result.Box.MarginBoxWidth())
		assert.Equal(t, 11, result.Box.MarginBoxHeight())
	})
}

func TestTextNode_Constraints(t *testing.T) {
	t.Run("MinWidth", func(t *testing.T) {
		node := &TextNode{
			Content:  "Hi", // 2 chars
			MinWidth: 10,
		}
		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 10, result.Box.Width, "stretched to MinWidth")
	})

	t.Run("MaxWidth_TriggersWrap", func(t *testing.T) {
		// Node's MaxWidth (6) intersects with parent's MaxWidth (100).
		// Effective MaxWidth is 6, so text wraps to "Hello" and "World".
		node := &TextNode{
			Content:  "Hello World", // 11 chars
			MaxWidth: 6,
			Wrap:     WrapWord,
		}
		result := node.ComputeLayout(Loose(100, 100))

		// Content is 5 chars wide ("Hello" or "World"), box hugs content
		assert.Equal(t, 5, result.Box.Width, "hugs content after wrapping")
		assert.Equal(t, 2, result.Box.Height, "text wrapped to 2 lines")
	})

	t.Run("MinHeight", func(t *testing.T) {
		node := &TextNode{
			Content:   "Hi",
			MinHeight: 5,
		}
		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 5, result.Box.Height, "stretched to MinHeight")
	})

	t.Run("MaxHeight", func(t *testing.T) {
		// Content exceeds MaxHeight - box is clamped, content overflows.
		// NOTE: This layout engine does NOT provide overflow indication.
		// The renderer is responsible for clipping content that exceeds the box.
		// Lines 4 and 5 exist in the content but won't be visible when rendered.
		node := &TextNode{
			Content:   "Line1\nLine2\nLine3\nLine4\nLine5", // 5 lines
			MaxHeight: 3,
		}
		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 3, result.Box.Height, "clamped to MaxHeight, content overflows")
	})

	t.Run("ParentConstraintsTakePrecedence", func(t *testing.T) {
		node := &TextNode{
			Content:  "Hello World",
			MinWidth: 50, // Node wants at least 50
		}
		// Parent only allows 20
		result := node.ComputeLayout(Loose(20, 100))

		assert.Equal(t, 20, result.Box.Width, "parent constraints win")
	})
}

func TestTextNode_ContentBoxSemantics(t *testing.T) {
	t.Run("PaddingReducesAvailableTextWidth", func(t *testing.T) {
		// With 10 padding on each side, available width for text is reduced
		node := &TextNode{
			Content: "Hello World", // 11 chars
			Wrap:    WrapWord,
			Padding: EdgeInsets{Left: 5, Right: 5}, // 10 horizontal padding
		}
		// Parent allows 16 width
		// Content area = 16 - 10 = 6
		// "Hello World" wraps to "Hello" (5) and "World" (5)
		result := node.ComputeLayout(Loose(16, 100))

		assert.Equal(t, 2, result.Box.Height, "text wrapped due to reduced content width")
		// Width = content (5, widest line) + padding (10) = 15
		assert.Equal(t, 15, result.Box.Width, "content width + padding")
		assert.Equal(t, 5, result.Box.ContentWidth(), "widest line is 5 chars")
	})
}

func TestTextNode_RealWorldScenarios(t *testing.T) {
	t.Run("TypicalLabel", func(t *testing.T) {
		// Simple label with padding
		node := &TextNode{
			Content: "Submit",
			Padding: EdgeInsets{Top: 0, Right: 2, Bottom: 0, Left: 2},
		}
		result := node.ComputeLayout(Loose(100, 1))

		// "Submit" = 6 chars + 4 padding = 10 width
		assert.Equal(t, 10, result.Box.Width)
		assert.Equal(t, 1, result.Box.Height)
	})

	t.Run("Paragraph", func(t *testing.T) {
		paragraph := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
		node := &TextNode{
			Content: paragraph,
			Wrap:    WrapWord,
			Padding: EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		}
		result := node.ComputeLayout(Loose(30, 100))

		// Should wrap to multiple lines
		assert.Greater(t, result.Box.Height, 2, "paragraph wraps to multiple lines")
		assert.LessOrEqual(t, result.Box.Width, 30, "respects max width")
	})

	t.Run("StatusMessage", func(t *testing.T) {
		node := &TextNode{
			Content:   "Error: File not found",
			MinWidth:  40, // Minimum width for status bar
			MinHeight: 1,
		}
		result := node.ComputeLayout(Loose(80, 1))

		assert.Equal(t, 40, result.Box.Width, "uses MinWidth")
		assert.Equal(t, 1, result.Box.Height)
	})
}
