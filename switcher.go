package terma

// Switcher displays one child at a time based on a string key.
// Only the child matching the Active key is rendered; others are not built.
//
// State preservation: Widget state lives in Signals and State objects
// (e.g., ListState, ScrollState) held by the App, not in widgets themselves.
// When switching between children, their state objects persist and the
// widget rebuilds with that state when reactivated.
//
// Example:
//
//	Switcher{
//	    Active: a.activeTab.Get(),
//	    Children: map[string]Widget{
//	        "home":     HomeTab{listState: a.homeListState},
//	        "settings": SettingsTab{},
//	        "profile":  ProfileTab{},
//	    },
//	}
type Switcher struct {
	// Active is the key of the currently visible child.
	Active string

	// Children maps string keys to widgets. Only the widget matching
	// Active is rendered.
	Children map[string]Widget

	// Width specifies the width of the switcher container.
	// Deprecated: use Style.Width.
	Width Dimension

	// Height specifies the height of the switcher container.
	// Deprecated: use Style.Height.
	Height Dimension

	// Style applies styling to the switcher container.
	Style Style
}

// Build returns the active child widget, or EmptyWidget if the key is not found.
func (s Switcher) Build(ctx BuildContext) Widget {
	if child, ok := s.Children[s.Active]; ok {
		return child
	}
	return EmptyWidget{}
}

// GetContentDimensions returns the configured width and height.
func (s Switcher) GetContentDimensions() (Dimension, Dimension) {
	dims := s.Style.GetDimensions()
	width, height := dims.Width, dims.Height
	if width.IsUnset() {
		width = s.Width
	}
	if height.IsUnset() {
		height = s.Height
	}
	return width, height
}

// GetStyle returns the configured style.
func (s Switcher) GetStyle() Style {
	return s.Style
}
