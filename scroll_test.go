package terma

import (
	"fmt"
	"math"
	"testing"
)

func TestNewScrollState_IsPinnedTrue(t *testing.T) {
	s := NewScrollState()

	if !s.isPinned {
		t.Error("expected new ScrollState to have isPinned=true")
	}
}

func TestScrollState_PinToBottom_Disabled_ByDefault(t *testing.T) {
	s := NewScrollState()

	if s.PinToBottom {
		t.Error("expected PinToBottom to be false by default")
	}
}

func TestScrollState_IsAtBottom_EmptyContent(t *testing.T) {
	s := NewScrollState()
	s.viewportHeight = 10
	s.contentHeight = 0

	if !s.IsAtBottom() {
		t.Error("expected IsAtBottom()=true with empty content")
	}
}

func TestScrollState_IsAtBottom_ContentSmallerThanViewport(t *testing.T) {
	s := NewScrollState()
	s.viewportHeight = 10
	s.contentHeight = 5
	s.Offset.Set(0)

	if !s.IsAtBottom() {
		t.Error("expected IsAtBottom()=true when content is smaller than viewport")
	}
}

func TestScrollState_IsAtBottom_AtMaxOffset(t *testing.T) {
	s := NewScrollState()
	s.viewportHeight = 10
	s.contentHeight = 20
	s.Offset.Set(10) // maxOffset = 20 - 10 = 10

	if !s.IsAtBottom() {
		t.Error("expected IsAtBottom()=true at max offset")
	}
}

func TestScrollState_IsAtBottom_NotAtBottom(t *testing.T) {
	s := NewScrollState()
	s.viewportHeight = 10
	s.contentHeight = 20
	s.Offset.Set(5)

	if s.IsAtBottom() {
		t.Error("expected IsAtBottom()=false when not at bottom")
	}
}

func TestScrollState_IsPinned_WhenDisabled(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = false

	if s.IsPinned() {
		t.Error("expected IsPinned()=false when PinToBottom is disabled")
	}
}

func TestScrollState_IsPinned_WhenEnabledAndPinned(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = true

	if !s.IsPinned() {
		t.Error("expected IsPinned()=true when PinToBottom is enabled and pinned")
	}
}

func TestScrollState_IsPinned_WhenEnabledButUnpinned(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = false

	if s.IsPinned() {
		t.Error("expected IsPinned()=false when PinToBottom is enabled but unpinned")
	}
}

func TestScrollState_ScrollUp_BreaksPin(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = true
	s.viewportHeight = 10
	s.contentHeight = 20
	s.Offset.Set(10) // At bottom

	s.ScrollUp(1)

	if s.isPinned {
		t.Error("expected pin to be broken after ScrollUp")
	}
}

func TestScrollState_ScrollUp_DoesNotBreakPinWhenDisabled(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = false
	s.isPinned = true
	s.viewportHeight = 10
	s.contentHeight = 20
	s.Offset.Set(10)

	s.ScrollUp(1)

	// isPinned should remain true (pin feature is disabled)
	if !s.isPinned {
		t.Error("expected isPinned to remain true when PinToBottom is disabled")
	}
}

func TestScrollState_ScrollDown_ReengagesPin(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = false
	s.viewportHeight = 10
	s.contentHeight = 20
	s.Offset.Set(9) // One line above bottom

	s.ScrollDown(1) // Should reach bottom (offset 10)

	if !s.isPinned {
		t.Error("expected pin to be re-engaged after scrolling to bottom")
	}
}

func TestScrollState_ScrollDown_DoesNotPinWhenNotAtBottom(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = false
	s.viewportHeight = 10
	s.contentHeight = 20
	s.Offset.Set(0)

	s.ScrollDown(1) // offset becomes 1, not at bottom

	if s.isPinned {
		t.Error("expected pin to remain broken when not at bottom")
	}
}

