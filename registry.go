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

// WidgetEntry stores a widget along with its position and identity.
type WidgetEntry struct {
	Widget Widget
	ID     string
	Bounds Rect
}

// WidgetRegistry tracks all widgets and their positions during render.
// Widgets are recorded in render order (depth-first), so later entries
// are "on top" visually and should receive events first.
type WidgetRegistry struct {
	entries []WidgetEntry
}

// NewWidgetRegistry creates a new widget registry.
func NewWidgetRegistry() *WidgetRegistry {
	return &WidgetRegistry{}
}

// Record adds a widget to the registry with its bounds and optional ID.
func (r *WidgetRegistry) Record(widget Widget, id string, bounds Rect) {
	r.entries = append(r.entries, WidgetEntry{
		Widget: widget,
		ID:     id,
		Bounds: bounds,
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
			if scrollable, ok := entry.Widget.(Scrollable); ok {
				return &scrollable
			}
		}
	}
	return nil
}

// Reset clears all entries for a new render pass.
func (r *WidgetRegistry) Reset() {
	r.entries = r.entries[:0]
}

