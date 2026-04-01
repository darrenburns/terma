package terma

import (
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

type buildCountWidget struct {
	builds *int
	child  Widget
}

func (w buildCountWidget) Build(ctx BuildContext) Widget {
	if w.builds != nil {
		(*w.builds)++
	}
	if w.child != nil {
		return w.child
	}
	return Text{Content: "x"}
}

type trackingScreen struct {
	*uv.Buffer
	touched map[[2]int]struct{}
}

func newTrackingScreen(width, height int) *trackingScreen {
	return &trackingScreen{
		Buffer:  uv.NewBuffer(width, height),
		touched: make(map[[2]int]struct{}),
	}
}

func (s *trackingScreen) SetCell(x, y int, c *uv.Cell) {
	s.touched[[2]int{x, y}] = struct{}{}
	s.Buffer.SetCell(x, y, c)
}

func (s *trackingScreen) ClearArea(area uv.Rectangle) {
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			s.touched[[2]int{x, y}] = struct{}{}
		}
	}
	s.Buffer.ClearArea(area)
}

func (s *trackingScreen) resetTouched() {
	s.touched = make(map[[2]int]struct{})
}

func newTestRenderer(screen CellBuffer, width, height int) *Renderer {
	return NewRenderer(
		screen,
		width,
		height,
		NewFocusManager(),
		NewAnySignal[Focusable](nil),
		NewAnySignal[Widget](nil),
	)
}

func TestSignal_PhaseSubscriptionsDriveDirtyLevels(t *testing.T) {
	t.Run("build", func(t *testing.T) {
		s := NewSignal(1)
		node := newWidgetNode(nil)
		withTrackedRead(node, readPhaseBuild, func() { _ = s.Get() })
		node.clearDirty()

		s.Set(2)

		if got := node.dirtyLevel(); got != DirtyBuild {
			t.Fatalf("expected DirtyBuild, got %v", got)
		}
	})

	t.Run("layout", func(t *testing.T) {
		s := NewSignal(1)
		node := newWidgetNode(nil)
		withTrackedRead(node, readPhaseLayout, func() { _ = s.Get() })
		node.clearDirty()

		s.Set(2)

		if got := node.dirtyLevel(); got != DirtyLayout {
			t.Fatalf("expected DirtyLayout, got %v", got)
		}
	})

	t.Run("paint", func(t *testing.T) {
		s := NewSignal(1)
		node := newWidgetNode(nil)
		withTrackedRead(node, readPhasePaint, func() { _ = s.Get() })
		node.clearDirty()

		s.Set(2)

		if got := node.dirtyLevel(); got != DirtyPaint {
			t.Fatalf("expected DirtyPaint, got %v", got)
		}
	})

	t.Run("escalates to highest phase", func(t *testing.T) {
		s := NewSignal(1)
		node := newWidgetNode(nil)
		withTrackedRead(node, readPhasePaint, func() { _ = s.Get() })
		withTrackedRead(node, readPhaseLayout, func() { _ = s.Get() })
		node.clearDirty()

		s.Set(2)

		if got := node.dirtyLevel(); got != DirtyLayout {
			t.Fatalf("expected DirtyLayout, got %v", got)
		}
	})
}

func TestSignal_DisposeClearsSubscriptions(t *testing.T) {
	s := NewSignal(1)
	node := newWidgetNode(nil)
	withTrackedRead(node, readPhasePaint, func() { _ = s.Get() })
	node.clearDirty()

	node.dispose()
	s.Set(2)

	if node.isDirty() {
		t.Fatal("disposed node should not be dirtied by future signal updates")
	}
	s.core.mu.Lock()
	defer s.core.mu.Unlock()
	if len(s.core.listeners) != 0 {
		t.Fatalf("expected listeners to be cleared, got %d", len(s.core.listeners))
	}
}

func TestRenderer_RowBuildsChildrenOncePerFrame(t *testing.T) {
	builds := 0
	widget := Row{
		Children: []Widget{
			buildCountWidget{builds: &builds},
		},
	}

	renderer := newTestRenderer(uv.NewBuffer(10, 1), 10, 1)
	renderer.Render(widget)

	if builds != 1 {
		t.Fatalf("expected child Build to run once, got %d", builds)
	}
}

