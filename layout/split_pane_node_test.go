package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func splitBox(w, h int) *BoxNode {
	return &BoxNode{Width: w, Height: h}
}

func TestSplitPaneNode_Horizontal(t *testing.T) {
	pane := &SplitPaneNode{
		First:       splitBox(0, 0),
		Second:      splitBox(0, 0),
		Axis:        Horizontal,
		Position:    0.5,
		DividerSize: 1,
		MinPaneSize: 1,
	}

	result := pane.ComputeLayout(Tight(100, 20))

	assert.Equal(t, 100, result.Box.Width)
	assert.Equal(t, 20, result.Box.Height)
	assert.Len(t, result.Children, 2)

	assert.Equal(t, 0, result.Children[0].X)
	assert.Equal(t, 0, result.Children[0].Y)
	assert.Equal(t, 49, result.Children[0].Layout.Box.Width)
	assert.Equal(t, 20, result.Children[0].Layout.Box.Height)

	assert.Equal(t, 50, result.Children[1].X)
	assert.Equal(t, 0, result.Children[1].Y)
	assert.Equal(t, 50, result.Children[1].Layout.Box.Width)
	assert.Equal(t, 20, result.Children[1].Layout.Box.Height)
}

func TestSplitPaneNode_Vertical(t *testing.T) {
	pane := &SplitPaneNode{
		First:       splitBox(0, 0),
		Second:      splitBox(0, 0),
		Axis:        Vertical,
		Position:    0.25,
		DividerSize: 1,
		MinPaneSize: 1,
	}

	result := pane.ComputeLayout(Tight(40, 20))

	assert.Equal(t, 40, result.Box.Width)
	assert.Equal(t, 20, result.Box.Height)
	assert.Len(t, result.Children, 2)

	assert.Equal(t, 0, result.Children[0].X)
	assert.Equal(t, 0, result.Children[0].Y)
	assert.Equal(t, 40, result.Children[0].Layout.Box.Width)
	assert.Equal(t, 4, result.Children[0].Layout.Box.Height)

	assert.Equal(t, 0, result.Children[1].X)
	assert.Equal(t, 5, result.Children[1].Y)
	assert.Equal(t, 40, result.Children[1].Layout.Box.Width)
	assert.Equal(t, 15, result.Children[1].Layout.Box.Height)
}

func TestSplitPaneNode_MinPaneClamp(t *testing.T) {
	pane := &SplitPaneNode{
		First:       splitBox(0, 0),
		Second:      splitBox(0, 0),
		Axis:        Horizontal,
		Position:    0.05,
		DividerSize: 1,
		MinPaneSize: 10,
	}

	result := pane.ComputeLayout(Tight(50, 10))

	assert.Equal(t, 10, result.Children[0].Layout.Box.Width)
	assert.Equal(t, 39, result.Children[1].Layout.Box.Width)
	assert.Equal(t, 11, result.Children[1].X)
}

func TestSplitPaneNode_DividerLargerThanAvailable(t *testing.T) {
	pane := &SplitPaneNode{
		First:       splitBox(0, 0),
		Second:      splitBox(0, 0),
		Axis:        Horizontal,
		Position:    0.5,
		DividerSize: 10,
		MinPaneSize: 1,
	}

	result := pane.ComputeLayout(Tight(5, 4))

	assert.Equal(t, 5, result.Box.Width)
	assert.Equal(t, 4, result.Box.Height)
	assert.Equal(t, 0, result.Children[0].Layout.Box.Width)
	assert.Equal(t, 0, result.Children[1].Layout.Box.Width)
}
