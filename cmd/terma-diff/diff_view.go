package main

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	t "terma"
	"terma/layout"
)

// DiffView is a purpose-built diff renderer with fixed gutter and scroll support.
type DiffView struct {
	ID              string
	DisableFocus    bool
	State           *DiffViewState
	VerticalScroll  *t.ScrollState
	HardWrap        bool
	HideChangeSigns bool
	Palette         ThemePalette
	Width           t.Dimension
	Height          t.Dimension
	Style           t.Style
}

func (d DiffView) Build(ctx t.BuildContext) t.Widget {
	d.Palette = NewThemePalette(ctx.Theme())
	return d
}

func (d DiffView) WidgetID() string {
	return d.ID
}

func (d DiffView) IsFocusable() bool {
	return !d.DisableFocus
}

func (d DiffView) GetContentDimensions() (width, height t.Dimension) {
	dims := d.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = d.Width
	}
	if height.IsUnset() {
		height = d.Height
	}
	return width, height
}

func (d DiffView) GetStyle() t.Style {
	return d.Style
}

func (d DiffView) BuildLayoutNode(ctx t.BuildContext) layout.LayoutNode {
	style := d.Style
	padding := toLayoutInsets(style.Padding)
	border := layout.EdgeInsetsAll(style.Border.Width())
	dims := d.Style.GetDimensions()
	if dims.Width.IsUnset() {
		dims.Width = d.Width
	}
	if dims.Height.IsUnset() {
		dims.Height = d.Height
	}

	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)
	expandWidth := dims.Width.IsFlex() || dims.Width.IsPercent()
	expandHeight := dims.Height.IsFlex() || dims.Height.IsPercent()

	return &layout.BoxNode{
		MinWidth:     minWidth,
		MaxWidth:     maxWidth,
		MinHeight:    minHeight,
		MaxHeight:    maxHeight,
		Padding:      padding,
		Border:       border,
		Margin:       toLayoutInsets(style.Margin),
		ExpandWidth:  expandWidth,
		ExpandHeight: expandHeight,
		MeasureFunc: func(constraints layout.Constraints) (int, int) {
			size := d.Layout(ctx, t.Constraints{
				MinWidth:  constraints.MinWidth,
				MaxWidth:  constraints.MaxWidth,
				MinHeight: constraints.MinHeight,
				MaxHeight: constraints.MaxHeight,
			})
			return size.Width, size.Height
		},
	}
}

func (d DiffView) Layout(ctx t.BuildContext, constraints t.Constraints) t.Size {
	rendered := d.currentRendered()

	dims := d.Style.GetDimensions()
	widthDim := dims.Width
	heightDim := dims.Height
	if widthDim.IsUnset() {
		widthDim = d.Width
	}
	if heightDim.IsUnset() {
		heightDim = d.Height
	}

	gutterWidth := renderedGutterWidth(rendered)
	contentWidth := 1
	contentHeight := 1
	if rendered != nil {
		contentWidth = max(1, gutterWidth+rendered.MaxContentWidth)
		contentHeight = max(1, len(rendered.Lines))
	}

	width := contentWidth
	switch {
	case widthDim.IsCells():
		width = widthDim.CellsValue()
	case widthDim.IsFlex(), widthDim.IsPercent():
		width = constraints.MaxWidth
	}

	width = clampInt(width, constraints.MinWidth, constraints.MaxWidth)

	if d.HardWrap && rendered != nil {
		wrapWidth := max(1, width-gutterWidth)
		contentHeight = wrappedContentHeight(rendered.Lines, wrapWidth)
	}

	height := contentHeight
	switch {
	case heightDim.IsCells():
		height = heightDim.CellsValue()
	case heightDim.IsFlex(), heightDim.IsPercent():
		height = constraints.MaxHeight
	}

	height = clampInt(height, constraints.MinHeight, constraints.MaxHeight)

	return t.Size{Width: width, Height: height}
}

