package terma

import "github.com/darrenburns/terma/layout"

type tableViewportMetrics struct {
	reclip                                     func(int, int) []layout.PositionedChild
	width, height, contentWidth, contentHeight int
	frozenWidth, frozenHeight, headerHeight    int
	columns                                    []tableColumnLayout
	rows                                       []tableRowLayout
}

// clipViewportCells keeps every cell in the intersection of the viewport and its
// frozen/scrolling pane. The passThrough cell wrapper supplies a real clip for
// both rendering and hit testing; its child keeps its full untrimmed layout.
func (t *tableNode) clipViewportCells(positioned []layout.PositionedChild, widths, heights []int, width, height int) []layout.PositionedChild {
	frozenCols := clampInt(t.FrozenColumns, 0, len(widths))
	frozenRows := clampInt(t.FrozenRows, 0, len(heights))
	frozenWidth, frozenHeight := 0, 0
	for i := 0; i < frozenCols; i++ {
		frozenWidth += widths[i]
		if i < len(widths)-1 {
			frozenWidth += t.ColumnSpacing
		}
	}
	for i := 0; i < frozenRows; i++ {
		frozenHeight += heights[i]
		if i < len(heights)-1 {
			frozenHeight += t.RowSpacing
		}
	}
	contentWidth := sumInts(widths) + max(0, len(widths)-1)*t.ColumnSpacing
	contentHeight := sumInts(heights) + max(0, len(heights)-1)*t.RowSpacing
	offsetX := clampInt(t.OffsetX, 0, max(0, contentWidth-width))
	offsetY := clampInt(t.OffsetY, 0, max(0, contentHeight-height))
	headerRows := 0
	if t.Header {
		headerRows = 1
	}
	metrics := tableViewportMetrics{width: width, height: height, contentWidth: contentWidth, contentHeight: contentHeight, frozenWidth: min(width, frozenWidth), frozenHeight: min(height, frozenHeight), columns: make([]tableColumnLayout, len(widths)), rows: make([]tableRowLayout, max(0, len(heights)-headerRows))}
	if t.Header && len(heights) > 0 {
		metrics.headerHeight = heights[0]
	}
	for i, cell := range positioned {
		row, col := i/t.Columns, i%t.Columns
		if row == 0 {
			metrics.columns[col] = tableColumnLayout{x: cell.X, width: widths[col]}
		}
		if col == 0 && row >= headerRows {
			metrics.rows[row-headerRows] = tableRowLayout{y: cell.Y, height: heights[row]}
		}
	}
	if t.ViewportMetrics != nil {
		original := append([]layout.PositionedChild(nil), positioned...)
		metrics.reclip = func(x, y int) []layout.PositionedChild {
			node := *t
			node.OffsetX, node.OffsetY = x, y
			node.ViewportMetrics = nil
			return node.clipViewportCells(append([]layout.PositionedChild(nil), original...), widths, heights, width, height)
		}
	}
	for i, cell := range positioned {
		row, col := i/t.Columns, i%t.Columns
		x, y := cell.X, cell.Y
		left, top := 0, 0
		if col >= frozenCols {
			x -= offsetX
			left = metrics.frozenWidth
		}
		if row >= frozenRows {
			y -= offsetY
			top = metrics.frozenHeight
		}
		clipLeft, clipTop := max(left, x), max(top, y)
		clipRight, clipBottom := min(width, x+widths[col]), min(height, y+heights[row])
		cell.X = clipLeft
		cell.Y = clipTop
		cell.Layout.Box.Width = max(0, clipRight-clipLeft)
		cell.Layout.Box.Height = max(0, clipBottom-clipTop)
		if len(cell.Layout.Children) > 0 {
			// Child layouts may come from retained layout caches. Copy their
			// positions before shifting to avoid accumulating offsets.
			cell.Layout.Children = append([]layout.PositionedChild(nil), cell.Layout.Children...)
			cell.Layout.Children[0].X += x - clipLeft
			cell.Layout.Children[0].Y += y - clipTop
		}
		positioned[i] = cell
	}
	if t.ViewportMetrics != nil {
		*t.ViewportMetrics = metrics
	}
	return positioned
}
