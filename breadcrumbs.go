package terma

// Breadcrumbs renders a clickable breadcrumb path.
type Breadcrumbs struct {
	ID        string
	Path      []string
	OnSelect  func(index int) // Click to navigate
	Separator string          // Default: ">"
	Width     Dimension // Deprecated: use Style.Width
	Height    Dimension // Deprecated: use Style.Height
	Style     Style
}

// Build renders the breadcrumb path as a row of text segments.
func (b Breadcrumbs) Build(ctx BuildContext) Widget {
	if len(b.Path) == 0 {
		return EmptyWidget{}
	}

	separator := b.Separator
	if separator == "" {
		separator = ">"
	}
	separator = " " + separator + " "

	children := make([]Widget, 0, len(b.Path)*2-1)
	for i, label := range b.Path {
		style := b.Style
		style.Width = Dimension{}
		style.Height = Dimension{}
		if style.ForegroundColor == nil || !style.ForegroundColor.IsSet() {
			if b.OnSelect != nil && i < len(b.Path)-1 {
				style.ForegroundColor = ctx.Theme().Link
			} else {
				style.ForegroundColor = ctx.Theme().Text
			}
		}

		index := i
		text := Text{
			Content: label,
			Style:   style,
		}
		if b.OnSelect != nil {
			text.Click = func(MouseEvent) {
				b.OnSelect(index)
			}
		}
		children = append(children, text)

		if i < len(b.Path)-1 {
			sepStyle := b.Style
			sepStyle.Width = Dimension{}
			sepStyle.Height = Dimension{}
			if sepStyle.ForegroundColor == nil || !sepStyle.ForegroundColor.IsSet() {
				sepStyle.ForegroundColor = ctx.Theme().TextMuted
			}
			children = append(children, Text{Content: separator, Style: sepStyle})
		}
	}

	rowStyle := b.Style
	if rowStyle.Padding == (EdgeInsets{}) {
		rowStyle.Padding = EdgeInsetsTRBL(0, 1, 0, 1)
	}
	if rowStyle.Width.IsUnset() {
		rowStyle.Width = b.Width
	}
	if rowStyle.Height.IsUnset() {
		rowStyle.Height = b.Height
	}
	return Row{
		ID:         b.ID,
		CrossAlign: CrossAxisCenter,
		Children:   children,
		Style:      rowStyle,
	}
}
