package terma

import "github.com/darrenburns/terma/layout"

type tableNode struct {
	Columns int
	Rows    int

	ColumnWidths  []Dimension
	ColumnSpacing int
	RowSpacing    int
	Children      []layout.LayoutNode

	Padding layout.EdgeInsets
	Border  layout.EdgeInsets
	Margin  layout.EdgeInsets

	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int

	ExpandWidth  bool
	ExpandHeight bool

	PreserveWidth  bool
	PreserveHeight bool
}

func (t *tableNode) ComputeLayout(constraints layout.Constraints) layout.ComputedLayout {
	effective := t.effectiveConstraints(constraints)

	if t.Columns <= 0 || t.Rows <= 0 || len(t.Children) == 0 {
		return t.emptyLayout(effective)
	}

	contentConstraints := t.toContentConstraints(effective)
	cols, rows := t.Columns, t.Rows

	if rows*cols > len(t.Children) {
		rows = len(t.Children) / cols
	}
	if rows == 0 {
		return t.emptyLayout(effective)
	}

	columnWidths := t.computeColumnWidths(rows, cols, contentConstraints)
	cellLayouts, rowHeights := t.layoutCells(rows, cols, columnWidths, contentConstraints)

	contentWidth := sumInts(columnWidths)
	if cols > 1 {
		contentWidth += t.ColumnSpacing * (cols - 1)
	}

	contentHeight := sumInts(rowHeights)
	if rows > 1 {
		contentHeight += t.RowSpacing * (rows - 1)
	}

	containerWidth := t.resolveContainerSize(contentConstraints.MinWidth, contentConstraints.MaxWidth, contentWidth, t.ExpandWidth)
	containerHeight := t.resolveContainerSize(contentConstraints.MinHeight, contentConstraints.MaxHeight, contentHeight, t.ExpandHeight)

	positioned := t.positionCells(rows, cols, columnWidths, rowHeights, cellLayouts)

	return t.buildResult(effective, containerWidth, containerHeight, positioned)
}

func (t *tableNode) effectiveConstraints(parent layout.Constraints) layout.Constraints {
	result := parent.WithNodeConstraints(t.MinWidth, t.MaxWidth, t.MinHeight, t.MaxHeight)
	if t.ExpandWidth {
		result.MinWidth = result.MaxWidth
	}
	if t.ExpandHeight {
		result.MinHeight = result.MaxHeight
	}
	return result
}

func (t *tableNode) emptyLayout(constraints layout.Constraints) layout.ComputedLayout {
	hInset := t.Padding.Horizontal() + t.Border.Horizontal()
	vInset := t.Padding.Vertical() + t.Border.Vertical()

	width, height := constraints.Constrain(hInset, vInset)

	return layout.ComputedLayout{
		Box: layout.BoxModel{
			Width:   width,
			Height:  height,
			Padding: t.Padding,
			Border:  t.Border,
			Margin:  t.Margin,
		},
		Children: nil,
	}
}

func (t *tableNode) toContentConstraints(constraints layout.Constraints) layout.Constraints {
	hInset := t.Padding.Horizontal() + t.Border.Horizontal()
	vInset := t.Padding.Vertical() + t.Border.Vertical()

	return layout.Constraints{
		MinWidth:  max(0, constraints.MinWidth-hInset),
		MaxWidth:  max(0, constraints.MaxWidth-hInset),
		MinHeight: max(0, constraints.MinHeight-vInset),
		MaxHeight: max(0, constraints.MaxHeight-vInset),
	}
}

