package terma

// FocusTrap is a transparent wrapper widget that traps focus within its subtree.
// When Active is true, Tab/Shift+Tab cycling is constrained to focusable widgets
// within the FocusTrap's child tree. When Active is false, focus cycles globally
// as normal.
//
// FocusTrap is transparent to layout — it delegates entirely to its Child.
//
// Example - trapping focus within a dialog:
//
//	FocusTrap{
//	    ID:     "dialog-trap",
//	    Active: true,
//	    Child: Column{
//	        Children: []Widget{
//	            TextInput{ID: "name", State: nameState},
//	            Button{ID: "submit", Label: "Submit", OnPress: submit},
//	        },
//	    },
//	}
type FocusTrap struct {
	// ID is a unique identifier for this focus trap scope.
	// Required when Active is true.
	ID string

	// Active controls whether focus trapping is enabled.
	// When false, the FocusTrap is fully transparent and focus cycles globally.
	Active bool

	// Child is the widget subtree in which focus is trapped.
	Child Widget
}

// WidgetID returns the focus trap's unique identifier.
func (ft FocusTrap) WidgetID() string {
	return ft.ID
}

// TrapsFocus returns true when the focus trap is active.
func (ft FocusTrap) TrapsFocus() bool {
	return ft.Active
}

// Build returns the child widget directly.
// FocusTrap is handled as a transparent wrapper by BuildRenderTree,
// so this method is only called as a fallback.
func (ft FocusTrap) Build(ctx BuildContext) Widget {
	if ft.Child == nil {
		return EmptyWidget{}
	}
	return ft.Child.Build(ctx)
}
