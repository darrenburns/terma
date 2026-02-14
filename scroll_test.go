package terma

import "testing"

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

func TestScrollState_UpdateLayout_DoesNotAutoScrollWhenContentShrinks(t *testing.T) {
	s := NewScrollState()
	s.PinToBottom = true
	s.isPinned = true
	s.viewportHeight = 10
	s.contentHeight = 30
	s.Offset.Set(20)

	// Content shrinks
	s.updateLayout(10, 20)

	// Should NOT auto-scroll when content shrinks
	// Offset stays at 20 (will be clamped elsewhere)
	if s.Offset.Peek() != 20 {
		t.Errorf("expected offset to stay at 20, got %d", s.Offset.Peek())
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