func (d DiffView) Render(ctx *t.RenderContext) {
	if ctx.Width <= 0 || ctx.Height <= 0 || d.State == nil {
		return
	}

	rendered := d.State.Rendered.Get()
	if rendered == nil {
		rendered = buildMetaRenderedFile("Diff", []string{"No diff content to display."})
	}

	clip := ctx.ClipBounds()
	visibleStart := 0
	if clip.Y > ctx.Y {
		visibleStart = clip.Y - ctx.Y
	}
	if visibleStart < 0 {
		visibleStart = 0
	}
	visibleEnd := ctx.Height
	clipEnd := clip.Y + clip.Height - ctx.Y
	if clipEnd < visibleEnd {
		visibleEnd = clipEnd
	}
	if visibleEnd > ctx.Height {
		visibleEnd = ctx.Height
	}
	if visibleEnd <= visibleStart {
		return
	}

	gutterWidth := renderedGutterWidth(rendered)
	d.State.SetViewport(ctx.Width, visibleEnd-visibleStart, gutterWidth)

	scrollY := d.State.ScrollY.Get()
	if d.VerticalScroll != nil {
		scrollY = d.VerticalScroll.Offset.Get()
		d.State.ScrollY.Set(scrollY)
	}
	scrollX := d.State.ScrollX.Get()
	if d.HardWrap {
		scrollX = 0
		if d.State.ScrollX.Peek() != 0 {
			d.State.ScrollX.Set(0)
		}
	}
	if scrollY < 0 {
		scrollY = 0
	}
	if scrollX < 0 {
		scrollX = 0
	}

	wrapWidth := max(1, ctx.Width-gutterWidth)
	for row := visibleStart; row < visibleEnd; row++ {
		contentRow := row
		if d.VerticalScroll == nil {
			contentRow = scrollY + row
		}

		var line RenderedDiffLine
		contentScrollX := scrollX
		continuation := false
		if d.HardWrap {
			var wrapRow int
			var ok bool
			line, wrapRow, ok = wrappedLineAtRow(rendered.Lines, wrapWidth, contentRow)
			if !ok {
				continue
			}
			contentScrollX = wrapRow * wrapWidth
			continuation = wrapRow > 0
		} else {
			if contentRow < 0 || contentRow >= len(rendered.Lines) {
				continue
			}
			line = rendered.Lines[contentRow]
		}

		if lineStyle, ok := d.Palette.LineStyleForKind(line.Kind); ok && lineStyle.BackgroundColor != nil && lineStyle.BackgroundColor.IsSet() {
			bg := lineStyle.BackgroundColor.ColorAt(ctx.Width, 1, 0, 0)
			ctx.FillRect(0, row, ctx.Width, 1, bg)
		}
		gutterLine := line
		if continuation {
			gutterLine.OldLine = 0
			gutterLine.NewLine = 0
			gutterLine.Prefix = " "
		}
		d.renderGutterLine(ctx, rendered, row, gutterLine)
		d.renderContentLine(ctx, row, gutterWidth, line, contentScrollX)
	}
}

func (d DiffView) renderGutterLine(ctx *t.RenderContext, rendered *RenderedFile, row int, line RenderedDiffLine) {
	oldNum := lineNumberText(line.OldLine, rendered.OldNumWidth)
	newNum := lineNumberText(line.NewLine, rendered.NewNumWidth)
	oldNumRole, newNumRole := lineNumberRolesForLine(line.Kind)

	x := 0
	if x < ctx.Width {
		d.drawText(ctx, x, row, oldNum, oldNumRole)
	}
	x += rendered.OldNumWidth
	if x < ctx.Width {
		ctx.DrawText(x, row, " ")
	}
	x++
	if x < ctx.Width {
		d.drawText(ctx, x, row, newNum, newNumRole)
	}
	x += rendered.NewNumWidth
	if x < ctx.Width {
		ctx.DrawText(x, row, " ")
	}
	x++

	prefixRole := TokenRoleDiffPrefixContext
	if role, ok := prefixRoleForLine(line.Kind); ok {
		prefixRole = role
	}
	if x < ctx.Width {
		prefix := displayLinePrefix(line, d.HideChangeSigns)
		d.drawText(ctx, x, row, prefix, prefixRole)
	}
	x++
	if x < ctx.Width {
		ctx.DrawText(x, row, " ")
	}
}

