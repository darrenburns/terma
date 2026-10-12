package terma

import (
	"testing"

	uv "github.com/darrenburns/ultraviolet"
	"github.com/stretchr/testify/require"
)

// EdgeInsets Tests

func TestEdgeInsetsAll_CreatesUniformInsets(t *testing.T) {
	insets := EdgeInsetsAll(10)

	if insets.Top != 10 {
		t.Errorf("expected Top = 10, got %d", insets.Top)
	}
	if insets.Right != 10 {
		t.Errorf("expected Right = 10, got %d", insets.Right)
	}
	if insets.Bottom != 10 {
		t.Errorf("expected Bottom = 10, got %d", insets.Bottom)
	}
	if insets.Left != 10 {
		t.Errorf("expected Left = 10, got %d", insets.Left)
	}
}

func TestEdgeInsetsAll_ZeroValue(t *testing.T) {
	insets := EdgeInsetsAll(0)

	if insets.Top != 0 || insets.Right != 0 || insets.Bottom != 0 || insets.Left != 0 {
		t.Errorf("expected all values to be 0, got %+v", insets)
	}
}

func TestEdgeInsetsXY_CreatesHorizontalVerticalInsets(t *testing.T) {
	insets := EdgeInsetsXY(15, 10)

	if insets.Top != 10 {
		t.Errorf("expected Top = 10 (vertical), got %d", insets.Top)
	}
	if insets.Right != 15 {
		t.Errorf("expected Right = 15 (horizontal), got %d", insets.Right)
	}
	if insets.Bottom != 10 {
		t.Errorf("expected Bottom = 10 (vertical), got %d", insets.Bottom)
	}
	if insets.Left != 15 {
		t.Errorf("expected Left = 15 (horizontal), got %d", insets.Left)
	}
}

func TestEdgeInsetsXY_ZeroValues(t *testing.T) {
	insets := EdgeInsetsXY(0, 0)

	if insets.Top != 0 || insets.Right != 0 || insets.Bottom != 0 || insets.Left != 0 {
		t.Errorf("expected all values to be 0, got %+v", insets)
	}
}

func TestEdgeInsetsTRBL_CreatesCustomInsets(t *testing.T) {
	insets := EdgeInsetsTRBL(1, 2, 3, 4)

	if insets.Top != 1 {
		t.Errorf("expected Top = 1, got %d", insets.Top)
	}
	if insets.Right != 2 {
		t.Errorf("expected Right = 2, got %d", insets.Right)
	}
	if insets.Bottom != 3 {
		t.Errorf("expected Bottom = 3, got %d", insets.Bottom)
	}
	if insets.Left != 4 {
		t.Errorf("expected Left = 4, got %d", insets.Left)
	}
}

