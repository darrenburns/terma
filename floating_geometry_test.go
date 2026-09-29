package terma

import (
	"fmt"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

type geometryAnchor struct {
	width, left Signal[int]
}

func (a *geometryAnchor) WidgetID() string { return "anchor" }

func (a *geometryAnchor) Build(BuildContext) Widget {
	return Text{ID: "anchor", Content: "anchor", Width: Cells(a.width.Get()), Style: Style{Margin: EdgeInsets{Left: a.left.Get()}}}
}

type geometryScene struct {
	anchor   *geometryAnchor
	label    Signal[string]
	geometry FloatGeometry
	builds   int
}

func newGeometryScene() *geometryScene {
	return &geometryScene{anchor: &geometryAnchor{width: NewSignal(12), left: NewSignal(3)}, label: NewSignal("summary")}
}
func (s *geometryScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Spacer{Height: Cells(2)}, s.anchor,
		Floating{Visible: true, Config: FloatConfig{AnchorID: "anchor", Anchor: AnchorBottomLeft},
			BuildChild: func(_ BuildContext, g FloatGeometry) Widget {
				s.geometry = g
				s.builds++
				if !g.AnchorFound {
					return nil
				}
				return Text{ID: "summary", Content: s.label.Get(), Width: Cells(g.AnchorBounds.Width)}
			},
		},
	}}
}
func geometryRenderer(t *testing.T, width, height int) (*Renderer, *uv.Buffer) {
	t.Helper()
	buffer := uv.NewBuffer(width, height)
	renderer := NewRenderer(reactivityScreen{buffer}, width, height, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	t.Cleanup(func() {
		if renderer.rootNode != nil {
			renderer.rootNode.dispose()
		}
		for _, f := range renderer.retainedFloats {
			f.root.dispose()
		}
	})
	return renderer, buffer
}
func TestFloatingGeometry_FirstFrame(t *testing.T) {
	renderer, _ := geometryRenderer(t, 30, 8)
	scene := newGeometryScene()
	renderer.Render(scene)
	require.Contains(t, renderer.ScreenText(), "summary", "summary must appear on the first frame without a geometry callback or second render")
	require.Equal(t, FloatGeometry{
		Screen: Rect{Width: 30, Height: 8}, AnchorFound: true,
		AnchorBounds:        Rect{X: 3, Y: 2, Width: 12, Height: 1},
		AnchorVisibleBounds: Rect{X: 3, Y: 2, Width: 12, Height: 1},
	}, scene.geometry)
	require.Equal(t, Rect{X: 3, Y: 3, Width: 12, Height: 1}, renderer.WidgetByID("summary").Bounds)
	require.Equal(t, 1, scene.builds)
}
func TestFloatingGeometry_MovementResizeAndReactiveReads(t *testing.T) {
	renderer, buffer := geometryRenderer(t, 30, 8)
	scene := newGeometryScene()
	renderer.Render(scene)
	scene.anchor.left.Set(8)
	scene.anchor.width.Set(15)
	renderer.Update(scene)
	require.Equal(t, Rect{X: 8, Y: 2, Width: 15, Height: 1}, scene.geometry.AnchorBounds)
	require.Equal(t, Rect{X: 8, Y: 3, Width: 15, Height: 1}, renderer.WidgetByID("summary").Bounds)
	require.Equal(t, 2, scene.builds, "owner Build was reused, but new geometry rebuilds content")
	scene.label.Set("changed")
	renderer.Update(scene)
	require.Contains(t, renderer.ScreenText(), "changed", "signal reads in the deferred builder must subscribe")
	require.Equal(t, 3, scene.builds)
	buffer.Resize(18, 7)
	renderer.Resize(18, 7)
	renderer.Update(scene)
	require.Equal(t, Rect{Width: 18, Height: 7}, scene.geometry.Screen)
	require.Equal(t, renderer.WidgetByID("anchor").Bounds, scene.geometry.AnchorBounds)
	require.Equal(t, scene.geometry.AnchorBounds.Width, renderer.WidgetByID("summary").Bounds.Width)
	require.Equal(t, 4, scene.builds)
	renderer.Update(scene)
	require.Equal(t, 4, scene.builds, "geometry is a snapshot, not a self-invalidating signal")
}
func TestFloatingGeometry_MissingAnchorAndNilContent(t *testing.T) {
	renderer, _ := geometryRenderer(t, 30, 8)
	var got FloatGeometry
	renderer.Render(Floating{Visible: true,
		Config:     FloatConfig{AnchorID: "missing", Anchor: AnchorBottomLeft},
		Child:      Text{Content: "must not appear"},
		BuildChild: func(_ BuildContext, g FloatGeometry) Widget { got = g; return nil },
	})
	require.False(t, got.AnchorFound)
	require.Equal(t, Rect{}, got.AnchorBounds)
	require.Equal(t, Rect{}, got.AnchorVisibleBounds)
	require.NotContains(t, renderer.ScreenText(), "must not appear", "BuildChild takes precedence and nil is empty content")
}
func TestFloatingGeometry_NestedOverlayAnchor(t *testing.T) {
	renderer, _ := geometryRenderer(t, 30, 8)
	var got FloatGeometry
	renderer.Render(Floating{Visible: true, Config: FloatConfig{Offset: Offset{X: 4, Y: 2}},
		Child: Column{Children: []Widget{
			Text{ID: "nested-anchor", Content: "outer", Width: Cells(10)},
			Floating{Visible: true, Config: FloatConfig{AnchorID: "nested-anchor", Anchor: AnchorBottomLeft},
				BuildChild: func(_ BuildContext, g FloatGeometry) Widget {
					got = g
					return Text{ID: "nested-summary", Content: "inner", Width: Cells(g.AnchorBounds.Width)}
				},
			},
		}},
	})
	require.True(t, got.AnchorFound)
	require.Equal(t, Rect{X: 4, Y: 2, Width: 10, Height: 1}, got.AnchorBounds)
	require.Equal(t, Rect{X: 4, Y: 3, Width: 10, Height: 1}, renderer.WidgetByID("nested-summary").Bounds)
	require.Contains(t, renderer.ScreenText(), "inner")
}
func TestFloatingGeometry_ScrollingAnchor(t *testing.T) {
	renderer, _ := geometryRenderer(t, 30, 8)
	state := NewScrollState()
	var got FloatGeometry
	var builds int
	root := Column{Children: []Widget{
		Scrollable{ID: "viewport", Width: Cells(15), Height: Cells(3), State: state,
			Child: Column{Children: []Widget{
				Text{ID: "scroll-anchor", Content: "anchor", Width: Cells(10), Height: Cells(3)},
				Text{Content: "after", Height: Cells(8)},
			}},
		},
		Floating{Visible: true, Config: FloatConfig{AnchorID: "scroll-anchor", Anchor: AnchorBottomLeft},
			BuildChild: func(_ BuildContext, g FloatGeometry) Widget {
				got = g
				builds++
				if !g.AnchorFound || g.AnchorVisibleBounds.Width == 0 || g.AnchorVisibleBounds.Height == 0 {
					return nil
				}
				return Text{ID: "scroll-summary", Content: fmt.Sprintf("y=%d", g.AnchorBounds.Y)}
			},
		},
	}}
	renderer.Render(root)
	require.Equal(t, Rect{X: 0, Y: 0, Width: 10, Height: 3}, got.AnchorBounds)
	state.SetOffset(1)
	renderer.Update(root)
	require.Equal(t, Rect{X: 0, Y: -1, Width: 10, Height: 3}, got.AnchorBounds, "paint-only scrolling must update deferred geometry")
	require.Equal(t, Rect{X: 0, Y: 0, Width: 10, Height: 2}, got.AnchorVisibleBounds)
	require.Equal(t, Rect{X: 0, Y: 2, Width: 4, Height: 1}, renderer.WidgetByID("scroll-summary").Bounds)
	require.Equal(t, 2, builds)
	state.SetOffset(4)
	renderer.Update(root)
	require.True(t, !got.AnchorFound || got.AnchorVisibleBounds.Height == 0)
	require.NotContains(t, renderer.ScreenText(), "y=", "the builder can hide content when the anchor leaves the viewport")
}
