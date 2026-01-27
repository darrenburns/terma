package terma

import (
	"fmt"
	"sync"
)

var dialogVisibility sync.Map

// Dialog is a convenience widget that wraps Floating + layout to create
// a modal dialog with a title, content, and action buttons.
//
// When Visible is true, the dialog is rendered as a centered modal overlay
// with a backdrop. Focus is automatically trapped within the dialog.
// The first button receives focus when the dialog becomes visible.
//
// Example:
//
//	Dialog{
//	    ID:      "confirm-dialog",
//	    Visible: a.showDialog.Get(),
//	    Title:   "Confirm Delete",
//	    Content: Text{Content: "Are you sure?"},
//	    Buttons: []Button{
//	        {Label: "Cancel", OnPress: func() { a.showDialog.Set(false) }},
//	        {Label: "Delete", Variant: ButtonError, OnPress: a.delete},
//	    },
//	    OnDismiss: func() { a.showDialog.Set(false) },
//	}
type Dialog struct {
	ID        string   // Optional stable ID for focus management and button IDs
	Visible   bool     // Controls visibility
	Title     string   // Shown in border title (optional)
	Content   Widget   // Body content
	Buttons   []Button // Action buttons, rendered left-to-right
	OnDismiss func()   // Called on Escape / click outside (optional)
	Style     Style    // Override default styling
}

// Build registers the dialog with the float collector (when visible) and
// returns EmptyWidget. Like Floating, the dialog content is rendered as an
// overlay after the main widget tree.
func (d Dialog) Build(ctx BuildContext) Widget {
	dialogID := d.ID
	if dialogID == "" {
		dialogID = ctx.AutoID()
	}

	wasVisible := false
	if v, ok := dialogVisibility.Load(dialogID); ok {
		if visible, ok := v.(bool); ok {
			wasVisible = visible
		}
	}

	if !d.Visible {
		dialogVisibility.Store(dialogID, false)
		return EmptyWidget{}
	}

	theme := ctx.Theme()

	// Auto-assign IDs to buttons and request focus on the first one when shown
	buttons := make([]Widget, len(d.Buttons))
	for i, btn := range d.Buttons {
		if btn.ID == "" {
			btn.ID = fmt.Sprintf("%s-btn-%d", dialogID, i)
		}
		if i == 0 && !wasVisible {
			ctx.RequestFocus(btn.ID)
		}
		buttons[i] = btn
	}

	// Build the dialog body
	var children []Widget
	if d.Content != nil {
		children = append(children, d.Content)
	}
	if len(buttons) > 0 {
		children = append(children, Row{
			MainAlign: MainAxisEnd,
			Spacing:   2,
			Children:  buttons,
		})
	}

	// Apply default style
	style := d.Style
	if style.BackgroundColor == nil || !style.BackgroundColor.IsSet() {
		style.BackgroundColor = theme.Surface
	}
	if style.ForegroundColor == nil || !style.ForegroundColor.IsSet() {
		style.ForegroundColor = theme.Text
	}
	if style.Padding == (EdgeInsets{}) {
		style.Padding = EdgeInsetsXY(2, 1)
	}
	if style.Border.IsZero() {
		decorations := []BorderDecoration{}
		if d.Title != "" {
			decorations = append(decorations, BorderTitleCenter(" "+d.Title+" "))
		}
		style.Border = RoundedBorder(theme.Border, decorations...)
	}
	if style.Width.IsUnset() {
		style.Width = Percent(60)
	}

	body := Column{
		Spacing:  1,
		Style:    style,
		Children: children,
	}

	// Register directly with the float collector (same mechanism as Floating.Build).
	// Dialog.Build() must not return a Floating widget because BuildRenderTree
	// only calls Build() on the original widget — the returned widget is used
	// for layout/children only, so a returned Floating would never register.
	if ctx.floatCollector != nil {
		ctx.floatCollector.Add(FloatEntry{
			Config: FloatConfig{
				Position:  FloatPositionCenter,
				Modal:     true,
				OnDismiss: d.OnDismiss,
			},
			Child: body,
		})
	}

	dialogVisibility.Store(dialogID, true)
	return EmptyWidget{}
}
