package terma

// Select returns fn applied to the signal's current value. Read during a
// tracked phase, it subscribes the widget only to changes in that result:
// after a Set, the widget is notified only if fn gives a different answer.
//
// This lets many widgets watch one signal while only the affected ones update.
// For example, each list row can ask whether it is the active row, so moving
// the cursor notifies just the old and new rows:
//
//	active := Select(state.CursorIndex, func(i int) bool { return i == row })
//
// fn runs on every Set of the signal, so it must be cheap and pure, and must
// not read other signals.
func Select[T, R comparable](s Signal[T], fn func(T) R) R {
	read := currentReadSubscription()

	s.core.mu.Lock()
	defer s.core.mu.Unlock()
	// Evaluate and register under the lock so no Set can slip between them.
	result := fn(s.core.value)
	if read.node != nil && read.phase != readPhaseNone {
		s.core.selectors.add(read.node, read.phase, func(value T) bool { return fn(value) != result })
		read.node.trackDependency(s.core, read.phase)
	}
	return result
}

// SelectAny is Select for an AnySignal.
func SelectAny[T any, R comparable](s AnySignal[T], fn func(T) R) R {
	read := currentReadSubscription()

	s.core.mu.Lock()
	defer s.core.mu.Unlock()
	result := fn(s.core.value)
	if read.node != nil && read.phase != readPhaseNone {
		s.core.selectors.add(read.node, read.phase, func(value T) bool { return fn(value) != result })
		read.node.trackDependency(s.core, read.phase)
	}
	return result
}

// selectorSet holds a signal's projected subscriptions.
type selectorSet[T any] struct {
	byNode map[*widgetNode][]signalSelector[T]
}

type signalSelector[T any] struct {
	phase   dependencyMask
	changed func(T) bool // Reports whether the projection differs from what the widget last read.
}

func (s *selectorSet[T]) add(node *widgetNode, phase dependencyMask, changed func(T) bool) {
	if s.byNode == nil {
		s.byNode = make(map[*widgetNode][]signalSelector[T])
	}
	s.byNode[node] = append(s.byNode[node], signalSelector[T]{phase: phase, changed: changed})
}

// remove drops the node's selectors for the given phases.
func (s *selectorSet[T]) remove(node *widgetNode, mask dependencyMask) {
	selectors, ok := s.byNode[node]
	if !ok {
		return
	}
	kept := selectors[:0]
	for _, selector := range selectors {
		if selector.phase&mask == 0 {
			kept = append(kept, selector)
		}
	}
	if len(kept) == 0 {
		delete(s.byNode, node)
		return
	}
	s.byNode[node] = kept
}

type listenerEntry struct {
	node *widgetNode
	mask dependencyMask
}

type nodeSelectors[T any] struct {
	node      *widgetNode
	selectors []signalSelector[T]
}

// signalNotification is captured while holding a signal's lock and delivered
// after releasing it: marking nodes dirty and running selectors must not hold it.
type signalNotification[T any] struct {
	value     T
	listeners []listenerEntry
	selectors []nodeSelectors[T]
}

func captureNotification[T any](value T, listeners map[*widgetNode]dependencyMask, selectors *selectorSet[T]) signalNotification[T] {
	n := signalNotification[T]{value: value, listeners: make([]listenerEntry, 0, len(listeners))}
	for node, mask := range listeners {
		n.listeners = append(n.listeners, listenerEntry{node: node, mask: mask})
	}
	if len(selectors.byNode) > 0 {
		n.selectors = make([]nodeSelectors[T], 0, len(selectors.byNode))
		selectorCount := 0
		for _, list := range selectors.byNode {
			selectorCount += len(list)
		}
		// Snapshot into one buffer instead of allocating a slice for every row.
		// The copy is still necessary: subscriptions can be compacted or replaced
		// after the signal lock is released and before delivery finishes.
		snapshot := make([]signalSelector[T], selectorCount)
		offset := 0
		for node, list := range selectors.byNode {
			end := offset + copy(snapshot[offset:], list)
			n.selectors = append(n.selectors, nodeSelectors[T]{node: node, selectors: snapshot[offset:end:end]})
			offset = end
		}
	}
	return n
}

func (n signalNotification[T]) deliver() {
	for _, listener := range n.listeners {
		listener.node.markSignalDirty(listener.mask)
	}
	for _, entry := range n.selectors {
		var mask dependencyMask
		for _, selector := range entry.selectors {
			if selector.changed(n.value) {
				mask |= selector.phase
			}
		}
		entry.node.markSignalDirty(mask)
	}
}