func (t *tableNode) computeColumnWidths(rows, cols int, contentConstraints layout.Constraints) []int {
	widths := make([]int, cols)
	intrinsic := t.measureIntrinsicWidths(rows, cols, contentConstraints)

	widthMax := contentConstraints.MaxWidth
	widthBounded := widthMax < maxTableInt()

	totalSpacing := 0
	if cols > 1 {
		totalSpacing = t.ColumnSpacing * (cols - 1)
	}

	available := widthMax - totalSpacing
	if available < 0 {
		available = 0
	}

	hasFlex := false
	totalFlex := 0.0
	autoFlags := make([]bool, cols)
	flexFlags := make([]bool, cols)
	flexValues := make([]float64, cols)
	fixedTotal := 0

	for i := 0; i < cols; i++ {
		var dim Dimension
		if i < len(t.ColumnWidths) {
			dim = t.ColumnWidths[i]
		}

		switch {
		case dim.IsCells():
			widths[i] = max(0, dim.CellsValue())
			fixedTotal += widths[i]
		case dim.IsPercent():
			if widthBounded {
				widths[i] = int(float64(available) * dim.PercentValue() / 100.0)
			} else {
				widths[i] = intrinsic[i]
				autoFlags[i] = true
			}
			fixedTotal += widths[i]
		case dim.IsFlex():
			flexFlags[i] = true
			hasFlex = true
			flexValues[i] = normalizeFlex(dim.FlexValue())
			totalFlex += flexValues[i]
		default:
			widths[i] = intrinsic[i]
			fixedTotal += widths[i]
			autoFlags[i] = true
		}
	}

	if !widthBounded {
		for i := 0; i < cols; i++ {
			if flexFlags[i] {
				widths[i] = intrinsic[i]
			}
		}
		return widths
	}

	if hasFlex {
		remaining := available - fixedTotal
		if remaining < 0 {
			remaining = 0
		}
		distributeFlex(widths, flexFlags, flexValues, totalFlex, remaining)
		return widths
	}

	extra := available - fixedTotal
	if extra <= 0 {
		return widths
	}

	distributeExtra(widths, autoFlags, extra)
	return widths
}

func (t *tableNode) measureIntrinsicWidths(rows, cols int, contentConstraints layout.Constraints) []int {
	intrinsic := make([]int, cols)
	maxWidth := contentConstraints.MaxWidth
	if maxWidth <= 0 {
		maxWidth = 0
	}
	if maxWidth >= maxTableInt() {
		maxWidth = maxTableInt()
	}
	maxHeight := contentConstraints.MaxHeight
	if maxHeight <= 0 {
		maxHeight = 0
	}
	if maxHeight >= maxTableInt() {
		maxHeight = maxTableInt()
	}

	for col := 0; col < cols; col++ {
		maxWidthForCol := 0
		for row := 0; row < rows; row++ {
			idx := row*cols + col
			if idx < 0 || idx >= len(t.Children) {
				continue
			}
			child := stripExpandHeight(t.Children[idx])
			layout := child.ComputeLayout(layout.Constraints{
				MinWidth:  0,
				MaxWidth:  maxWidth,
				MinHeight: 0,
				MaxHeight: maxHeight,
			})
			width := layout.Box.BorderBoxWidth()
			if width > maxWidthForCol {
				maxWidthForCol = width
			}
		}
		intrinsic[col] = maxWidthForCol
	}

	return intrinsic
}

func (t *tableNode) layoutCells(rows, cols int, columnWidths []int, contentConstraints layout.Constraints) ([]layout.ComputedLayout, []int) {
	rowHeights := make([]int, rows)

	maxHeight := contentConstraints.MaxHeight
	if maxHeight < 0 {
		maxHeight = 0
	}
	if maxHeight >= maxTableInt() {
		maxHeight = maxTableInt()
	}

	for row := 0; row < rows; row++ {
		rowHeight := 0
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if idx < 0 || idx >= len(t.Children) {
				continue
			}
			width := columnWidths[col]
			if width < 0 {
				width = 0
			}
			child := stripExpandHeight(t.Children[idx])
			cellLayout := child.ComputeLayout(layout.Constraints{
				MinWidth:  width,
				MaxWidth:  width,
				MinHeight: 0,
				MaxHeight: maxHeight,
			})
			height := cellLayout.Box.BorderBoxHeight()
			if height > rowHeight {
				rowHeight = height
			}
		}
		rowHeights[row] = rowHeight
	}

	cellLayouts := make([]layout.ComputedLayout, rows*cols)
	for row := 0; row < rows; row++ {
		rowHeight := rowHeights[row]
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if idx < 0 || idx >= len(t.Children) {
				continue
			}
			width := columnWidths[col]
			if width < 0 {
				width = 0
			}
			child := t.Children[idx]
			cellLayout := child.ComputeLayout(layout.Constraints{
				MinWidth:  width,
				MaxWidth:  width,
				MinHeight: rowHeight,
				MaxHeight: rowHeight,
			})
			cellLayouts[idx] = cellLayout
		}
	}

	return cellLayouts, rowHeights
}

