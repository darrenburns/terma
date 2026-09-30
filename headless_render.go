package terma

const maxHeadlessRenderFrames = 16

// renderHeadlessFrames supplies a frame boundary for standalone renderers. An
// app already owns its boundary, so its queue must be left to the event loop.
func (r *Renderer) renderHeadlessFrames(root Widget, incremental bool) (focusables []FocusableEntry, layoutWidth, layoutHeight int) {
	queue, release := beginHeadlessDispatch()
	if queue == nil {
		if incremental {
			focusables, layoutWidth, layoutHeight = r.updateInternal(root)
		} else {
			focusables, layoutWidth, layoutHeight = r.renderFull(root)
		}
		if r.finishImages(root) {
			focusables, layoutWidth, layoutHeight = r.lastFocusables, r.lastLayoutWidth, r.lastLayoutHeight
		}
		return focusables, layoutWidth, layoutHeight
	}
	defer release()

	for frame := 0; frame < maxHeadlessRenderFrames; frame++ {
		if incremental && frame == 0 {
			focusables, layoutWidth, layoutHeight = r.updateInternal(root)
		} else {
			if frame > 0 {
				// A bare headless buffer needs the same clear as a terminal full
				// frame; otherwise shortened or removed content leaves stale cells.
				r.clearRect(Rect{Width: r.width, Height: r.height})
			}
			// Dispatch also supports plain Go fields, so a callback may have
			// changed Build inputs without recording reactive invalidation.
			focusables, layoutWidth, layoutHeight = r.renderFull(root)
		}
		// The first-image switch performs another full paint. Keep its layout
		// callbacks inside this frame boundary and return its final geometry.
		if r.finishImages(root) {
			focusables, layoutWidth, layoutHeight = r.lastFocusables, r.lastLayoutWidth, r.lastLayoutHeight
		}
		pending := queue.takePending()
		if len(pending) == 0 {
			return focusables, layoutWidth, layoutHeight
		}
		if frame == maxHeadlessRenderFrames-1 {
			panic("terma: headless Dispatch did not settle within 16 frames")
		}
		for _, fn := range pending {
			fn()
		}
	}
	return focusables, layoutWidth, layoutHeight
}
