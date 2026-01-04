package terma

import "testing"

func TestNewSignal_InitialValue(t *testing.T) {
	s := NewSignal(42)

	if s.value != 42 {
		t.Errorf("expected initial value 42, got %d", s.value)
	}
}

func TestSignal_Get_ReturnsValue(t *testing.T) {
	s := NewSignal("hello")

	if s.Get() != "hello" {
		t.Errorf("expected 'hello', got '%s'", s.Get())
	}
}

func TestSignal_Set_UpdatesValue(t *testing.T) {
	s := NewSignal(10)
	s.Set(20)

	if s.Get() != 20 {
		t.Errorf("expected 20, got %d", s.Get())
	}
}

func TestSignal_Peek_ReturnsValue(t *testing.T) {
	s := NewSignal(100)

	if s.Peek() != 100 {
		t.Errorf("expected Peek() = 100, got %d", s.Peek())
	}
}

func TestSignal_Update_FunctionalUpdate(t *testing.T) {
	s := NewSignal(5)
	s.Update(func(v int) int { return v * 2 })

	if s.Get() != 10 {
		t.Errorf("expected 10 after doubling, got %d", s.Get())
	}
}

func TestSignal_Update_ChainedUpdates(t *testing.T) {
	s := NewSignal(1)
	s.Update(func(v int) int { return v + 1 }) // 2
	s.Update(func(v int) int { return v * 3 }) // 6
	s.Update(func(v int) int { return v - 1 }) // 5

	if s.Get() != 5 {
		t.Errorf("expected 5, got %d", s.Get())
	}
}

func TestSignal_Peek_DoesNotSubscribe(t *testing.T) {
	s := NewSignal(42)

	// Simulate being in a build context
	node := newWidgetNode(nil)
	oldNode := currentBuildingNode
	currentBuildingNode = node
	defer func() { currentBuildingNode = oldNode }()

	// Peek should not subscribe
	_ = s.Peek()

	if len(s.listeners) != 0 {
		t.Errorf("expected no listeners after Peek, got %d", len(s.listeners))
	}
}

func TestSignal_Get_DuringBuild_Subscribes(t *testing.T) {
	s := NewSignal(42)

	// Simulate being in a build context
	node := newWidgetNode(nil)
	oldNode := currentBuildingNode
	currentBuildingNode = node
	defer func() { currentBuildingNode = oldNode }()

	// Get should subscribe
	_ = s.Get()

	if len(s.listeners) != 1 {
		t.Errorf("expected 1 listener after Get during build, got %d", len(s.listeners))
	}
	if _, ok := s.listeners[node]; !ok {
		t.Error("expected node to be in listeners")
	}
}

func TestSignal_Get_OutsideBuild_NoSubscription(t *testing.T) {
	s := NewSignal(42)

	// Ensure we're not in a build context
	oldNode := currentBuildingNode
	currentBuildingNode = nil
	defer func() { currentBuildingNode = oldNode }()

	_ = s.Get()

	if len(s.listeners) != 0 {
		t.Errorf("expected no listeners when not in build context, got %d", len(s.listeners))
	}
}

func TestSignal_Set_SameValue_NoRebuild(t *testing.T) {
	s := NewSignal(42)

	// Subscribe a node
	node := newWidgetNode(nil)
	node.dirty = false // Start clean
	s.listeners[node] = struct{}{}

	// Set same value
	s.Set(42)

	// Node should still be clean
	if node.dirty {
		t.Error("expected node to remain clean when setting same value")
	}
}

func TestSignal_Set_DifferentValue_MarksDirty(t *testing.T) {
	s := NewSignal(42)

	// Subscribe a node
	node := newWidgetNode(nil)
	node.dirty = false // Start clean
	s.listeners[node] = struct{}{}

	// Set different value
	s.Set(100)

	// Node should be dirty
	if !node.dirty {
		t.Error("expected node to be dirty when value changes")
	}
}

func TestSignal_MultipleSubscribers(t *testing.T) {
	s := NewSignal(0)

	// Subscribe multiple nodes
	node1 := newWidgetNode(nil)
	node2 := newWidgetNode(nil)
	node3 := newWidgetNode(nil)

	oldNode := currentBuildingNode
	defer func() { currentBuildingNode = oldNode }()

	currentBuildingNode = node1
	_ = s.Get()
	currentBuildingNode = node2
	_ = s.Get()
	currentBuildingNode = node3
	_ = s.Get()

	if len(s.listeners) != 3 {
		t.Errorf("expected 3 listeners, got %d", len(s.listeners))
	}
}

func TestSignal_Set_NotifiesAllSubscribers(t *testing.T) {
	s := NewSignal(0)

	// Subscribe multiple nodes
	node1 := newWidgetNode(nil)
	node1.dirty = false
	node2 := newWidgetNode(nil)
	node2.dirty = false
	node3 := newWidgetNode(nil)
	node3.dirty = false

	s.listeners[node1] = struct{}{}
	s.listeners[node2] = struct{}{}
	s.listeners[node3] = struct{}{}

	// Change value
	s.Set(1)

	// All should be dirty
	if !node1.dirty {
		t.Error("expected node1 to be dirty")
	}
	if !node2.dirty {
		t.Error("expected node2 to be dirty")
	}
	if !node3.dirty {
		t.Error("expected node3 to be dirty")
	}
}

