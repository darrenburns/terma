package terma

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPageDownAfterDirectOffsetWriteUsesNewPosition(t *testing.T) {
	advance := installScrollAnimationClock(t)
	scroll := newMeasuredScrollState(10, 100)
	scroll.PageDown()
	advance(testFrame)

	// A handler can write the public signal and request another page before
	// the animation controller gets its next tick.
	scroll.Offset.Set(40)
	require.True(t, scroll.PageDown())
	require.Equal(t, 40, scroll.GetOffset(), "paging must not jump immediately")
	require.Equal(t, 50, scroll.scrollTarget(), "the old glide target is superseded")
	glide(t, scroll, advance)
	require.Equal(t, 50, scroll.GetOffset())
}

func TestScrollToViewAfterDirectOffsetWriteRevealsRegion(t *testing.T) {
	advance := installScrollAnimationClock(t)
	scroll := newMeasuredScrollState(10, 100)
	scroll.PageDown()
	advance(testFrame)

	// Row 15 lies within the canceled glide's destination viewport, but is
	// outside the viewport established by this direct offset write.
	scroll.Offset.Set(40)
	scroll.ScrollToView(15, 1)
	require.Nil(t, scroll.animation, "the direct write canceled the glide")
	require.Equal(t, 15, scroll.GetOffset(), "an ordinary reveal moves immediately")
	glide(t, scroll, advance)
	require.Equal(t, 15, scroll.GetOffset(), "the requested row must become visible")
}
