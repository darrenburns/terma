package terma

import "github.com/darrenburns/terma/layout"

// LayoutMetrics provides convenient access to computed layout information.
// Child bounds are returned in the parent's content coordinate space.
type LayoutMetrics struct {
	layout layout.ComputedLayout
}

// Box returns the computed BoxModel for this widget.
func (m LayoutMetrics) Box() layout.BoxModel {
	return m.layout.Box
}

// ChildCount returns the number of positioned children.
func (m LayoutMetrics) ChildCount() int {
	return len(m.layout.Children)
}

// ChildLayout returns the computed layout for the child at index i.
func (m LayoutMetrics) ChildLayout(i int) (layout.ComputedLayout, bool) {
	if i < 0 || i >= len(m.layout.Children) {
		return layout.ComputedLayout{}, false
	}
	return m.layout.Children[i].Layout, true
}

// ChildBounds returns the child's border-box bounds in content coordinates.
func (m LayoutMetrics) ChildBounds(i int) (Rect, bool) {
	if i < 0 || i >= len(m.layout.Children) {
		return Rect{}, false
	}
	child := m.layout.Children[i]
	return Rect{
		X:      child.X,
		Y:      child.Y,
		Width:  child.Layout.Box.BorderBoxWidth(),
		Height: child.Layout.Box.BorderBoxHeight(),
	}, true
}

// ChildMarginBounds returns the child's margin-box bounds in content coordinates.
func (m LayoutMetrics) ChildMarginBounds(i int) (Rect, bool) {
	if i < 0 || i >= len(m.layout.Children) {
		return Rect{}, false
	}
	child := m.layout.Children[i]
	return Rect{
		X:      child.X - child.Layout.Box.Margin.Left,
		Y:      child.Y - child.Layout.Box.Margin.Top,
		Width:  child.Layout.Box.MarginBoxWidth(),
		Height: child.Layout.Box.MarginBoxHeight(),
	}, true
}

// ChildY returns the child's border-box Y position in content coordinates.
func (m LayoutMetrics) ChildY(i int) (int, bool) {
	bounds, ok := m.ChildBounds(i)
	if !ok {
		return 0, false
	}
	return bounds.Y, true
}

// ChildHeight returns the child's border-box height in cells.
func (m LayoutMetrics) ChildHeight(i int) (int, bool) {
	bounds, ok := m.ChildBounds(i)
	if !ok {
		return 0, false
	}
	return bounds.Height, true
}