func lineNumberRolesForLine(kind RenderedLineKind) (oldRole TokenRole, newRole TokenRole) {
	oldRole = TokenRoleOldLineNumber
	newRole = TokenRoleNewLineNumber
	switch kind {
	case RenderedLineAdd:
		return TokenRoleLineNumberAdd, TokenRoleLineNumberAdd
	case RenderedLineRemove:
		return TokenRoleLineNumberRemove, TokenRoleLineNumberRemove
	default:
		return oldRole, newRole
	}
}

func displayLinePrefix(line RenderedDiffLine, hideChangeSigns bool) string {
	if hideChangeSigns {
		switch line.Kind {
		case RenderedLineAdd, RenderedLineRemove:
			return " "
		}
	}
	prefix := line.Prefix
	if prefix == "" {
		return " "
	}
	return prefix
}

func (d DiffView) renderContentLine(ctx *t.RenderContext, row int, gutterWidth int, line RenderedDiffLine, scrollX int) {
	if gutterWidth >= ctx.Width {
		return
	}

	visibleWidth := ctx.Width - gutterWidth
	if visibleWidth <= 0 {
		return
	}

	contentCol := 0
	for _, segment := range line.Segments {
		if segment.Text == "" {
			continue
		}
		style := d.styleForRole(segment.Role)
		remaining := segment.Text
		for len(remaining) > 0 {
			grapheme, width := ansi.FirstGraphemeCluster(remaining, ansi.GraphemeWidth)
			if grapheme == "" {
				break
			}
			if width <= 0 {
				width = ansi.StringWidth(grapheme)
			}
			if width <= 0 {
				width = 1
			}

			nextCol := contentCol + width
			if nextCol <= scrollX {
				contentCol = nextCol
				remaining = remaining[len(grapheme):]
				continue
			}
			if contentCol >= scrollX+visibleWidth {
				return
			}

			drawX := gutterWidth + (contentCol - scrollX)
			if drawX >= gutterWidth && drawX < gutterWidth+visibleWidth {
				if width == 1 || (drawX+width) <= gutterWidth+visibleWidth {
					ctx.DrawStyledText(drawX, row, grapheme, style)
				}
			}

			contentCol = nextCol
			remaining = remaining[len(grapheme):]
		}
	}
}

func (d DiffView) drawText(ctx *t.RenderContext, x int, y int, value string, role TokenRole) {
	if x >= ctx.Width || y < 0 || y >= ctx.Height || value == "" {
		return
	}
	ctx.DrawStyledText(x, y, value, d.styleForRole(role))
}

func (d DiffView) styleForRole(role TokenRole) t.Style {
	span, ok := d.Palette.StyleForRole(role)
	if !ok {
		return t.Style{}
	}
	style := t.Style{
		Bold:           span.Bold,
		Faint:          span.Faint,
		Italic:         span.Italic,
		Underline:      span.Underline,
		UnderlineColor: span.UnderlineColor,
		Blink:          span.Blink,
		Reverse:        span.Reverse,
		Conceal:        span.Conceal,
		Strikethrough:  span.Strikethrough,
	}
	if span.Foreground.IsSet() {
		style.ForegroundColor = span.Foreground
	}
	if span.Background.IsSet() {
		style.BackgroundColor = span.Background
	}
	return style
}

func (d DiffView) currentRendered() *RenderedFile {
	if d.State == nil {
		return nil
	}
	return d.State.Rendered.Peek()
}