func TestRenderer_SpinnerTickUsesPartialPaint(t *testing.T) {
	screen := newTrackingScreen(20, 1)
	renderer := newTestRenderer(screen, 20, 1)

	state := NewSpinnerState(SpinnerLine)
	builds := 0
	root := buildCountWidget{
		builds: &builds,
		child: Row{
			Children: []Widget{
				Spinner{State: state},
				Text{Content: " loading"},
			},
		},
	}

	renderer.Update(root)
	if renderer.lastFrameMode != rendererFrameFull {
		t.Fatalf("expected initial full render, got %q", renderer.lastFrameMode)
	}
	if builds != 1 {
		t.Fatalf("expected one build on initial render, got %d", builds)
	}

	screen.resetTouched()
	state.animation.Value().Set("|")
	renderer.Update(root)

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint on spinner tick, got %q", renderer.lastFrameMode)
	}
	if builds != 1 {
		t.Fatalf("spinner tick should not rebuild root, got %d builds", builds)
	}
	if renderer.lastLayoutCount != 0 {
		t.Fatalf("spinner tick should not relayout, got %d layout ops", renderer.lastLayoutCount)
	}
	for pos := range screen.touched {
		if pos[0] != 0 || pos[1] != 0 {
			t.Fatalf("expected only spinner cell to be touched, got (%d,%d)", pos[0], pos[1])
		}
	}
}

func TestRenderer_MixedWidthSpinnerFramesStayPaintOnly(t *testing.T) {
	screen := newTrackingScreen(20, 1)
	renderer := newTestRenderer(screen, 20, 1)

	state := NewSpinnerState(SpinnerStyle{
		Frames:    []string{"-", ">>"},
		FrameTime: 50 * time.Millisecond,
	})

	renderer.Update(Spinner{State: state})
	screen.resetTouched()

	state.animation.Value().Set(">>")
	renderer.Update(Spinner{State: state})

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint for mixed-width spinner tick, got %q", renderer.lastFrameMode)
	}
	if renderer.lastLayoutCount != 0 {
		t.Fatalf("mixed-width spinner tick should not relayout, got %d layout ops", renderer.lastLayoutCount)
	}
}

func TestRenderer_TextInputAutoWidthPromotesLayoutOnlyWhenIntrinsicWidthChanges(t *testing.T) {
	screen := newTrackingScreen(20, 1)
	renderer := newTestRenderer(screen, 20, 1)

	state := NewTextInputState("ab")
	widget := TextInput{State: state, Width: Auto}

	renderer.Update(widget)

	state.CursorLeft()
	renderer.Update(widget)
	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("cursor move should stay paint-only, got %q", renderer.lastFrameMode)
	}

	state.Insert("c")
	renderer.Update(widget)
	if renderer.lastFrameMode != rendererFrameFull {
		t.Fatalf("content width growth should force full render, got %q", renderer.lastFrameMode)
	}
	if renderer.lastLayoutCount == 0 {
		t.Fatal("expected layout work after intrinsic width changed")
	}
}

func TestRenderer_CleanUpdateDoesNotFallbackToFullRender(t *testing.T) {
	renderer := newTestRenderer(uv.NewBuffer(10, 1), 10, 1)
	root := Text{Content: "steady"}

	renderer.Update(root)
	fullCount := renderer.fullRenderCount
	partialCount := renderer.partialRenderCount

	renderer.Update(root)

	if renderer.fullRenderCount != fullCount {
		t.Fatalf("clean update should not trigger a full render: got %d, want %d", renderer.fullRenderCount, fullCount)
	}
	if renderer.partialRenderCount != partialCount {
		t.Fatalf("clean update should not trigger a partial render: got %d, want %d", renderer.partialRenderCount, partialCount)
	}
}

func TestRenderer_TextInputBlurRepaintsVirtualCursor(t *testing.T) {
	screen := newTrackingScreen(20, 1)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(
		screen,
		20,
		1,
		focusManager,
		focusedSignal,
		NewAnySignal[Widget](nil),
	)

	state := NewTextInputState("hello")
	widget := TextInput{ID: "input", State: state, Width: Cells(10)}

	focusManager.focusedID = "input"
	focusedSignal.Set(widget)
	renderer.Update(widget)

	screen.resetTouched()
	focusManager.focusedID = ""
	focusedSignal.Set(nil)
	renderer.Update(widget)

	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint on blur, got %q", renderer.lastFrameMode)
	}
	if len(screen.touched) == 0 {
		t.Fatal("expected blur to repaint the input and clear the virtual cursor")
	}
}