func TestScrollState_ScrollToBottom_SetsOffset(t *testing.T) {
	s := NewScrollState()
	s.viewportHeight = 10
	s.contentHeight = 20
	s.Offset.Set(0)

	s.ScrollToBottom()

	if s.Offset.Peek() != 10 {
		t.Errorf("expected offset=10, got %d", s.Offset.Peek())
	}
}

func TestScrollState_ScrollToBottom_ReengagesPin(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = false
	s.viewportHeight = 10
	s.contentHeight = 20

	s.ScrollToBottom()

	if !s.isPinned {
		t.Error("expected pin to be re-engaged after ScrollToBottom")
	}
}

func TestScrollState_UpdateLayout_AutoScrollsWhenPinned(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = true
	s.viewportHeight = 10
	s.contentHeight = 20
	s.Offset.Set(10) // At bottom (maxOffset)

	// Content grows by 5 lines
	s.updateLayout(10, 25)

	// Should auto-scroll to new bottom (25 - 10 = 15)
	if s.Offset.Peek() != 15 {
		t.Errorf("expected offset=15 after content growth, got %d", s.Offset.Peek())
	}
}

func TestScrollState_UpdateLayout_DoesNotAutoScrollWhenUnpinned(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = false // Pin broken
	s.viewportHeight = 10
	s.contentHeight = 20
	s.Offset.Set(5)

	// Content grows by 5 lines
	s.updateLayout(10, 25)

	// Should NOT auto-scroll - stay at offset 5
	if s.Offset.Peek() != 5 {
		t.Errorf("expected offset to stay at 5, got %d", s.Offset.Peek())
	}
}

func TestScrollState_UpdateLayout_DoesNotAutoScrollWhenDisabled(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = false
	s.isPinned = true
	s.viewportHeight = 10
	s.contentHeight = 20
	s.Offset.Set(10)

	// Content grows by 5 lines
	s.updateLayout(10, 25)

	// Should NOT auto-scroll
	if s.Offset.Peek() != 10 {
		t.Errorf("expected offset to stay at 10, got %d", s.Offset.Peek())
	}
}

func TestScrollState_UpdateLayout_DoesNotAutoScrollOnFirstRender(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = true
	// contentHeight starts at 0 (first render)

	s.updateLayout(10, 20)

	// Should NOT auto-scroll on initial render (oldContentHeight = 0)
	if s.Offset.Peek() != 0 {
		t.Errorf("expected offset to stay at 0 on first render, got %d", s.Offset.Peek())
	}
}

func TestScrollState_UpdateLayout_StaysAtBottomWhenContentShrinks(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = true
	s.viewportHeight = 10
	s.contentHeight = 30
	s.Offset.Set(20)

	// Content shrinks
	s.updateLayout(10, 20)

	// Pinned content stays at the (new) bottom, as the layout shows it.
	if s.Offset.Peek() != 10 {
		t.Errorf("expected offset=10 at the new bottom, got %d", s.Offset.Peek())
	}
}

func TestScrollState_UpdateLayout_DoesNotAutoScrollWhenNotAtBottom(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = true // Engaged initially, but content started out taller than the viewport
	s.viewportHeight = 10
	s.contentHeight = 20

	s.updateLayout(10, 25)

	if s.Offset.Peek() != 0 {
		t.Errorf("expected offset to stay at 0 until scrolled to the bottom, got %d", s.Offset.Peek())
	}
}

