package terma

import (
	"fmt"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

// installScrollAnimationClock runs animations on a controller with a fake
// clock, as a running app would. The returned function moves the clock on and
// ticks the controller.
func installScrollAnimationClock(t *testing.T) func(time.Duration) {
	t.Helper()
	previous := currentController
	controller := NewAnimationController(60)
	now := time.Unix(1, 0)
	controller.now = func() time.Time { return now }
	currentController = controller
	t.Cleanup(func() { controller.Stop(); currentController = previous })
	return func(d time.Duration) {
		now = now.Add(d)
		controller.Update()
	}
}

const testFrame = 16 * time.Millisecond

// newMeasuredScrollState returns a scroll state as a Scrollable leaves it
// after layout.
func newMeasuredScrollState(viewportHeight, contentHeight int) *ScrollState {
	s := NewScrollState()
	s.updateLayout(viewportHeight, contentHeight)
	return s
}

// glide ticks frames until the scroll animation ends, returning the offset
// drawn on each frame.
func glide(t *testing.T, s *ScrollState, advance func(time.Duration)) []int {
	t.Helper()
	var offsets []int
	for i := 0; s.animation != nil; i++ {
		require.Less(t, i, 100, "the animation never finished")
		advance(testFrame)
		offsets = append(offsets, s.GetOffset())
	}
	return offsets
}

func requireGlide(t *testing.T, offsets []int, from, to int) {
	t.Helper()
	require.GreaterOrEqual(t, len(offsets), 3, "the scroll takes several frames: %v", offsets)
	require.Equal(t, to, offsets[len(offsets)-1], "the glide ends at the target: %v", offsets)
	previous := from
	for _, offset := range offsets {
		if to > from {
			require.GreaterOrEqual(t, offset, previous, "the glide never goes backwards: %v", offsets)
		} else {
			require.LessOrEqual(t, offset, previous, "the glide never goes backwards: %v", offsets)
		}
		previous = offset
	}
	require.NotEqual(t, to, offsets[0], "the first frame is part-way: %v", offsets)
}

func TestPageDownJumpsWithoutAnimationController(t *testing.T) {
	s := newMeasuredScrollState(10, 100)
	require.True(t, s.PageDown())
	require.Equal(t, 10, s.GetOffset())
	require.Nil(t, s.animation)
	require.True(t, s.PageUp())
	require.Equal(t, 0, s.GetOffset())
}

func TestPageDownGlidesToNextPage(t *testing.T) {
	advance := installScrollAnimationClock(t)
	s := newMeasuredScrollState(10, 100)

	require.True(t, s.PageDown())
	require.Equal(t, 0, s.GetOffset(), "the viewport starts moving on the next frame")
	require.Equal(t, 10, s.scrollTarget())

	advance(testFrame)
	position := s.position.Peek()
	require.Greater(t, position, 0.0)
	require.Less(t, position, 10.0)
	require.Equal(t, int(position+0.5), s.GetOffset(), "content is drawn at the nearest whole line")

	offsets := append([]int{s.GetOffset()}, glide(t, s, advance)...)
	requireGlide(t, offsets, 0, 10)
	require.Equal(t, 10.0, s.position.Peek())

	require.True(t, s.PageUp())
	requireGlide(t, glide(t, s, advance), 10, 0)
}

func TestRepeatedPageDownAddsUpWithoutJumping(t *testing.T) {
	advance := installScrollAnimationClock(t)
	s := newMeasuredScrollState(10, 100)

	s.PageDown()
	advance(testFrame)
	advance(testFrame)
	midway := s.GetOffset()
	require.Greater(t, midway, 0)
	require.Less(t, midway, 10)

	s.PageDown()
	s.PageDown()
	require.Equal(t, 30, s.scrollTarget(), "pages add up from where the glide is heading")
	require.Equal(t, midway, s.GetOffset(), "retargeting doesn't move the viewport")
	requireGlide(t, glide(t, s, advance), midway, 30)
}

func TestPageDownStopsAtTheEnd(t *testing.T) {
	advance := installScrollAnimationClock(t)
	s := newMeasuredScrollState(10, 25)

	require.True(t, s.PageDown())
	require.True(t, s.PageDown())
	require.Equal(t, 15, s.scrollTarget())
	require.False(t, s.PageDown(), "already heading for the end")
	glide(t, s, advance)
	require.Equal(t, 15, s.GetOffset())
	require.False(t, s.PageDown())
	require.Nil(t, s.animation)
}

func TestOtherScrollingInterruptsPageGlide(t *testing.T) {
	advance := installScrollAnimationClock(t)

	t.Run("SetOffset", func(t *testing.T) {
		s := newMeasuredScrollState(10, 100)
		s.PageDown()
		advance(testFrame)
		s.SetOffset(3)
		require.Nil(t, s.animation)
		advance(time.Second)
		require.Equal(t, 3, s.GetOffset())
	})

	t.Run("wheel", func(t *testing.T) {
		s := newMeasuredScrollState(10, 100)
		s.PageDown()
		advance(testFrame)
		drawn := s.GetOffset()
		s.ScrollDown(1)
		advance(time.Second)
		require.Equal(t, drawn+1, s.GetOffset(), "the wheel scrolls from what is on screen")
	})

	t.Run("Offset signal", func(t *testing.T) {
		s := newMeasuredScrollState(10, 100)
		s.PageDown()
		advance(testFrame)
		s.Offset.Set(40)
		advance(testFrame)
		require.Nil(t, s.animation)
		advance(time.Second)
		require.Equal(t, 40, s.GetOffset())
	})
}

func TestScrollToViewRetargetsPageGlide(t *testing.T) {
	advance := installScrollAnimationClock(t)
	s := newMeasuredScrollState(10, 100)

	s.PageDown()
	advance(testFrame)
	drawn := s.GetOffset()

	// Visible once the glide ends, even if not on screen yet: nothing changes.
	s.ScrollToView(18, 1)
	require.Equal(t, 10, s.scrollTarget())

	// A cursor below where the glide ends extends it.
	s.ScrollToView(24, 1)
	require.Equal(t, 15, s.scrollTarget())
	require.Equal(t, drawn, s.GetOffset())
	requireGlide(t, glide(t, s, advance), drawn, 15)
}

func TestPageDownToBottomRepins(t *testing.T) {
	advance := installScrollAnimationClock(t)
	s := newMeasuredScrollState(10, 20)
	s.PinToBottom = true
	s.ScrollUp(1)
	require.False(t, s.IsPinned())

	s.PageDown()
	glide(t, s, advance)
	require.True(t, s.IsPinned())
	require.True(t, s.pinnedToEnd())
}

func TestFocusedScrollablePageKeysGlide(t *testing.T) {
	advance := installScrollAnimationClock(t)
	s := newMeasuredScrollState(10, 100)
	scrollable := Scrollable{State: s, Focusable: true}

	require.True(t, scrollable.OnKey(makeKeyEvent(uv.KeyPgDown, 0)))
	requireGlide(t, glide(t, s, advance), 0, 10)
	require.True(t, scrollable.OnKey(makeKeyEvent('d', uv.ModCtrl)))
	requireGlide(t, glide(t, s, advance), 10, 20)
	require.True(t, scrollable.OnKey(makeKeyEvent(uv.KeyPgUp, 0)))
	requireGlide(t, glide(t, s, advance), 20, 10)

	// Home and End still jump: a glide across a long document is a blur.
	require.True(t, scrollable.OnKey(makeKeyEvent(uv.KeyEnd, 0)))
	require.Equal(t, 90, s.GetOffset())
	require.Nil(t, s.animation)
}

type pageGlideListScene struct {
	list   *ListState[string]
	scroll *ScrollState
}

func (s *pageGlideListScene) Build(BuildContext) Widget {
	return Scrollable{ID: "scroll", State: s.scroll, Height: Cells(5), Child: s.widget()}
}

func (s *pageGlideListScene) widget() List[string] {
	return List[string]{ID: "list", State: s.list, ScrollState: s.scroll}
}

// Paging a list moves the cursor at once and glides the viewport after it.
// Every frame of the glide must match a full render.
func TestReactivityListPageDownGlides(t *testing.T) {
	advance := installScrollAnimationClock(t)
	sequence := newReactivitySequence(t, 20, 5, func() *pageGlideListScene {
		items := make([]string, 60)
		for i := range items {
			items[i] = fmt.Sprintf("item %d", i)
		}
		return &pageGlideListScene{list: NewListState(items), scroll: NewScrollState()}
	})
	sequence.frame("Initial", nil)
	sequence.focus("list")
	sequence.frame("Focus list", nil)

	sequence.frame("Page down moves the cursor", func(s *pageGlideListScene) { s.widget().pageDown() })
	actual := sequence.actual.root
	require.Equal(t, 10, actual.list.CursorIndex.Peek())
	require.Equal(t, 0, actual.scroll.GetOffset(), "the viewport hasn't moved yet")

	var offsets []int
	for i := 0; actual.scroll.animation != nil; i++ {
		require.Less(t, i, 100)
		advance(testFrame)
		sequence.frame(fmt.Sprintf("Glide frame %d", i+1), nil)
		offsets = append(offsets, actual.scroll.GetOffset())
	}
	requireGlide(t, offsets, 0, 6)
	require.Contains(t, sequence.actual.renderer.ScreenText(), "item 10")

	sequence.frame("Two quick pages", func(s *pageGlideListScene) {
		s.widget().pageDown()
		s.widget().pageDown()
	})
	advance(testFrame)
	sequence.frame("Cursor down during the glide", func(s *pageGlideListScene) { s.widget().keyCursorDown() })
	require.Equal(t, 31, actual.list.CursorIndex.Peek())
	require.Equal(t, 27, actual.scroll.scrollTarget(), "the glide was extended, not cut short")
	offsets = offsets[:0]
	for i := 0; actual.scroll.animation != nil; i++ {
		require.Less(t, i, 100)
		advance(testFrame)
		sequence.frame(fmt.Sprintf("Second glide frame %d", i+1), nil)
		offsets = append(offsets, actual.scroll.GetOffset())
	}
	require.Equal(t, 27, offsets[len(offsets)-1])
}

func TestTextAreaPageDownGlides(t *testing.T) {
	advance := installScrollAnimationClock(t)
	scroll := NewScrollState()
	lines := make([]string, 40)
	for i := range lines {
		lines[i] = fmt.Sprintf("Line %02d", i)
	}
	state := NewTextAreaState(strings.Join(lines, "\n"))
	state.CursorIndex.Set(0)
	area := TextArea{ID: "area", State: state, ScrollState: scroll}
	scene := newWheelScene(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: area}, 24, 5)

	// The first page moves the cursor to the bottom row; the next scrolls.
	area.cursorPageDown()
	scene.draw()
	require.Equal(t, 0, scroll.GetOffset())
	area.cursorPageDown()
	scene.draw()
	require.Equal(t, 0, scroll.GetOffset(), "the viewport hasn't moved yet")

	var offsets []int
	for i := 0; scroll.animation != nil; i++ {
		require.Less(t, i, 100)
		advance(testFrame)
		scene.draw()
		offsets = append(offsets, scroll.GetOffset())
	}
	requireGlide(t, offsets, 0, 4)
	require.Contains(t, scene.renderer.ScreenText(), "Line 08")
}
