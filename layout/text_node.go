package layout

// TextNode is a leaf node for text content that may wrap.
// It's a thin wrapper around BoxNode that provides a declarative API for text.
//
// Rather than duplicating BoxNode's constraint/inset handling logic,
// TextNode creates an internal BoxNode with a MeasureFunc that calls MeasureText.
// This ensures bug fixes to BoxNode automatically apply to TextNode.
//
// Example:
//
//	node := &TextNode{
//	    Content: "Hello, World!",
//	    Wrap:    WrapWord,
//	    Padding: EdgeInsets{Top: 1, Right: 2, Bottom: 1, Left: 2},
//	}
//	result := node.ComputeLayout(Loose(80, 24))
type TextNode struct {
	// Content is the text to measure and display.
	Content string

	// Wrap controls how text wraps when it exceeds available width.
	// Default is WrapNone (no wrapping).
	Wrap WrapMode

	// Insets (passed through to BoxNode)
	Padding EdgeInsets
	Border  EdgeInsets
	Margin  EdgeInsets

	// Optional constraints (passed through to BoxNode).
	// NOTE: 0 means "no constraint" (unconstrained), not "zero size".
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
}

// ComputeLayout computes the layout for this text node.
// It delegates to a BoxNode with a MeasureFunc that calls MeasureText.
func (t *TextNode) ComputeLayout(constraints Constraints) ComputedLayout {
	box := &BoxNode{
		Padding:   t.Padding,
		Border:    t.Border,
		Margin:    t.Margin,
		MinWidth:  t.MinWidth,
		MaxWidth:  t.MaxWidth,
		MinHeight: t.MinHeight,
		MaxHeight: t.MaxHeight,
		MeasureFunc: func(c Constraints) (int, int) {
			return MeasureText(t.Content, t.Wrap, c.MaxWidth)
		},
	}
	return box.ComputeLayout(constraints)
}
