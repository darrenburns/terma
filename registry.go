package terma

// Rect represents a rectangular region in terminal coordinates.
type Rect struct {
	X, Y          int
	Width, Height int
}

// Contains returns true if the point (x, y) is within this rectangle.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width &&
		y >= r.Y && y < r.Y+r.Height
}

// IsEmpty returns true if the rect has zero or negative area.
func (r Rect) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

// Intersect returns the intersection of two rectangles.
// Returns a zero-size rect if they don't overlap.
func (r Rect) Intersect(other Rect) Rect {
	x1 := max(r.X, other.X)
	y1 := max(r.Y, other.Y)
	x2 := min(r.X+r.Width, other.X+other.Width)
	y2 := min(r.Y+r.Height, other.Y+other.Height)

	if x2 <= x1 || y2 <= y1 {
		return Rect{} // No intersection
	}
	return Rect{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}
}

// WidgetEntry stores a widget along with its position and identity.
type WidgetEntry struct {
	Widget      Widget
	EventWidget Widget
	ID          string
	Bounds      Rect
}

// WidgetRegistry tracks all widgets and their positions during render.
// Widgets are recorded in render order (depth-first), so later entries
// are "on top" visually and should receive events first.
type WidgetRegistry struct {
	entries    []WidgetEntry
	totalCount int // All widgets including those scrolled out of view
}

// NewWidgetRegistry creates a new widget registry.
func NewWidgetRegistry() *WidgetRegistry {
	return &WidgetRegistry{}
}

// Record adds a widget to the registry with its bounds and optional ID.
func (r *WidgetRegistry) Record(widget Widget, eventWidget Widget, id string, bounds Rect) {
	if eventWidget == nil {
		eventWidget = widget
	}
	r.entries = append(r.entries, WidgetEntry{
		Widget:      widget,
		EventWidget: eventWidget,
		ID:          id,
		Bounds:      bounds,
	})
}

// WidgetAt returns the topmost widget containing the point (x, y).
// Returns nil if no widget contains the point.
// Since widgets are recorded in render order, we search back-to-front
// to find the topmost (last rendered) widget at this position.
func (r *WidgetRegistry) WidgetAt(x, y int) *WidgetEntry {
	// Search from back to front (topmost widgets are rendered last)
	for i := len(r.entries) - 1; i >= 0; i-- {
		if r.entries[i].Bounds.Contains(x, y) {
			return &r.entries[i]
		}
	}
	return nil
}

// Entries returns all recorded widget entries.
func (r *WidgetRegistry) Entries() []WidgetEntry {
	return r.entries
}

// WidgetByID returns the widget entry with the given ID.
// Returns nil if no widget has that ID.
func (r *WidgetRegistry) WidgetByID(id string) *WidgetEntry {
	if id == "" {
		return nil
	}
	for i := range r.entries {
		if r.entries[i].ID == id {
			return &r.entries[i]
		}
	}
	return nil
}

// ScrollableAt returns the innermost Scrollable widget containing the point (x, y).
// Returns nil if no Scrollable contains the point.
// Since widgets are recorded in render order (parents before children),
// we search back-to-front to find the innermost scrollable.
func (r *WidgetRegistry) ScrollableAt(x, y int) *Scrollable {
	for i := len(r.entries) - 1; i >= 0; i-- {
		entry := &r.entries[i]
		if entry.Bounds.Contains(x, y) {
			// Check for pointer first (e.g., &Scrollable{...})
			if scrollable, ok := entry.Widget.(*Scrollable); ok {
				return scrollable
			}
			// Then check for value (e.g., Scrollable{...})
			if scrollable, ok := entry.Widget.(Scrollable); ok {
				return &scrollable
			}
		}
	}
	return nil
}

// ScrollablesAt returns all Scrollable widgets containing the point (x, y),
// ordered from innermost to outermost.
// Since widgets are recorded in render order (parents before children),
// we search back-to-front and collect all matching scrollables.
func (r *WidgetRegistry) ScrollablesAt(x, y int) []*Scrollable {
	var scrollables []*Scrollable
	for i := len(r.entries) - 1; i >= 0; i-- {
		entry := &r.entries[i]
		if entry.Bounds.Contains(x, y) {
			// Check for pointer first (e.g., &Scrollable{...})
			if scrollable, ok := entry.Widget.(*Scrollable); ok {
				scrollables = append(scrollables, scrollable)
			}
			// Then check for value (e.g., Scrollable{...})
			if scrollable, ok := entry.Widget.(Scrollable); ok {
				scrollables = append(scrollables, &scrollable)
			}
		}
	}
	return scrollables
}

// Reset clears all entries for a new render pass.
func (r *WidgetRegistry) Reset() {
	r.entries = r.entries[:0]
	r.totalCount = 0
}

// IncrementTotal increments the total widget count (including non-visible widgets).
func (r *WidgetRegistry) IncrementTotal() {
	r.totalCount++
}

// TotalCount returns the total number of widgets rendered, including those
// scrolled out of view that aren't in the entries list.
func (r *WidgetRegistry) TotalCount() int {
	return r.totalCount
}