func (t *tableNode) positionCells(rows, cols int, columnWidths []int, rowHeights []int, cellLayouts []layout.ComputedLayout) []layout.PositionedChild {
	positioned := make([]layout.PositionedChild, rows*cols)

	y := 0
	for row := 0; row < rows; row++ {
		x := 0
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if idx < 0 || idx >= len(cellLayouts) {
				continue
			}
			positioned[idx] = layout.PositionedChild{
				X:      x,
				Y:      y,
				Layout: cellLayouts[idx],
			}
			x += columnWidths[col]
			if col < cols-1 {
				x += t.ColumnSpacing
			}
		}
		y += rowHeights[row]
		if row < rows-1 {
			y += t.RowSpacing
		}
	}

	return positioned
}

func (t *tableNode) resolveContainerSize(minVal, maxVal, content int, expand bool) int {
	if expand {
		return maxVal
	}
	return max(minVal, min(maxVal, content))
}

func (t *tableNode) buildResult(constraints layout.Constraints, contentWidth, contentHeight int, children []layout.PositionedChild) layout.ComputedLayout {
	borderWidth := contentWidth + t.Padding.Horizontal() + t.Border.Horizontal()
	borderHeight := contentHeight + t.Padding.Vertical() + t.Border.Vertical()

	borderWidth, borderHeight = constraints.Constrain(borderWidth, borderHeight)

	return layout.ComputedLayout{
		Box: layout.BoxModel{
			Width:   borderWidth,
			Height:  borderHeight,
			Padding: t.Padding,
			Border:  t.Border,
			Margin:  t.Margin,
		},
		Children: children,
	}
}

func (t *tableNode) PreservesWidth() bool {
	return t.PreserveWidth
}

func (t *tableNode) PreservesHeight() bool {
	return t.PreserveHeight
}

func distributeFlex(widths []int, flexFlags []bool, flexValues []float64, totalFlex float64, remaining int) {
	if totalFlex <= 0 || remaining <= 0 {
		return
	}

	allocatedSoFar := 0.0
	actualAllocatedSoFar := 0

	for i := range widths {
		if !flexFlags[i] {
			continue
		}
		allocatedSoFar += flexValues[i]
		share := float64(remaining) * allocatedSoFar / totalFlex
		allocation := int(share) - actualAllocatedSoFar
		actualAllocatedSoFar += allocation
		widths[i] = allocation
	}
}

func distributeExtra(widths []int, autoFlags []bool, extra int) {
	autoTotal := 0
	autoCount := 0
	for i, width := range widths {
		if autoFlags[i] {
			autoTotal += width
			autoCount++
		}
	}

	if autoCount == 0 {
		return
	}

	if autoTotal <= 0 {
		each := extra / autoCount
		remainder := extra % autoCount
		for i := range widths {
			if !autoFlags[i] {
				continue
			}
			widths[i] += each
			if remainder > 0 {
				widths[i]++
				remainder--
			}
		}
		return
	}

	allocatedSoFar := 0.0
	actualAllocatedSoFar := 0
	for i, width := range widths {
		if !autoFlags[i] {
			continue
		}
		allocatedSoFar += float64(width)
		targetTotal := float64(extra) * allocatedSoFar / float64(autoTotal)
		allocation := int(targetTotal) - actualAllocatedSoFar
		actualAllocatedSoFar += allocation
		widths[i] += allocation
	}
}

func normalizeFlex(value float64) float64 {
	if value <= 0 {
		return 1
	}
	return value
}

func stripExpandHeight(node layout.LayoutNode) layout.LayoutNode {
	switch n := node.(type) {
	case *layout.RowNode:
		copy := *n
		copy.ExpandHeight = false
		return &copy
	case *layout.ColumnNode:
		copy := *n
		copy.ExpandHeight = false
		return &copy
	case *layout.StackNode:
		copy := *n
		copy.ExpandHeight = false
		return &copy
	default:
		return node
	}
}

func maxTableInt() int {
	return int(^uint(0) >> 1)
}

func sumInts(vals []int) int {
	total := 0
	for _, v := range vals {
		total += v
	}
	return total
}
