package terma

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Layout chooses the tab presentation after Build has read the previous state.
// This mirrors responsive tab bars which schedule reactive updates in OnLayout.
type headlessDispatchTabs struct {
	label    Signal[string]
	inFrame  bool
	midFrame bool
	calls    int
}

func (w *headlessDispatchTabs) Build(BuildContext) Widget {
	w.inFrame = true
	return headlessDispatchTabsRow{
		Row: Row{Style: Style{Width: Flex(1), Height: Cells(1)}, Children: []Widget{
			Text{Content: w.label.Get()},
			headlessDispatchFrameEnd{Text: Text{Content: " "}, owner: w},
		}},
		owner: w,
	}
}

type headlessDispatchTabsRow struct {
	Row
	owner *headlessDispatchTabs
}

func (w headlessDispatchTabsRow) Build(BuildContext) Widget { return w }
func (w headlessDispatchTabsRow) ChildWidgets() []Widget    { return w.Children }
func (w headlessDispatchTabsRow) OnLayout(_ BuildContext, metrics LayoutMetrics) {
	label := "compact"
	if metrics.Box().Width >= 20 {
		label = "expanded"
	}
	if w.owner.label.Peek() == label {
		return
	}
	Dispatch(func() {
		w.owner.calls++
		w.owner.midFrame = w.owner.midFrame || w.owner.inFrame
		w.owner.label.Set(label)
	})
}

type headlessDispatchFrameEnd struct {
	Text
	owner *headlessDispatchTabs
}

func (w headlessDispatchFrameEnd) Build(BuildContext) Widget { return w }
func (w headlessDispatchFrameEnd) Render(*RenderContext)     { w.owner.inFrame = false }

func TestHeadlessDispatchSettlesTabLayoutAndResize(t *testing.T) {
	for _, method := range []string{"Render", "RenderWithSize", "Update"} {
		t.Run(method, func(t *testing.T) {
			widget := &headlessDispatchTabs{label: NewSignal("initial")}
			buffer := uv.NewBuffer(20, 2)
			renderer := NewRenderer(buffer, 20, 2, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
			draw := func() {
				switch method {
				case "Render":
					renderer.Render(widget)
				case "RenderWithSize":
					renderer.RenderWithSize(widget)
				case "Update":
					renderer.Update(widget)
				}
			}
			draw()
			assert.False(t, widget.midFrame, "Dispatch must run after painting finishes")
			assert.Equal(t, "expanded", strings.TrimSpace(strings.Split(renderer.ScreenText(), "\n")[0]), "a single caller render must show the layout-driven state")
			renderer.Resize(10, 2)
			draw()
			assert.False(t, widget.midFrame, "resize callback must run after painting finishes")
			assert.Equal(t, "compact", strings.TrimSpace(strings.Split(renderer.ScreenText(), "\n")[0]), "resize must settle without a second caller render")
			require.Equal(t, 2, widget.calls)
		})
	}
}

// A leaf schedules from either OnLayout or paint, which covers both frame
// phases without mutating signals in Build.
type headlessDispatchLeaf struct {
	Text
	onLayout func()
	onPaint  func()
}

func (w headlessDispatchLeaf) Build(BuildContext) Widget { return w }
func (w headlessDispatchLeaf) OnLayout(BuildContext, LayoutMetrics) {
	if w.onLayout != nil {
		w.onLayout()
	}
}
func (w headlessDispatchLeaf) Render(ctx *RenderContext) {
	w.Text.Render(ctx)
	if w.onPaint != nil {
		w.onPaint()
	}
}

func TestHeadlessDispatchKeepsAppQueueOwnership(t *testing.T) {
	installTestAppRuntime(t)
	widget := &headlessDispatchTabs{label: NewSignal("initial")}
	RenderToBuffer(widget, 20, 2)
	require.Equal(t, "initial", widget.label.Peek())
	require.Zero(t, widget.calls, "a headless render must not drain a running app's queue")
	drainPendingDispatches()
	require.Equal(t, "expanded", widget.label.Peek())
	require.False(t, widget.midFrame)
}

func TestHeadlessDispatchFIFOAndPlainFields(t *testing.T) {
	for _, phase := range []string{"layout", "paint"} {
		t.Run(phase, func(t *testing.T) {
			var order []int
			label := "initial"
			scheduled := false
			schedule := func() {
				if scheduled {
					return
				}
				scheduled = true
				Dispatch(func() {
					order = append(order, 1)
					Dispatch(func() { order = append(order, 3); label = "done" })
				})
				Dispatch(func() { order = append(order, 2) })
			}
			root := headlessDispatchPlainRoot{build: func() Widget {
				leaf := headlessDispatchLeaf{Text: Text{Content: label}}
				if phase == "layout" {
					leaf.onLayout = schedule
				} else {
					leaf.onPaint = schedule
				}
				return leaf
			}}
			buf := RenderToBuffer(root, 20, 2)
			require.Equal(t, []int{1, 2, 3}, order)
			require.Equal(t, "d", buf.CellAt(0, 0).Content)
			require.Equal(t, "e", buf.CellAt(3, 0).Content)
			require.True(t, buf.CellAt(4, 0) == nil || buf.CellAt(4, 0).Content == " ", "settling must clear shortened content")
		})
	}
}

type headlessDispatchPlainRoot struct{ build func() Widget }

func (w headlessDispatchPlainRoot) Build(BuildContext) Widget { return w.build() }

func TestHeadlessDispatchBoundsSelfSchedulingAndRestoresImmediateDispatch(t *testing.T) {
	for _, recursive := range []bool{false, true} {
		t.Run(map[bool]string{false: "every-frame", true: "recursive-callback"}[recursive], func(t *testing.T) {
			callbacks := 0
			var repeat func()
			repeat = func() {
				callbacks++
				if recursive {
					Dispatch(repeat)
				}
			}
			scheduled := false
			widget := headlessDispatchLeaf{Text: Text{Content: "test"}, onLayout: func() {
				if recursive && scheduled {
					return
				}
				scheduled = true
				Dispatch(repeat)
			}}
			require.PanicsWithValue(t, "terma: headless Dispatch did not settle within 16 frames", func() {
				RenderToBuffer(widget, 20, 2)
			})
			require.Equal(t, 15, callbacks)
			immediate := false
			Dispatch(func() { immediate = true })
			require.True(t, immediate, "panic must restore outside-render Dispatch behavior")
		})
	}
}

func TestHeadlessDispatchCallbackPanicRestoresImmediateDispatch(t *testing.T) {
	widget := headlessDispatchLeaf{Text: Text{Content: "test"}, onLayout: func() {
		Dispatch(func() { panic("callback failed") })
	}}
	require.PanicsWithValue(t, "callback failed", func() { RenderToBuffer(widget, 20, 2) })
	immediate := false
	Dispatch(func() { immediate = true })
	require.True(t, immediate)
}

func TestHeadlessDispatchReturnsSettledLayoutDimensions(t *testing.T) {
	label := "x"
	scheduled := false
	root := headlessDispatchPlainRoot{build: func() Widget {
		return headlessDispatchLeaf{Text: Text{Content: label}, onLayout: func() {
			if scheduled {
				return
			}
			scheduled = true
			Dispatch(func() { label = "expanded" })
		}}
	}}
	buf, width, height := RenderToBufferWithSize(root, 20, 2)
	require.Equal(t, len("expanded"), width)
	require.Equal(t, 1, height)
	require.Equal(t, "d", buf.CellAt(7, 0).Content)
}
