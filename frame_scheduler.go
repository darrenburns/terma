package terma

import "time"

// frameScheduler coalesces visual updates on the event loop. Input may force
// renderNow before the deadline so the next key/paste sees the current tree.
type frameScheduler struct {
	interval  time.Duration
	nextFrame time.Time
	pending   bool
	timer     *time.Timer
	timerCh   <-chan time.Time
	now       func() time.Time
	display   func()
}

func newFrameScheduler(interval time.Duration, display func()) *frameScheduler {
	return &frameScheduler{interval: interval, now: time.Now, display: display}
}

func (s *frameScheduler) stopTimer() {
	if s.timer != nil && !s.timer.Stop() {
		select {
		case <-s.timer.C:
		default:
		}
	}
	s.timerCh = nil
}

func (s *frameScheduler) renderNow() {
	s.stopTimer()
	s.pending = false
	s.nextFrame = s.now().Add(s.interval)
	s.display()
	// Rendering consumes the frame budget. Skip deadlines already missed so
	// an overrun cannot cause a burst of catch-up frames.
	if now := s.now(); !now.Before(s.nextFrame) {
		missed := now.Sub(s.nextFrame)/s.interval + 1
		s.nextFrame = s.nextFrame.Add(missed * s.interval)
	}
}

// request reports whether this request was coalesced with a pending frame.
func (s *frameScheduler) request() bool {
	now := s.now()
	if s.nextFrame.IsZero() || !now.Before(s.nextFrame) {
		s.renderNow()
		return false
	}
	if s.pending {
		return true
	}
	s.pending = true
	wait := s.nextFrame.Sub(now)
	if s.timer == nil {
		s.timer = time.NewTimer(wait)
	} else {
		s.stopTimer()
		s.timer.Reset(wait)
	}
	s.timerCh = s.timer.C
	return false
}
