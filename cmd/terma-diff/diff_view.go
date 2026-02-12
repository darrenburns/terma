package main

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	t "terma"
	"terma/layout"
)

// DiffView is a purpose-built diff renderer with fixed gutter and scroll support.
type DiffView struct {
	ID           string
	DisableFocus bool
	State        *DiffViewState
	Palette      ThemePalette
	Width        t.Dimension
	Height       t.Dimension
	Style        t.Style
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

func (d DiffView) Keybinds() []t.Keybind {
	if d.State == nil {
		return nil
	}
	return []t.Keybind{
		{Key: "up", Hidden: true, Action: func() { d.scrollY(-1) }},
		{Key: "k", Hidden: true, Action: func() { d.scrollY(-1) }},
		{Key: "down", Hidden: true, Action: func() { d.scrollY(1) }},
		{Key: "j", Hidden: true, Action: func() { d.scrollY(1) }},
		{Key: "pgup", Hidden: true, Action: func() { d.pageUp() }},
		{Key: "pgdown", Hidden: true, Action: func() { d.pageDown() }},
		{Key: "ctrl+u", Hidden: true, Action: func() { d.halfPageUp() }},
		{Key: "ctrl+d", Hidden: true, Action: func() { d.halfPageDown() }},
		{Key: "g", Hidden: true, Action: func() { d.toTop() }},
		{Key: "G", Hidden: true, Action: func() { d.toBottom() }},
		{Key: "left", Hidden: true, Action: func() { d.scrollX(-1) }},
		{Key: "h", Hidden: true, Action: func() { d.scrollX(-1) }},
		{Key: "right", Hidden: true, Action: func() { d.scrollX(1) }},
		{Key: "l", Hidden: true, Action: func() { d.scrollX(1) }},
	}
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

	contentWidth := 1
	contentHeight := 1
	if rendered != nil {
		contentWidth = max(1, renderedGutterWidth(rendered)+rendered.MaxContentWidth)
		contentHeight = max(1, len(rendered.Lines))
	}

	width := contentWidth
	switch {
	case widthDim.IsCells():
		width = widthDim.CellsValue()
	case widthDim.IsFlex(), widthDim.IsPercent():
		width = constraints.MaxWidth
	}

	height := contentHeight
	switch {
	case heightDim.IsCells():
		height = heightDim.CellsValue()
	case heightDim.IsFlex(), heightDim.IsPercent():
		height = constraints.MaxHeight
	}

	width = clampInt(width, constraints.MinWidth, constraints.MaxWidth)
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

	gutterWidth := renderedGutterWidth(rendered)
	d.State.SetViewport(ctx.Width, ctx.Height, gutterWidth)

	scrollY := d.State.ScrollY.Get()
	scrollX := d.State.ScrollX.Get()
	if scrollY < 0 {
		scrollY = 0
	}
	if scrollX < 0 {
		scrollX = 0
	}

	for row := 0; row < ctx.Height; row++ {
		lineIdx := scrollY + row
		if lineIdx < 0 || lineIdx >= len(rendered.Lines) {
			continue
		}
		line := rendered.Lines[lineIdx]
		if lineStyle, ok := d.Palette.LineStyleForKind(line.Kind); ok && lineStyle.BackgroundColor != nil && lineStyle.BackgroundColor.IsSet() {
			bg := lineStyle.BackgroundColor.ColorAt(ctx.Width, 1, 0, 0)
			ctx.FillRect(0, row, ctx.Width, 1, bg)
		}
		d.renderGutterLine(ctx, rendered, row, line)
		d.renderContentLine(ctx, row, gutterWidth, line, scrollX)
	}
}

func (d DiffView) renderGutterLine(ctx *t.RenderContext, rendered *RenderedFile, row int, line RenderedDiffLine) {
	oldNum := lineNumberText(line.OldLine, rendered.OldNumWidth)
	newNum := lineNumberText(line.NewLine, rendered.NewNumWidth)

	x := 0
	if x < ctx.Width {
		d.drawText(ctx, x, row, oldNum, TokenRoleOldLineNumber)
	}
	x += rendered.OldNumWidth
	if x < ctx.Width {
		ctx.DrawText(x, row, " ")
	}
	x++
	if x < ctx.Width {
		d.drawText(ctx, x, row, newNum, TokenRoleNewLineNumber)
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
		prefix := line.Prefix
		if prefix == "" {
			prefix = " "
		}
		d.drawText(ctx, x, row, prefix, prefixRole)
	}
	x++
	if x < ctx.Width {
		ctx.DrawText(x, row, " ")
	}
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

func (d DiffView) scrollY(delta int) {
	if d.State == nil {
		return
	}
	d.State.MoveY(delta, d.gutterWidth())
}

func (d DiffView) scrollX(delta int) {
	if d.State == nil {
		return
	}
	d.State.MoveX(delta, d.gutterWidth())
}

func (d DiffView) pageUp() {
	if d.State == nil {
		return
	}
	d.State.PageUp(d.gutterWidth())
}

func (d DiffView) pageDown() {
	if d.State == nil {
		return
	}
	d.State.PageDown(d.gutterWidth())
}

func (d DiffView) halfPageUp() {
	if d.State == nil {
		return
	}
	d.State.HalfPageUp(d.gutterWidth())
}

func (d DiffView) halfPageDown() {
	if d.State == nil {
		return
	}
	d.State.HalfPageDown(d.gutterWidth())
}

func (d DiffView) toTop() {
	if d.State == nil {
		return
	}
	d.State.GoTop(d.gutterWidth())
}

func (d DiffView) toBottom() {
	if d.State == nil {
		return
	}
	d.State.GoBottom(d.gutterWidth())
}

func (d DiffView) gutterWidth() int {
	return renderedGutterWidth(d.currentRendered())
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