func TestEdgeInsets_Horizontal_SumsLeftAndRight(t *testing.T) {
	tests := []struct {
		name     string
		insets   EdgeInsets
		expected int
	}{
		{"uniform", EdgeInsetsAll(5), 10},
		{"different values", EdgeInsetsTRBL(1, 7, 3, 8), 15},
		{"zero", EdgeInsetsAll(0), 0},
		{"asymmetric", EdgeInsetsXY(12, 5), 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.insets.Horizontal()
			if got != tt.expected {
				t.Errorf("Horizontal() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestEdgeInsets_Vertical_SumsTopAndBottom(t *testing.T) {
	tests := []struct {
		name     string
		insets   EdgeInsets
		expected int
	}{
		{"uniform", EdgeInsetsAll(5), 10},
		{"different values", EdgeInsetsTRBL(6, 2, 9, 3), 15},
		{"zero", EdgeInsetsAll(0), 0},
		{"asymmetric", EdgeInsetsXY(8, 7), 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.insets.Vertical()
			if got != tt.expected {
				t.Errorf("Vertical() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestEdgeInsets_ZeroValue(t *testing.T) {
	var insets EdgeInsets

	if insets.Top != 0 || insets.Right != 0 || insets.Bottom != 0 || insets.Left != 0 {
		t.Errorf("expected zero value EdgeInsets to be all zeros, got %+v", insets)
	}
	if insets.Horizontal() != 0 {
		t.Errorf("expected Horizontal() = 0, got %d", insets.Horizontal())
	}
	if insets.Vertical() != 0 {
		t.Errorf("expected Vertical() = 0, got %d", insets.Vertical())
	}
}

// Border Tests

func TestBorder_Width_ReturnsZeroForNoBorder(t *testing.T) {
	var border Border

	if border.Width() != 0 {
		t.Errorf("expected Width() = 0 for zero value border, got %d", border.Width())
	}
}

func TestBorder_Width_ReturnsOneForSquareBorder(t *testing.T) {
	border := SquareBorder(RGB(255, 255, 255))

	if border.Width() != 1 {
		t.Errorf("expected Width() = 1 for square border, got %d", border.Width())
	}
}

func TestBorder_Width_ReturnsOneForRoundedBorder(t *testing.T) {
	border := RoundedBorder(RGB(255, 255, 255))

	if border.Width() != 1 {
		t.Errorf("expected Width() = 1 for rounded border, got %d", border.Width())
	}
}

func TestBorder_IsZero_TrueForNoBorder(t *testing.T) {
	var border Border

	if !border.IsZero() {
		t.Error("expected IsZero() = true for zero value border")
	}
}

func TestBorder_IsZero_FalseForSquareBorder(t *testing.T) {
	border := SquareBorder(RGB(255, 255, 255))

	if border.IsZero() {
		t.Error("expected IsZero() = false for square border")
	}
}

func TestBorder_IsZero_FalseForRoundedBorder(t *testing.T) {
	border := RoundedBorder(RGB(255, 255, 255))

	if border.IsZero() {
		t.Error("expected IsZero() = false for rounded border")
	}
}

// Style Tests

func TestStyle_IsZero_TrueForZeroValue(t *testing.T) {
	var style Style

	if !style.IsZero() {
		t.Error("expected IsZero() = true for zero value style")
	}
}

func TestStyle_IsZero_FalseWhenForegroundColorSet(t *testing.T) {
	style := Style{ForegroundColor: RGB(255, 0, 0)}

	if style.IsZero() {
		t.Error("expected IsZero() = false when foreground color is set")
	}
}

func TestStyle_IsZero_FalseWhenBackgroundColorSet(t *testing.T) {
	style := Style{BackgroundColor: RGB(0, 255, 0)}

	if style.IsZero() {
		t.Error("expected IsZero() = false when background color is set")
	}
}

func TestStyle_IsZero_FalseWhenReverseSet(t *testing.T) {
	style := Style{Reverse: true}

	if style.IsZero() {
		t.Error("expected IsZero() = false when reverse is set")
	}
}

func TestStyle_IsZero_FalseWhenPaddingSet(t *testing.T) {
	style := Style{Padding: EdgeInsetsAll(5)}

	if style.IsZero() {
		t.Error("expected IsZero() = false when padding is set")
	}
}

func TestStyle_IsZero_FalseWhenMarginSet(t *testing.T) {
	style := Style{Margin: EdgeInsetsAll(5)}

	if style.IsZero() {
		t.Error("expected IsZero() = false when margin is set")
	}
}

func TestStyle_IsZero_FalseWhenBorderSet(t *testing.T) {
	style := Style{Border: SquareBorder(RGB(255, 255, 255))}

	if style.IsZero() {
		t.Error("expected IsZero() = false when border is set")
	}
}

// New Border Type Tests

func TestBorder_AllStyles_WidthReturnsOne(t *testing.T) {
	white := RGB(255, 255, 255)
	tests := []struct {
		name   string
		border Border
	}{
		{"Square", SquareBorder(white)},
		{"Rounded", RoundedBorder(white)},
		{"Double", DoubleBorder(white)},
		{"Heavy", HeavyBorder(white)},
		{"Dashed", DashedBorder(white)},
		{"Ascii", AsciiBorder(white)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.border.Width() != 1 {
				t.Errorf("%s border Width() = %d, want 1", tt.name, tt.border.Width())
			}
		})
	}
}

func TestBorder_AllStyles_IsZeroReturnsFalse(t *testing.T) {
	white := RGB(255, 255, 255)
	tests := []struct {
		name   string
		border Border
	}{
		{"Square", SquareBorder(white)},
		{"Rounded", RoundedBorder(white)},
		{"Double", DoubleBorder(white)},
		{"Heavy", HeavyBorder(white)},
		{"Dashed", DashedBorder(white)},
		{"Ascii", AsciiBorder(white)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.border.IsZero() {
				t.Errorf("%s border IsZero() = true, want false", tt.name)
			}
		})
	}
}

func TestGetBorderCharSet_ReturnsCorrectCharacters(t *testing.T) {
	tests := []struct {
		style   BorderStyle
		topLeft string
		top     string
		left    string
	}{
		{BorderSquare, "┌", "─", "│"},
		{BorderRounded, "╭", "─", "│"},
		{BorderDouble, "╔", "═", "║"},
		{BorderHeavy, "┏", "━", "┃"},
		{BorderDashed, "┏", "╍", "╏"},
		{BorderAscii, "+", "-", "|"},
		{BorderOuter, "▛", "▀", "▌"},
		{BorderTall, "▊", "▔", "▊"},
	}

	for _, tt := range tests {
		t.Run(tt.topLeft, func(t *testing.T) {
			chars := GetBorderCharSet(tt.style)
			if chars.TopLeft != tt.topLeft {
				t.Errorf("TopLeft = %q, want %q", chars.TopLeft, tt.topLeft)
			}
			if chars.Top != tt.top {
				t.Errorf("Top = %q, want %q", chars.Top, tt.top)
			}
			if chars.Left != tt.left {
				t.Errorf("Left = %q, want %q", chars.Left, tt.left)
			}
		})
	}
}

func TestGetBorderCharSet_BorderNone_ReturnsEmptyCharSet(t *testing.T) {
	chars := GetBorderCharSet(BorderNone)

	if chars.TopLeft != "" {
		t.Errorf("expected empty TopLeft for BorderNone, got %q", chars.TopLeft)
	}
	if chars.Top != "" {
		t.Errorf("expected empty Top for BorderNone, got %q", chars.Top)
	}
}

func TestBorder_Constructors_SetCorrectStyle(t *testing.T) {
	white := RGB(255, 255, 255)
	tests := []struct {
		name     string
		border   Border
		expected BorderStyle
	}{
		{"Square", SquareBorder(white), BorderSquare},
		{"Rounded", RoundedBorder(white), BorderRounded},
		{"Double", DoubleBorder(white), BorderDouble},
		{"Heavy", HeavyBorder(white), BorderHeavy},
		{"Dashed", DashedBorder(white), BorderDashed},
		{"Ascii", AsciiBorder(white), BorderAscii},
		{"Outer", OuterBorder(white), BorderOuter},
		{"Tall", TallBorder(white), BorderTall},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.border.Style != tt.expected {
				t.Errorf("%s border Style = %d, want %d", tt.name, tt.border.Style, tt.expected)
			}
		})
	}
}

func TestBorder_Constructors_SetColor(t *testing.T) {
	red := RGB(255, 0, 0)
	border := DoubleBorder(red)

	if border.Color != red {
		t.Errorf("expected border color to be set")
	}
}

func TestBorder_Constructors_AcceptDecorations(t *testing.T) {
	white := RGB(255, 255, 255)
	title := BorderTitle("Test")
	subtitle := BorderSubtitle("Sub")

	border := DoubleBorder(white, title, subtitle)

	if len(border.Decorations) != 2 {
		t.Errorf("expected 2 decorations, got %d", len(border.Decorations))
	}
}

func TestDrawBorder_UsesEachEdgesOwnCharacter(t *testing.T) {
	widget := Column{Width: Cells(4), Height: Cells(3), Style: Style{Border: OuterBorder(RGB(255, 255, 255))}}
	buf := renderToBufferWithFocus(widget, 4, 3, "")
	want := []string{"▛▀▀▜", "▌  ▐", "▙▄▄▟"}
	for y, row := range want {
		for x, ch := range []rune(row) {
			if got := buf.CellAt(x, y).Content; got != string(ch) {
				t.Errorf("cell %d,%d = %q, want %q", x, y, got, string(ch))
			}
		}
	}
}

func TestTallBorder_Rendering(t *testing.T) {
	inner, border := RGB(30, 50, 70), RGB(210, 160, 90)
	for _, tt := range []struct {
		name  string
		outer ColorProvider
	}{
		{"terminal background", nil},
		{"solid parent", RGB(90, 20, 40)},
		{"gradient parent", NewGradient(RGB(100, 0, 0), RGB(0, 0, 200)).WithAngle(90)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			widget := Column{
				Width: Cells(14), Height: Cells(4),
				Style: Style{BackgroundColor: tt.outer, Padding: EdgeInsetsAll(1)},
				Children: []Widget{Column{
					Width: Cells(10), Height: Cells(2),
					Style: Style{BackgroundColor: inner, Border: TallBorder(border,
						BorderTitle("T"), BorderSubtitleRight("S"))},
					Children: []Widget{Text{Content: "x"}},
				}},
			}
			buf := renderToBufferWithFocus(widget, 16, 6, "")
			want := []string{"▊ T ▔▔▔▔▔▔▔▎", "▊x         ▎", "▊          ▎", "▊▁▁▁▁▁▁▁ S ▎"}
			for y, row := range want {
				for x, ch := range []rune(row) {
					cell := buf.CellAt(x+1, y+1)
					require.Equal(t, string(ch), cell.Content, "glyph at %d,%d", x, y)
					bg := inner
					if x == 0 || x == 11 {
						bg = Color{}
						if tt.outer != nil {
							bg = tt.outer.ColorAt(16, 6, x+1, y+1)
						}
					}
					require.Equal(t, bg, FromANSI(cell.Style.Bg), "background at %d,%d", x, y)
					require.Equal(t, x == 0, cell.Style.Attrs&uv.AttrReverse != 0, "reverse at %d,%d", x, y)
					if x == 0 || x == 11 || y == 0 || y == 3 {
						require.Equal(t, border, FromANSI(cell.Style.Fg), "border color at %d,%d", x, y)
					}
				}
			}
		})
	}
}

func TestTallBorder_FloatingBackground(t *testing.T) {
	for _, mode := range []string{"floating", "modal", "overlapping floats"} {
		t.Run(mode, func(t *testing.T) {
			const width, height = 24, 10
			outer := NewGradient(RGB(100, 20, 40), RGB(20, 60, 200)).WithAngle(35)
			inner := NewSignal(RGB(30, 50, 70))
			border := RGB(210, 160, 90)
			lower := RGB(60, 100, 80)
			widget := Stack{
				Width: Cells(width), Height: Cells(height),
				Style: Style{BackgroundColor: Magenta},
				Children: []Widget{
					Text{Width: Cells(width), Height: Cells(height), Style: Style{BackgroundColor: outer}},
					Floating{
						Visible: mode == "overlapping floats",
						Config:  FloatConfig{Offset: Offset{X: 1, Y: 1}},
						Child:   Text{Width: Cells(8), Height: Cells(8), Style: Style{BackgroundColor: lower}},
					},
					Floating{
						Visible: true,
						Config:  FloatConfig{Offset: Offset{X: 3, Y: 2}, Modal: mode == "modal", BackdropColor: Black.WithAlpha(.5)},
						BuildChild: func(BuildContext, FloatGeometry) Widget {
							return Text{
								ID: "popup", Content: "Content", Width: Cells(12), Height: Cells(3),
								Style: Style{BackgroundColor: inner.Get(), Border: TallBorder(border,
									BorderTitle("T"), BorderSubtitleRight("S"))},
							}
						},
					},
				},
			}
			renderer, buf := geometryRenderer(t, width, height)
			renderer.Render(widget)
			check := func() {
				t.Helper()
				bounds := renderer.WidgetByID("popup").Bounds
				for y := bounds.Y; y < bounds.Y+bounds.Height; y++ {
					for _, x := range []int{bounds.X, bounds.X + bounds.Width - 1} {
						cell := buf.CellAt(x, y)
						bg := outer.ColorAt(width, height, x, y)
						if mode == "modal" {
							bg = Black.WithAlpha(.5).BlendOver(bg)
						} else if mode == "overlapping floats" && x < 9 {
							bg = lower
						}
						require.Equal(t, bg, FromANSI(cell.Style.Bg), "outside background at %d,%d", x, y)
						require.Equal(t, border, FromANSI(cell.Style.Fg))
						require.Equal(t, x == bounds.X, cell.Style.Attrs&uv.AttrReverse != 0)
						if x == bounds.X {
							require.Equal(t, "▊", cell.Content)
						} else {
							require.Equal(t, "▎", cell.Content)
						}
					}
				}
				// Titles, horizontal strokes and content retain the popup's fill.
				for _, pos := range []Offset{{4, 2}, {5, 2}, {10, 2}, {4, 3}, {10, 6}, {15, 6}} {
					require.Equal(t, inner.Peek(), FromANSI(buf.CellAt(pos.X, pos.Y).Style.Bg), "inside background at %+v", pos)
				}
			}
			check()
			inner.Set(RGB(80, 35, 65))
			renderer.Update(widget)
			check()
		})
	}
}
