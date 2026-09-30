package terma

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type floatingOrderScene struct {
	switcher Signal[bool]
}

func (s *floatingOrderScene) Build(ctx BuildContext) Widget {
	panel := func(id, label string) Widget {
		return Button{ID: id, Label: label, Style: Style{Width: Cells(32), Height: Cells(8), BackgroundColor: ctx.Theme().Surface}}
	}
	return Column{Children: []Widget{
		Floating{Visible: true, Config: FloatConfig{Modal: true, Position: FloatPositionCenter}, Child: Column{Children: []Widget{
			Text{Content: "Parent modal"},
			Floating{Visible: true, Config: FloatConfig{Modal: true, Position: FloatPositionCenter}, Child: panel("confirmation", "Nested confirmation")},
		}}},
		Floating{Visible: s.switcher.Get(), Config: FloatConfig{Modal: true, Position: FloatPositionCenter}, Child: panel("switcher", "Global switcher")},
	}}
}

func TestFloatingOrder_LaterSiblingAboveNestedModal(t *testing.T) {
	scene := &floatingOrderScene{switcher: NewSignal(true)}
	renderer, _ := geometryRenderer(t, 60, 20)
	renderer.Render(scene)
	require.Contains(t, renderer.ScreenText(), "Global switcher", "a later sibling must cover the earlier modal and its nested confirmation")
	require.NotContains(t, renderer.ScreenText(), "Nested confirmation")
	entry := renderer.WidgetByID("switcher")
	require.NotNil(t, entry)
	require.Equal(t, "switcher", renderer.PointerOwnerAt(entry.Bounds.X, entry.Bounds.Y).ID, "pointer routing must agree with paint order")
	AssertSnapshot(t, scene, 60, 20)

	scene.switcher.Set(false)
	renderer.Update(scene)
	require.Contains(t, renderer.ScreenText(), "Nested confirmation")
	require.NotContains(t, renderer.ScreenText(), "Global switcher")
	scene.switcher.Set(true)
	renderer.Update(scene)
	require.Contains(t, renderer.ScreenText(), "Global switcher", "reopening must preserve the same ordering on retained frames")
}

func FuzzFloatingOrder_NestedSiblings(f *testing.F) {
	f.Add(uint8(0), uint8(0), uint8(0))
	f.Add(uint8(4), uint8(0), uint8(1))
	f.Add(uint8(5), uint8(3), uint8(2))
	f.Fuzz(func(t *testing.T, first, second, third uint8) {
		chain := func(depth uint8, id string) Widget {
			var child Widget = Button{ID: id, Label: id, Style: Style{Width: Cells(30), Height: Cells(4), BackgroundColor: Black}}
			for i := 0; i <= int(depth%6); i++ {
				child = Floating{Visible: true, Config: FloatConfig{Modal: true, Position: FloatPositionCenter}, Child: child}
			}
			return child
		}
		root := Column{Children: []Widget{chain(first, "first"), chain(second, "second"), chain(third, "last")}}
		renderer, _ := geometryRenderer(t, 50, 12)
		renderer.Render(root)
		require.Contains(t, renderer.ScreenText(), "[last]")
		require.NotContains(t, renderer.ScreenText(), "[first]")
		require.NotContains(t, renderer.ScreenText(), "[second]")
		bounds := renderer.WidgetByID("last").Bounds
		require.Equal(t, "last", renderer.PointerOwnerAt(bounds.X, bounds.Y).ID)
	})
}