func TestSignal_Unsubscribe(t *testing.T) {
	s := NewSignal(42)

	node := newWidgetNode(nil)
	s.listeners[node] = struct{}{}

	if len(s.listeners) != 1 {
		t.Fatalf("expected 1 listener, got %d", len(s.listeners))
	}

	s.unsubscribe(node)

	if len(s.listeners) != 0 {
		t.Errorf("expected 0 listeners after unsubscribe, got %d", len(s.listeners))
	}
}

func TestSignal_Unsubscribe_NonExistent(t *testing.T) {
	s := NewSignal(42)
	node := newWidgetNode(nil)

	// Should not panic when unsubscribing non-existent node
	s.unsubscribe(node)

	if len(s.listeners) != 0 {
		t.Errorf("expected 0 listeners, got %d", len(s.listeners))
	}
}

// AnySignal tests

func TestNewAnySignal_InitialValue(t *testing.T) {
	s := NewAnySignal([]int{1, 2, 3})

	if len(s.value) != 3 {
		t.Errorf("expected slice of length 3, got %d", len(s.value))
	}
}

func TestAnySignal_Get_ReturnsValue(t *testing.T) {
	s := NewAnySignal([]string{"a", "b"})

	result := s.Get()
	if len(result) != 2 || result[0] != "a" || result[1] != "b" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestAnySignal_Set_UpdatesValue(t *testing.T) {
	s := NewAnySignal([]int{1})
	s.Set([]int{1, 2, 3})

	if len(s.Get()) != 3 {
		t.Errorf("expected slice of length 3, got %d", len(s.Get()))
	}
}

func TestAnySignal_Peek_ReturnsValue(t *testing.T) {
	s := NewAnySignal(map[string]int{"a": 1})

	result := s.Peek()
	if result["a"] != 1 {
		t.Errorf("expected map with a=1, got %v", result)
	}
}

func TestAnySignal_Update_FunctionalUpdate(t *testing.T) {
	s := NewAnySignal([]int{1, 2, 3})
	s.Update(func(v []int) []int {
		return append(v, 4)
	})

	if len(s.Get()) != 4 {
		t.Errorf("expected slice of length 4, got %d", len(s.Get()))
	}
}

func TestAnySignal_Set_AlwaysNotifies(t *testing.T) {
	s := NewAnySignal([]int{1, 2, 3})

	node := newWidgetNode(nil)
	node.dirty = false
	s.listeners[node] = struct{}{}

	// Set same content (but AnySignal can't compare, so it always notifies)
	s.Set([]int{1, 2, 3})

	if !node.dirty {
		t.Error("expected node to be dirty - AnySignal always notifies")
	}
}

func TestAnySignal_Peek_DoesNotSubscribe(t *testing.T) {
	s := NewAnySignal([]int{1, 2, 3})

	node := newWidgetNode(nil)
	oldNode := currentBuildingNode
	currentBuildingNode = node
	defer func() { currentBuildingNode = oldNode }()

	_ = s.Peek()

	if len(s.listeners) != 0 {
		t.Errorf("expected no listeners after Peek, got %d", len(s.listeners))
	}
}

func TestAnySignal_Get_DuringBuild_Subscribes(t *testing.T) {
	s := NewAnySignal([]int{1, 2, 3})

	node := newWidgetNode(nil)
	oldNode := currentBuildingNode
	currentBuildingNode = node
	defer func() { currentBuildingNode = oldNode }()

	_ = s.Get()

	if len(s.listeners) != 1 {
		t.Errorf("expected 1 listener after Get during build, got %d", len(s.listeners))
	}
}

// Test with different types

func TestSignal_StringType(t *testing.T) {
	s := NewSignal("")
	s.Set("hello")
	s.Update(func(v string) string { return v + " world" })

	if s.Get() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", s.Get())
	}
}

func TestSignal_BoolType(t *testing.T) {
	s := NewSignal(false)
	s.Set(true)

	if !s.Get() {
		t.Error("expected true")
	}

	s.Update(func(v bool) bool { return !v })

	if s.Get() {
		t.Error("expected false after toggle")
	}
}

func TestSignal_FloatType(t *testing.T) {
	s := NewSignal(0.0)
	s.Set(3.14)

	if s.Get() != 3.14 {
		t.Errorf("expected 3.14, got %f", s.Get())
	}
}

func TestAnySignal_StructType(t *testing.T) {
	type person struct {
		name string
		age  int
	}

	s := NewAnySignal(person{name: "Alice", age: 30})

	if s.Get().name != "Alice" {
		t.Errorf("expected 'Alice', got '%s'", s.Get().name)
	}
	if s.Get().age != 30 {
		t.Errorf("expected 30, got %d", s.Get().age)
	}
}

func TestAnySignal_MapType(t *testing.T) {
	s := NewAnySignal(map[string]int{})
	s.Update(func(m map[string]int) map[string]int {
		m["a"] = 1
		return m
	})

	if s.Get()["a"] != 1 {
		t.Errorf("expected map['a'] = 1, got %d", s.Get()["a"])
	}
}
