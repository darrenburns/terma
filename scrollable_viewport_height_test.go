package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScrollableViewportHeightIntegration_NoUnboundedVirtualHeight(t *testing.T) {
	t.Run("TextAreaFlexHeight_IsViewportBounded", func(t *testing.T) {
		scrollState := NewScrollState()
		textAreaState := NewTextAreaState("one\ntwo\nthree")

		widget := Scrollable{
			State:  scrollState,
			Height: Cells(12),
			Child: TextArea{
				State: textAreaState,
				Style: Style{
					Width:  Flex(1),
					Height: Flex(1),
				},
			},
		}

		RenderToBuffer(widget, 40, 24)

		assert.Equal(t, 12, scrollState.viewportHeight)
		assert.Equal(t, 12, scrollState.contentHeight)
		assert.Less(t, scrollState.contentHeight, 100_000)
	})

	t.Run("SparklinePercentHeight_IsViewportBounded", func(t *testing.T) {
		scrollState := NewScrollState()

		widget := Scrollable{
			State:  scrollState,
			Height: Cells(12),
			Child: Sparkline{
				Values: []float64{1, 3, 2, 4, 3, 5, 2, 6},
				Style: Style{
					Width:  Flex(1),
					Height: Percent(50),
				},
			},
		}

		RenderToBuffer(widget, 40, 24)

		assert.Equal(t, 12, scrollState.viewportHeight)
		assert.Greater(t, scrollState.contentHeight, 0)
		assert.LessOrEqual(t, scrollState.contentHeight, scrollState.viewportHeight)
		assert.Less(t, scrollState.contentHeight, 100_000)
	})
}