func renderedGutterWidth(rendered *RenderedFile) int {
	if rendered == nil {
		return 6
	}
	oldWidth := rendered.OldNumWidth
	if oldWidth <= 0 {
		oldWidth = 1
	}
	newWidth := rendered.NewNumWidth
	if newWidth <= 0 {
		newWidth = 1
	}
	return oldWidth + 1 + newWidth + 1 + 1 + 1
}

func wrappedContentHeight(lines []RenderedDiffLine, wrapWidth int) int {
	if len(lines) == 0 {
		return 1
	}
	total := 0
	for _, line := range lines {
		total += wrappedLineRowCount(line, wrapWidth)
	}
	if total <= 0 {
		return 1
	}
	return total
}

func wrappedLineAtRow(lines []RenderedDiffLine, wrapWidth int, rowIdx int) (RenderedDiffLine, int, bool) {
	if rowIdx < 0 {
		return RenderedDiffLine{}, 0, false
	}
	remaining := rowIdx
	for _, line := range lines {
		rows := wrappedLineRowCount(line, wrapWidth)
		if remaining < rows {
			return line, remaining, true
		}
		remaining -= rows
	}
	return RenderedDiffLine{}, 0, false
}

func wrappedLineRowCount(line RenderedDiffLine, wrapWidth int) int {
	if wrapWidth <= 0 {
		return 1
	}
	if line.ContentWidth <= 0 {
		return 1
	}
	rows := (line.ContentWidth + wrapWidth - 1) / wrapWidth
	if rows <= 0 {
		return 1
	}
	return rows
}

func toLayoutInsets(in t.EdgeInsets) layout.EdgeInsets {
	return layout.EdgeInsets{
		Top:    in.Top,
		Right:  in.Right,
		Bottom: in.Bottom,
		Left:   in.Left,
	}
}

func dimensionSetToMinMax(ds t.DimensionSet, padding layout.EdgeInsets, border layout.EdgeInsets) (minW int, maxW int, minH int, maxH int) {
	explicitMinW := dimensionToCells(ds.MinWidth)
	explicitMaxW := dimensionToCells(ds.MaxWidth)
	explicitMinH := dimensionToCells(ds.MinHeight)
	explicitMaxH := dimensionToCells(ds.MaxHeight)

	if ds.Width.IsCells() {
		width := clampFixedDimension(ds.Width.CellsValue(), explicitMinW, explicitMaxW)
		minW, maxW = width, width
	} else {
		minW, maxW = explicitMinW, explicitMaxW
	}
	if ds.Height.IsCells() {
		height := clampFixedDimension(ds.Height.CellsValue(), explicitMinH, explicitMaxH)
		minH, maxH = height, height
	} else {
		minH, maxH = explicitMinH, explicitMaxH
	}

	hInset := padding.Horizontal() + border.Horizontal()
	vInset := padding.Vertical() + border.Vertical()

	if minW > 0 {
		minW += hInset
	}
	if maxW > 0 {
		maxW += hInset
	}
	if minH > 0 {
		minH += vInset
	}
	if maxH > 0 {
		maxH += vInset
	}
	return minW, maxW, minH, maxH
}

func dimensionToCells(dim t.Dimension) int {
	if dim.IsCells() {
		return dim.CellsValue()
	}
	return 0
}

func clampFixedDimension(value int, minValue int, maxValue int) int {
	if minValue > 0 && maxValue > 0 && maxValue < minValue {
		return minValue
	}
	if minValue > 0 && value < minValue {
		value = minValue
	}
	if maxValue > 0 && value > maxValue {
		value = maxValue
	}
	return value
}

func lineText(line RenderedDiffLine) string {
	if len(line.Segments) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, segment := range line.Segments {
		builder.WriteString(segment.Text)
	}
	return builder.String()
}

func clampInt(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
