package terma

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFrameSchedulerDisplayUsesFrameBudget(t *testing.T) {
	interval := time.Second / 60
	for _, delay := range []time.Duration{0, 5 * time.Millisecond, 10 * time.Millisecond, 25 * time.Millisecond} {
		t.Run(delay.String(), func(t *testing.T) {
			now := time.Unix(1, 0)
			starts := []time.Time{}
			scheduler := newFrameScheduler(interval, func() {
				starts = append(starts, now)
				now = now.Add(delay)
			})
			scheduler.now = func() time.Time { return now }
			defer scheduler.stopTimer()
			require.False(t, scheduler.request())
			wantSpacing := interval
			if delay >= interval {
				wantSpacing = 2 * interval
			}
			require.Equal(t, starts[0].Add(wantSpacing), scheduler.nextFrame)
			// An immediate request after display waits only for the budget remaining.
			require.False(t, scheduler.request())
			require.True(t, scheduler.pending)
			require.True(t, scheduler.request())
			require.Len(t, starts, 1)
			require.Equal(t, wantSpacing-delay, scheduler.nextFrame.Sub(now))
			now = scheduler.nextFrame
			scheduler.renderNow()
			require.Equal(t, wantSpacing, starts[1].Sub(starts[0]))
			require.False(t, scheduler.pending)
			require.Nil(t, scheduler.timerCh)
		})
	}
}

func TestFrameSchedulerInputBarrier(t *testing.T) {
	for _, input := range []string{"key", "paste"} {
		t.Run(input, func(t *testing.T) {
			now := time.Unix(1, 0)
			model, drawn := "initial", ""
			scheduler := newFrameScheduler(time.Second/60, func() { drawn = model })
			scheduler.now = func() time.Time { return now }
			defer scheduler.stopTimer()
			scheduler.renderNow()
			model = "dialog open"
			scheduler.request()
			require.Equal(t, "initial", drawn)
			// app.go forces the pending frame before routing either input kind.
			if scheduler.pending {
				scheduler.renderNow()
			}
			require.Equal(t, "dialog open", drawn)
			require.False(t, scheduler.pending)
			require.Nil(t, scheduler.timerCh)
			// Another update can arm the same timer after an input-forced frame.
			model = "next"
			scheduler.request()
			require.True(t, scheduler.pending)
			require.NotNil(t, scheduler.timerCh)
			scheduler.stopTimer()
			require.Nil(t, scheduler.timerCh)
		})
	}
}

func TestFrameSchedulerIdleAndLargeOverrun(t *testing.T) {
	now := time.Unix(1, 0)
	frames := 0
	scheduler := newFrameScheduler(10*time.Millisecond, func() {
		frames++
		now = now.Add(125 * time.Millisecond)
	})
	scheduler.now = func() time.Time { return now }
	defer scheduler.stopTimer()
	scheduler.request()
	require.Equal(t, 5*time.Millisecond, scheduler.nextFrame.Sub(now))
	scheduler.request()
	require.Equal(t, 1, frames)
	require.True(t, scheduler.pending)
	now = now.Add(time.Hour)
	scheduler.request()
	require.Equal(t, 2, frames)
	require.False(t, scheduler.pending)
	require.Nil(t, scheduler.timerCh)
}