// Integration test: simulates chat-like behavior
func TestScrollState_ChatLikeBehavior(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true

	// Initial state: empty content
	s.updateLayout(10, 0)
	if !s.IsPinned() {
		t.Error("expected to be pinned initially")
	}

	// First message arrives (content = 5 lines)
	s.viewportHeight = 10
	s.contentHeight = 0 // Will be set by updateLayout
	s.updateLayout(10, 5)
	// No auto-scroll on first render (oldContentHeight was 0)
	if s.Offset.Peek() != 0 {
		t.Errorf("expected offset=0 after first message, got %d", s.Offset.Peek())
	}

	// More messages arrive (content = 15 lines)
	s.updateLayout(10, 15)
	// Should auto-scroll to bottom (15 - 10 = 5)
	if s.Offset.Peek() != 5 {
		t.Errorf("expected offset=5 after content growth, got %d", s.Offset.Peek())
	}

	// User scrolls up to read history
	s.ScrollUp(3)
	if s.isPinned {
		t.Error("expected pin to break after scroll up")
	}
	if s.Offset.Peek() != 2 {
		t.Errorf("expected offset=2 after scroll up, got %d", s.Offset.Peek())
	}

	// New message arrives while scrolled up
	s.updateLayout(10, 20)
	// Should NOT auto-scroll (pin is broken)
	if s.Offset.Peek() != 2 {
		t.Errorf("expected offset to stay at 2, got %d", s.Offset.Peek())
	}

	// User scrolls back to bottom
	s.ScrollDown(8) // offset 2 + 8 = 10, which is max for viewport=10, content=20
	if !s.isPinned {
		t.Error("expected pin to re-engage at bottom")
	}

	// New message arrives
	s.updateLayout(10, 25)
	// Should auto-scroll again
	if s.Offset.Peek() != 15 {
		t.Errorf("expected offset=15 after resuming pin, got %d", s.Offset.Peek())
	}
}

func TestScrollState_SetOffsetX_ClampsToBounds(t *testing.T) {
	s := NewScrollState()
	s.updateHorizontalLayout(10, 30) // maxOffsetX = 20

	s.SetOffsetX(-5)
	if s.GetOffsetX() != 0 {
		t.Errorf("expected horizontal offset=0, got %d", s.GetOffsetX())
	}

	s.SetOffsetX(100)
	if s.GetOffsetX() != 20 {
		t.Errorf("expected horizontal offset=20, got %d", s.GetOffsetX())
	}
}

func TestScrollState_ScrollLeftRight(t *testing.T) {
	s := NewScrollState()
	s.updateHorizontalLayout(10, 30) // maxOffsetX = 20

	if handled := s.ScrollRight(3); !handled {
		t.Error("expected ScrollRight to be handled")
	}
	if s.GetOffsetX() != 3 {
		t.Errorf("expected horizontal offset=3, got %d", s.GetOffsetX())
	}

	if handled := s.ScrollLeft(2); !handled {
		t.Error("expected ScrollLeft to be handled")
	}
	if s.GetOffsetX() != 1 {
		t.Errorf("expected horizontal offset=1, got %d", s.GetOffsetX())
	}
}

func TestScrollState_ScrollLeftRight_Callbacks(t *testing.T) {
	s := NewScrollState()
	leftCalls := 0
	rightCalls := 0
	s.OnScrollLeft = func(cols int) bool {
		leftCalls += cols
		return true
	}
	s.OnScrollRight = func(cols int) bool {
		rightCalls += cols
		return true
	}

	if handled := s.ScrollRight(4); !handled {
		t.Error("expected ScrollRight callback to handle")
	}
	if handled := s.ScrollLeft(3); !handled {
		t.Error("expected ScrollLeft callback to handle")
	}

	if rightCalls != 4 {
		t.Errorf("expected right callback count=4, got %d", rightCalls)
	}
	if leftCalls != 3 {
		t.Errorf("expected left callback count=3, got %d", leftCalls)
	}
	if s.GetOffsetX() != 0 {
		t.Errorf("expected horizontal offset unchanged at 0, got %d", s.GetOffsetX())
	}
}

func TestScrollable_DragScrollbarThumb(t *testing.T) {
	state := NewScrollState()
	state.updateLayout(5, 20) // max offset = 15
	state.layoutCache = scrollableLayoutCache{
		valid:         true,
		contentWidth:  10,
		contentHeight: 5,
		scrollableY:   true,
	}

	scrollable := Scrollable{State: state}

	scrollable.OnMouseDown(MouseEvent{LocalX: 9, LocalY: 0})
	if !state.scrollbarDragging {
		t.Fatal("expected dragging to start when mouse is down on thumb")
	}

	scrollable.OnMouseMove(MouseEvent{LocalX: 9, LocalY: 4})
	if state.GetOffset() == 0 {
		t.Fatalf("expected scroll offset to change while dragging, got %d", state.GetOffset())
	}

	scrollable.OnMouseUp(MouseEvent{})
	if state.scrollbarDragging {
		t.Fatal("expected dragging to stop on mouse up")
	}
}

