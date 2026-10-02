package inputs

import (
	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/docs/widget-examples/demo"
)

// --8<-- [start:button]
func Button() t.Widget {
	status := t.NewSignal("Ready to save")
	return t.Column{Spacing: 1, Children: []t.Widget{
		t.Button{ID: "save", Label: "Save", Variant: t.ButtonPrimary,
			OnPress: func() { status.Set("Saved") }},
		t.SignalText(status, func(value string) string { return value }),
	}}
}

// --8<-- [end:button]

// --8<-- [start:checkbox]
func Checkbox() t.Widget {
	state := t.NewCheckboxState(true)
	return &t.Checkbox{
		ID: "notifications", State: state, Label: "Enable notifications",
	}
}

// --8<-- [end:checkbox]

// --8<-- [start:textinput]
func TextInput() t.Widget {
	state := t.NewTextInputState("hello@terma.dev")
	status := t.NewSignal("Press Enter to submit")
	statusText := t.SignalText(status, func(value string) string { return value })
	statusText.LayoutStyle.Width = t.Cells(40)
	return t.Column{Spacing: 1, Children: []t.Widget{
		t.TextInput{ID: "email", State: state, Placeholder: "Email address",
			Style:    t.Style{Width: t.Cells(30)},
			OnSubmit: func(value string) { status.Set("Submitted " + value) }},
		statusText,
	}}
}

// --8<-- [end:textinput]

// --8<-- [start:textarea]
func TextArea() t.Widget {
	state := t.NewTextAreaState("Release notes\n\nAdd a summary of your changes.")
	return t.TextArea{
		ID: "notes", State: state,
		Style: t.Style{Width: t.Cells(38), Height: t.Cells(6)},
	}
}

// --8<-- [end:textarea]

// --8<-- [start:autocomplete]
func Autocomplete() t.Widget {
	state := t.NewAutocompleteState()
	state.SetSuggestions([]t.Suggestion{
		{Label: "London", Description: "United Kingdom"},
		{Label: "Lisbon", Description: "Portugal"},
		{Label: "Lima", Description: "Peru"},
	})
	state.Show()
	return t.Autocomplete{
		ID: "cities", State: state, DismissOnBlur: t.BoolPtr(false),
		PopupWidth: t.Cells(32), AnchorToInput: true,
		Child: t.TextInput{ID: "city", State: t.NewTextInputState("L"),
			Style: t.Style{Width: t.Cells(32)}},
	}
}

// --8<-- [end:autocomplete]

// --8<-- [start:commandpalette]
func CommandPalette() t.Widget {
	state := t.NewCommandPaletteState("Commands", nil)
	status := t.NewSignal("Choose a command")
	state.SetItems([]t.CommandPaletteItem{
		{Label: "New note", Action: func() { status.Set("New note selected"); state.Close() }},
		{Label: "Open settings", Action: func() { status.Set("Open settings selected"); state.Close() }},
	})
	state.Open()
	return t.Column{Spacing: 1, Children: []t.Widget{
		t.Button{ID: "commands", Label: "Commands", OnPress: state.Open},
		t.SignalText(status, func(value string) string { return value }),
		t.CommandPalette{ID: "palette", State: state,
			Style: t.Style{Width: t.Cells(38), Height: t.Cells(6)}},
	}}
}

// --8<-- [end:commandpalette]

// --8<-- [start:menu]
type menuExample struct {
	state   *t.MenuState
	visible t.Signal[bool]
	status  t.Signal[string]
}

func Menu() t.Widget {
	t.RequestFocus("file-menu")
	return menuExample{
		state: t.NewMenuState([]t.MenuItem{
			{Label: "New note"}, {Label: "Open note"},
			{Divider: "More"}, {Label: "Export", Disabled: true},
		}),
		visible: t.NewSignal(true), status: t.NewSignal("Choose a menu item"),
	}
}

func (m menuExample) Build(ctx t.BuildContext) t.Widget {
	return t.Column{Spacing: 1, Children: []t.Widget{
		t.Button{ID: "file", Label: "File", OnPress: func() { m.visible.Set(true); t.RequestFocus("file-menu") }},
		t.SignalText(m.status, func(value string) string { return value }),
		t.ShowWhen(m.visible.Get(), t.Menu{
			ID: "file-menu", State: m.state, AnchorID: "file",
			Style:     t.Style{Width: t.Cells(24)},
			OnSelect:  func(item t.MenuItem) { m.status.Set(item.Label); m.visible.Set(false) },
			OnDismiss: func() { m.visible.Set(false) },
		}),
	}}
}

// --8<-- [end:menu]

// --8<-- [start:dialog]
type dialogExample struct{ visible t.Signal[bool] }

func Dialog() t.Widget {
	return dialogExample{visible: t.NewSignal(true)}
}

func (d dialogExample) Build(ctx t.BuildContext) t.Widget {
	closeDialog := func() { d.visible.Set(false) }
	return t.Column{Children: []t.Widget{
		t.Button{ID: "show-dialog", Label: "Show dialog", OnPress: func() { d.visible.Set(true) }},
		t.Dialog{
			ID: "welcome", Visible: d.visible.Get(), Title: "Welcome",
			Content:   t.Text{Content: "Your workspace is ready."},
			Buttons:   []t.Button{{Label: "Continue", Variant: t.ButtonPrimary, OnPress: closeDialog}},
			OnDismiss: closeDialog, Style: t.Style{Width: t.Cells(30)},
		},
	}}
}

// --8<-- [end:dialog]

func Examples() []demo.Example {
	return []demo.Example{
		{Name: "button", Widget: Button, Width: 32, Height: 5},
		{Name: "checkbox", Widget: Checkbox, Width: 32, Height: 3},
		{Name: "textinput", Widget: TextInput, Width: 40, Height: 5},
		{Name: "textarea", Widget: TextArea, Width: 42, Height: 8},
		{Name: "autocomplete", Widget: Autocomplete, Width: 38, Height: 9},
		{Name: "commandpalette", Widget: CommandPalette, Width: 48, Height: 13},
		{Name: "menu", Widget: Menu, Width: 36, Height: 10},
		{Name: "dialog", Widget: Dialog, Width: 44, Height: 11},
	}
}
