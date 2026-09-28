package terma

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarginAppliedOnceToLinearChildren(t *testing.T) {
	widget := Column{Children: []Widget{
		Text{Content: "above"},
		Text{Content: "x", Style: Style{Margin: EdgeInsets{Left: 3}}},
		Text{Content: "y", Style: Style{Margin: EdgeInsets{Top: 1}}},
		Row{Children: []Widget{
			Text{Content: "a"},
			Text{Content: "b", Style: Style{Margin: EdgeInsets{Left: 2}}},
		}},
	}}

	require.Equal(t, "above\n   x \n     \ny    \na  b ", RenderToPlainString(widget, 12, 5))
}
