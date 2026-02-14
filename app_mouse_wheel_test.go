package terma

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDispatchMouseWheel_HorizontalMovesScrollableOffsetX(t *testing.T) {
	state := NewScrollState()
	widget := Scrollable{
		ID:    "scrollable",
		State: state,
		Width: Cells(10),
		Child: Text{
			Content: strings.Repeat("x", 30),
			Width:   Cells(30),
		},
	}

	buf := uv.NewBuffer(20, 4)
	focusManager := NewFocusManager()
	renderer := NewRenderer(buf, 20, 4, focusManager, NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	renderer.Render(widget)
	state.updateHorizontalLayout(10, 30)

	handled := dispatchMouseWheel(renderer, 0, 0, uv.MouseWheelRight)
	require.True(t, handled)
	assert.Equal(t, 1, state.GetOffsetX())

	handled = dispatchMouseWheel(renderer, 0, 0, uv.MouseWheelLeft)
	require.True(t, handled)
	assert.Equal(t, 0, state.GetOffsetX())
}

func TestDispatchMouseWheel_HorizontalBubblesToOuterScrollable(t *testing.T) {
	outerState := NewScrollState()
	innerState := NewScrollState()
	widget := Scrollable{
		ID:    "outer",
		State: outerState,
		Width: Cells(10),
		Child: Scrollable{
			ID:    "inner",
			State: innerState,
			Width: Cells(20),
			Child: Text{
				Content: "short",
				Width:   Cells(5),
			},
		},
	}

	buf := uv.NewBuffer(20, 4)
	focusManager := NewFocusManager()
	renderer := NewRenderer(buf, 20, 4, focusManager, NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	renderer.Render(widget)
	innerState.updateHorizontalLayout(10, 10)
	outerState.updateHorizontalLayout(10, 30)

	handled := dispatchMouseWheel(renderer, 0, 0, uv.MouseWheelRight)
	require.True(t, handled)
	assert.Equal(t, 1, outerState.GetOffsetX())
	assert.Equal(t, 0, innerState.GetOffsetX())
}

func TestDispatchMouseWheel_HorizontalUsesCallbacksWhenNoLayoutOverflow(t *testing.T) {
	state := NewScrollState()
	calls := 0
	state.OnScrollRight = func(cols int) bool {
		calls += cols
		return true
	}

	widget := Scrollable{
		ID:    "scrollable",
		State: state,
		Width: Cells(10),
		Child: Text{
			Content: "fits",
		},
	}

	buf := uv.NewBuffer(20, 4)
	focusManager := NewFocusManager()
	renderer := NewRenderer(buf, 20, 4, focusManager, NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	renderer.Render(widget)

	handled := dispatchMouseWheel(renderer, 0, 0, uv.MouseWheelRight)
	require.True(t, handled)
	assert.Equal(t, 1, calls)
	assert.Equal(t, 0, state.GetOffsetX())
}