func TestScrollable_DragScrollbar_IgnoresClicksOutsideScrollbar(t *testing.T) {
	state := NewScrollState()
	state.updateLayout(5, 20)
	state.layoutCache = scrollableLayoutCache{
		valid:         true,
		contentWidth:  10,
		contentHeight: 5,
		scrollableY:   true,
	}

	scrollable := Scrollable{State: state}

	scrollable.OnMouseDown(MouseEvent{LocalX: 8, LocalY: 0}) // Not on scrollbar column
	if state.scrollbarDragging {
		t.Fatal("expected dragging to remain disabled when clicking outside scrollbar")
	}

	scrollable.OnMouseMove(MouseEvent{LocalX: 8, LocalY: 4})
	if state.GetOffset() != 0 {
		t.Fatalf("expected offset to remain unchanged, got %d", state.GetOffset())
	}
}

func TestScrollable_DragScrollbar_UsesContentOffsets(t *testing.T) {
	state := NewScrollState()
	state.updateLayout(5, 20)
	state.layoutCache = scrollableLayoutCache{
		valid:          true,
		contentWidth:   10,
		contentHeight:  5,
		contentOffsetX: 2,
		contentOffsetY: 1,
		scrollableY:    true,
	}

	scrollable := Scrollable{State: state}

	scrollable.OnMouseDown(MouseEvent{LocalX: 9, LocalY: 1}) // Inside content but not scrollbar column
	if state.scrollbarDragging {
		t.Fatal("expected dragging to remain disabled when click is not on scrollbar column")
	}

	scrollable.OnMouseDown(MouseEvent{LocalX: 11, LocalY: 1}) // contentOffsetX + (contentWidth-1)
	if !state.scrollbarDragging {
		t.Fatal("expected dragging to start when clicking on offset scrollbar column")
	}
}

// With a pixel-precise pointer, the thumb stays under the point where it was
// grabbed, to the eighth of a cell, while content moves by whole lines.
func TestScrollable_DragScrollbar_FollowsSubCellPointer(t *testing.T) {
	state := NewScrollState()
	state.updateLayout(10, 40) // max offset = 30
	state.layoutCache = scrollableLayoutCache{
		valid:         true,
		contentWidth:  10,
		contentHeight: 10,
		scrollableY:   true,
	}
	scrollable := Scrollable{State: state}

	// The thumb is 2.5 cells long and can travel 7.5 cells. Grab it 1.25
	// cells down.
	scrollable.OnMouseDown(MouseEvent{LocalX: 9, LocalY: 1, SubCellY: 0.25})
	if !state.scrollbarDragging {
		t.Fatal("expected dragging to start on the thumb")
	}

	for step := 1; step <= 24; step++ {
		pointer := 1.25 + float64(step)/scrollbarSubCellCount
		cell := math.Floor(pointer)
		scrollable.OnMouseMove(MouseEvent{LocalX: 9, LocalY: int(cell), SubCellY: pointer - cell})

		thumb, _ := scrollable.scrollbarThumb(state.layoutCache)
		if thumb.start != step {
			t.Fatalf("step %d: thumb starts %d eighths down, expected it under the pointer at %d", step, thumb.start, step)
		}
		position := state.position.Peek()
		if want := float64(step) / 2; math.Abs(position-want) > 1e-9 {
			t.Fatalf("step %d: expected exact position %.2f, got %.4f", step, want, position)
		}
		if state.GetOffset() != int(math.Round(position)) {
			t.Fatalf("step %d: offset %d is not position %.2f rounded", step, state.GetOffset(), position)
		}
	}
}

