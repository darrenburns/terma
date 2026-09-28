package terma

import (
	"fmt"
	"testing"
)

func selectorNotificationFixture(count int) selectorSet[int] {
	var selectors selectorSet[int]
	for i := 0; i < count; i++ {
		selectors.add(newWidgetNode(nil), readPhasePaint, func(int) bool { return false })
	}
	return selectors
}

func TestSelectorNotificationAllocationCount(t *testing.T) {
	selectors := selectorNotificationFixture(100)
	var notification signalNotification[int]
	allocs := testing.AllocsPerRun(10, func() {
		notification = captureNotification(1, nil, &selectors)
	})
	if len(notification.selectors) != 100 {
		t.Fatalf("captured %d subscribers, want 100", len(notification.selectors))
	}
	if allocs > 3 {
		t.Fatalf("notification allocated %.0f times; selector snapshots must use a shared buffer", allocs)
	}
}

func TestSelectorNotificationSnapshotSurvivesSubscriptionChanges(t *testing.T) {
	node := newWidgetNode(nil)
	var selectors selectorSet[int]
	selectors.add(node, readPhaseBuild, func(int) bool { return true })
	selectors.add(node, readPhasePaint, func(int) bool { return false })
	notification := captureNotification(1, nil, &selectors)
	// Removing the first selector compacts the live backing array. An in-flight
	// notification must keep its original selectors after the lock is released.
	selectors.remove(node, readPhaseBuild)
	selectors.add(node, readPhaseLayout, func(int) bool { return false })
	node.clearDirty()
	notification.deliver()
	if node.dirtyLevel() != DirtyBuild {
		t.Fatalf("captured build selector was lost: dirty level = %v", node.dirtyLevel())
	}
}

func BenchmarkSignalSelectorFanout(b *testing.B) {
	for _, count := range []int{100, 1000} {
		b.Run(fmt.Sprintf("subscribers=%d", count), func(b *testing.B) {
			signal := NewSignal(0)
			signal.core.selectors = selectorNotificationFixture(count)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				signal.Set(1 + i%2)
			}
		})
	}
}
