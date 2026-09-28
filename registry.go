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

// Intersects returns true if two rectangles overlap.
func (r Rect) Intersects(other Rect) bool {
	return !r.Intersect(other).IsEmpty()
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

// Union returns the smallest rectangle covering both rectangles.
func (r Rect) Union(other Rect) Rect {
	if r.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return r
	}
	x1 := min(r.X, other.X)
	y1 := min(r.Y, other.Y)
	x2 := max(r.X+r.Width, other.X+other.Width)
	y2 := max(r.Y+r.Height, other.Y+other.Height)
	return Rect{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}
}

// WidgetEntry stores a widget along with its position and identity.
type WidgetEntry struct {
	Widget      Widget
	EventWidget Widget
	ID          string
	// Bounds is the widget's border box in screen coordinates. Local mouse
	// coordinates are relative to it. It can extend past what is drawn, for
	// example when an ancestor has scrolled part of the widget out of view.
	Bounds Rect
	// Visible is the part of Bounds left after clipping by every ancestor and
	// the screen. Only this area receives pointer events; it is empty for a
	// widget scrolled entirely out of view.
	Visible Rect
}

// WidgetRegistry tracks all widgets and their positions during render.
// Widgets are recorded in render order (depth-first), so later entries
// are "on top" visually and should receive events first.
type WidgetRegistry struct {
	entries    []WidgetEntry
	totalCount int // All widgets including those scrolled out of view

	// rows[y] lists, in record order, the entries whose visible area covers
	// screen row y. It is built by the first hit test after the entries
	// change, so a pointer event costs the widgets on one row rather than
	// every widget recorded (including those scrolled out of view).
	rows      [][]int32
	rowsValid bool
}

// NewWidgetRegistry creates a new widget registry.
func NewWidgetRegistry() *WidgetRegistry {
	return &WidgetRegistry{}
}

// Record adds a widget to the registry with its bounds, visible area and
// optional ID.
func (r *WidgetRegistry) Record(widget Widget, eventWidget Widget, id string, bounds, visible Rect) {
	if eventWidget == nil {
		eventWidget = widget
	}
	r.entries = append(r.entries, WidgetEntry{
		Widget:      widget,
		EventWidget: eventWidget,
		ID:          id,
		Bounds:      bounds,
		Visible:     visible,
	})
	r.rowsValid = false
}

// appendEntries re-records entries from an earlier frame.
func (r *WidgetRegistry) appendEntries(entries []WidgetEntry) {
	r.entries = append(r.entries, entries...)
	r.rowsValid = false
}

// row returns the indexes of the entries whose visible area covers row y.
func (r *WidgetRegistry) row(y int) []int32 {
	if !r.rowsValid {
		r.buildRows()
	}
	if y < 0 || y >= len(r.rows) {
		return nil
	}
	return r.rows[y]
}

func (r *WidgetRegistry) buildRows() {
	for y := range r.rows {
		r.rows[y] = r.rows[y][:0]
	}
	for i := range r.entries {
		visible := r.entries[i].Visible
		if visible.IsEmpty() {
			continue
		}
		start, end := max(visible.Y, 0), visible.Y+visible.Height
		for len(r.rows) < end {
			r.rows = append(r.rows, nil)
		}
		for y := start; y < end; y++ {
			r.rows[y] = append(r.rows[y], int32(i))
		}
	}
	r.rowsValid = true
}

// topmostIn returns the last entry in [lo, hi) whose visible area contains
// (x, y) and that match accepts (nil accepts any).
func (r *WidgetRegistry) topmostIn(x, y, lo, hi int, match func(*WidgetEntry) bool) *WidgetEntry {
	row := r.row(y)
	for i := len(row) - 1; i >= 0; i-- {
		index := int(row[i])
		if index >= hi {
			continue
		}
		if index < lo {
			break
		}
		entry := &r.entries[index]
		if entry.Visible.Contains(x, y) && (match == nil || match(entry)) {
			return entry
		}
	}
	return nil
}

// WidgetAt returns the topmost widget visible at the point (x, y).
// Returns nil if no widget contains the point.
// Since widgets are recorded in render order, the topmost (last rendered)
// widget at this position wins.
func (r *WidgetRegistry) WidgetAt(x, y int) *WidgetEntry {
	return r.widgetAtIn(x, y, 0, len(r.entries))
}

func (r *WidgetRegistry) widgetAtIn(x, y, lo, hi int) *WidgetEntry {
	return r.topmostIn(x, y, lo, hi, nil)
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

// ScrollableAt returns the innermost Scrollable widget visible at the point (x, y).
// Returns nil if no Scrollable contains the point.
func (r *WidgetRegistry) ScrollableAt(x, y int) *Scrollable {
	scrollables := r.scrollablesAtIn(x, y, 0, len(r.entries))
	if len(scrollables) == 0 {
		return nil
	}
	return scrollables[0]
}

// ScrollablesAt returns all Scrollable widgets visible at the point (x, y),
// ordered from innermost to outermost.
func (r *WidgetRegistry) ScrollablesAt(x, y int) []*Scrollable {
	return r.scrollablesAtIn(x, y, 0, len(r.entries))
}

func (r *WidgetRegistry) scrollablesAtIn(x, y, lo, hi int) []*Scrollable {
	var scrollables []*Scrollable
	r.topmostIn(x, y, lo, hi, func(entry *WidgetEntry) bool {
		// Check for pointer first (e.g., &Scrollable{...})
		if scrollable, ok := entry.Widget.(*Scrollable); ok {
			scrollables = append(scrollables, scrollable)
		} else if scrollable, ok := entry.Widget.(Scrollable); ok {
			// Then check for value (e.g., Scrollable{...})
			scrollables = append(scrollables, &scrollable)
		}
		return false
	})
	return scrollables
}

// FocusableAt returns the innermost focusable widget visible at the point (x, y).
// Returns nil if no focusable widget contains the point.
func (r *WidgetRegistry) FocusableAt(x, y int) *WidgetEntry {
	return r.focusableAtIn(x, y, 0, len(r.entries))
}

func (r *WidgetRegistry) focusableAtIn(x, y, lo, hi int) *WidgetEntry {
	return r.topmostIn(x, y, lo, hi, func(entry *WidgetEntry) bool {
		focusable, ok := entry.EventWidget.(Focusable)
		return ok && focusable.IsFocusable()
	})
}

// Reset clears all entries for a new render pass.
func (r *WidgetRegistry) Reset() {
	// A fresh slice, not a truncation: retained nodes keep views of the entries
	// their subtree recorded and replay them when a frame skips the subtree.
	r.entries = make([]WidgetEntry, 0, cap(r.entries))
	r.totalCount = 0
	r.rowsValid = false
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
