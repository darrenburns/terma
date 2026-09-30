package terma

// imageWorker prepares image data (PNG encoding for kitty, palettes for
// sixel) off the event loop in an app, where a large image would otherwise
// hold up input and frames for as long as it takes. Elsewhere (tests,
// headless rendering) the work is done in place, so frames stay
// deterministic; the half-block preview shows until an image is ready.
type imageWorker struct{ async bool }

// runImageWork calls work, then done with its result. When the worker is
// async, work runs on its own goroutine and done on the event loop before the
// next frame (which it schedules); done isn't called if the app has stopped.
func runImageWork[T any](w imageWorker, work func() T, done func(T)) {
	if !w.async {
		done(work())
		return
	}
	go func() {
		result := work()
		dispatchIfRunning(func() { done(result) })
	}()
}