func TestScrollState_ThumbPositionFollowsExactPosition(t *testing.T) {
	s := NewScrollState()
	s.updateLayout(10, 40) // max offset = 30

	s.setPosition(3.4)
	if s.GetOffset() != 3 || s.thumbPosition(3) != 3.4 {
		t.Fatalf("expected offset 3 with the thumb at 3.4, got offset %d, thumb %.2f", s.GetOffset(), s.thumbPosition(3))
	}

	s.SetOffset(7)
	if s.thumbPosition(7) != 7 {
		t.Errorf("SetOffset should put the thumb at the whole line, got %.2f", s.thumbPosition(7))
	}

	s.setPosition(5.6)
	s.Offset.Set(12) // Moved without going through ScrollState
	if s.thumbPosition(12) != 12 {
		t.Errorf("a stale exact position should be ignored, got %.2f", s.thumbPosition(12))
	}

	s.setPosition(99)
	if s.GetOffset() != 30 || s.position.Peek() != 30 {
		t.Errorf("expected position clamped to 30, got offset %d, position %.2f", s.GetOffset(), s.position.Peek())
	}
}

// The thumb's length is computed once from the track and content heights, so
// it never changes size as it moves.
func TestScrollbarThumb_KeepsItsLengthAsItMoves(t *testing.T) {
	for _, tc := range []struct{ track, content int }{{10, 23}, {6, 40}, {20, 21}, {5, 1000}, {3, 4}, {1, 50}} {
		maxScroll := tc.content - tc.track
		first := newScrollbarThumb(0, maxScroll, tc.track, tc.content)
		if first.start != 0 || first.length < scrollbarSubCellCount {
			t.Fatalf("track %d, content %d: bad thumb at top: %+v", tc.track, tc.content, first)
		}
		previous := 0
		for step := 0; step <= maxScroll*scrollbarSubCellCount; step++ {
			thumb := newScrollbarThumb(float64(step)/scrollbarSubCellCount, maxScroll, tc.track, tc.content)
			if thumb.length != first.length {
				t.Fatalf("track %d, content %d: length %d at step %d, expected %d", tc.track, tc.content, thumb.length, step, first.length)
			}
			if thumb.start < previous {
				t.Fatalf("track %d, content %d: thumb moved backwards at step %d", tc.track, tc.content, step)
			}
			previous = thumb.start
		}
		last := newScrollbarThumb(float64(maxScroll), maxScroll, tc.track, tc.content)
		if last.start+last.length != tc.track*scrollbarSubCellCount {
			t.Errorf("track %d, content %d: thumb should end at the bottom, got %+v", tc.track, tc.content, last)
		}
	}
}

type scrollbarRows struct{ count int }

func (r scrollbarRows) Build(BuildContext) Widget {
	rows := make([]Widget, r.count)
	for i := range rows {
		rows[i] = Text{Content: fmt.Sprintf("row %d", i)}
	}
	return Column{Children: rows}
}

func TestSnapshot_Scrollbar_ThumbPositions(t *testing.T) {
	state := NewScrollState()
	widget := Scrollable{
		State:  state,
		Height: Cells(6),
		Style:  Style{Border: RoundedBorder(Hex("#6e6a86")), Padding: EdgeInsetsXY(1, 0)},
		Child:  scrollbarRows{count: 10},
	}
	RenderToBuffer(widget, 14, 8) // Lays out the content, so positions can be set.

	for _, tc := range []struct {
		name     string
		position float64
		desc     string
	}{
		{"top", 0, "Thumb at the top of the track"},
		{"third_line", 1.0 / 3, "Scrolled a third of a line: content unmoved, thumb moved by its share"},
		{"between_lines", 2.6, "Content at line 3, thumb between the positions for lines 2 and 3"},
		{"bottom", 4, "Thumb at the bottom, the same length as at the top"},
	} {
		state.setPosition(tc.position)
		AssertSnapshotNamed(t, "scrollbar_thumb_"+tc.name, widget, 14, 8, tc.desc)
	}
}
