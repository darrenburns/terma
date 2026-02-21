package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCheckboxState(t *testing.T) {
	t.Run("creates unchecked state", func(t *testing.T) {
		state := NewCheckboxState(false)
		assert.False(t, state.IsChecked())
	})

	t.Run("creates checked state", func(t *testing.T) {
		state := NewCheckboxState(true)
		assert.True(t, state.IsChecked())
	})
}

func TestCheckboxState_Toggle(t *testing.T) {
	state := NewCheckboxState(false)
	assert.False(t, state.IsChecked())

	state.Toggle()
	assert.True(t, state.IsChecked())

	state.Toggle()
	assert.False(t, state.IsChecked())
}

func TestCheckboxState_SetChecked(t *testing.T) {
	state := NewCheckboxState(false)

	state.SetChecked(true)
	assert.True(t, state.IsChecked())

	state.SetChecked(false)
	assert.False(t, state.IsChecked())

	// Setting to same value should still work
	state.SetChecked(false)
	assert.False(t, state.IsChecked())
}

func TestCheckbox_WidgetID(t *testing.T) {
	checkbox := &Checkbox{ID: "test-checkbox"}
	assert.Equal(t, "test-checkbox", checkbox.WidgetID())
}

func TestCheckbox_IsFocusable(t *testing.T) {
	checkbox := &Checkbox{}
	assert.True(t, checkbox.IsFocusable())
}

func TestCheckbox_Keybinds(t *testing.T) {
	checkbox := &Checkbox{}
	keybinds := checkbox.Keybinds()

	assert.Len(t, keybinds, 2)

	// Check for Enter keybind
	hasEnter := false
	hasSpace := false
	for _, kb := range keybinds {
		if kb.Key == "enter" {
			hasEnter = true
			assert.Equal(t, "Toggle", kb.Name)
		}
		if kb.Key == " " {
			hasSpace = true
			assert.Equal(t, "Toggle", kb.Name)
		}
	}
	assert.True(t, hasEnter, "should have enter keybind")
	assert.True(t, hasSpace, "should have space keybind")
}

func TestCheckbox_Toggle(t *testing.T) {
	t.Run("toggles state via keybind action", func(t *testing.T) {
		state := NewCheckboxState(false)
		checkbox := &Checkbox{State: state}

		keybinds := checkbox.Keybinds()
		keybinds[0].Action() // Toggle via keybind

		assert.True(t, state.IsChecked())
	})

	t.Run("calls OnChange callback", func(t *testing.T) {
		state := NewCheckboxState(false)
		var callbackValue bool
		callbackCalled := false

		checkbox := &Checkbox{
			State: state,
			OnChange: func(checked bool) {
				callbackCalled = true
				callbackValue = checked
			},
		}

		keybinds := checkbox.Keybinds()
		keybinds[0].Action()

		assert.True(t, callbackCalled)
		assert.True(t, callbackValue)
	})

	t.Run("handles nil state gracefully", func(t *testing.T) {
		checkbox := &Checkbox{}
		keybinds := checkbox.Keybinds()

		// Should not panic
		keybinds[0].Action()
	})
}

func TestCheckbox_OnKey(t *testing.T) {
	checkbox := &Checkbox{}
	// OnKey should always return false since keybinds handle everything
	assert.False(t, checkbox.OnKey(KeyEvent{}))
}

func TestCheckbox_GetContentDimensions(t *testing.T) {
	checkbox := &Checkbox{
		Width:  Cells(20),
		Height: Cells(1),
	}

	w, h := checkbox.GetContentDimensions()
	assert.Equal(t, Cells(20), w)
	assert.Equal(t, Cells(1), h)
}

func TestCheckbox_OnClick(t *testing.T) {
	t.Run("toggles state on click", func(t *testing.T) {
		state := NewCheckboxState(false)
		checkbox := &Checkbox{State: state}

		checkbox.OnClick(MouseEvent{})

		assert.True(t, state.IsChecked())
	})

	t.Run("calls OnChange on click", func(t *testing.T) {
		state := NewCheckboxState(false)
		callbackCalled := false

		checkbox := &Checkbox{
			State: state,
			OnChange: func(checked bool) {
				callbackCalled = true
			},
		}

		checkbox.OnClick(MouseEvent{})

		assert.True(t, callbackCalled)
	})

	t.Run("calls Click callback", func(t *testing.T) {
		state := NewCheckboxState(false)
		clickCalled := false

		checkbox := &Checkbox{
			State: state,
			Click: func(event MouseEvent) {
				clickCalled = true
			},
		}

		checkbox.OnClick(MouseEvent{})

		assert.True(t, clickCalled)
	})
}

func TestCheckbox_OnMouseDown(t *testing.T) {
	t.Run("calls MouseDown callback", func(t *testing.T) {
		mouseDownCalled := false
		checkbox := &Checkbox{
			MouseDown: func(event MouseEvent) {
				mouseDownCalled = true
			},
		}

		checkbox.OnMouseDown(MouseEvent{})

		assert.True(t, mouseDownCalled)
	})

	t.Run("handles nil callback", func(t *testing.T) {
		checkbox := &Checkbox{}
		// Should not panic
		checkbox.OnMouseDown(MouseEvent{})
	})
}

func TestCheckbox_OnMouseUp(t *testing.T) {
	t.Run("calls MouseUp callback", func(t *testing.T) {
		mouseUpCalled := false
		checkbox := &Checkbox{
			MouseUp: func(event MouseEvent) {
				mouseUpCalled = true
			},
		}

		checkbox.OnMouseUp(MouseEvent{})

		assert.True(t, mouseUpCalled)
	})

	t.Run("handles nil callback", func(t *testing.T) {
		checkbox := &Checkbox{}
		// Should not panic
		checkbox.OnMouseUp(MouseEvent{})
	})
}

func TestCheckbox_OnHover(t *testing.T) {
	t.Run("calls Hover callback with HoverEnter", func(t *testing.T) {
		var received HoverEvent
		checkbox := &Checkbox{
			Hover: func(event HoverEvent) {
				received = event
			},
		}

		checkbox.OnHover(HoverEvent{Type: HoverEnter, WidgetID: "checkbox"})
		assert.Equal(t, HoverEnter, received.Type)
		assert.Equal(t, "checkbox", received.WidgetID)
	})

	t.Run("calls Hover callback with HoverLeave", func(t *testing.T) {
		var received HoverEvent
		checkbox := &Checkbox{
			Hover: func(event HoverEvent) {
				received = event
			},
		}

		checkbox.OnHover(HoverEvent{Type: HoverLeave, WidgetID: "checkbox"})
		assert.Equal(t, HoverLeave, received.Type)
		assert.Equal(t, "checkbox", received.WidgetID)
	})

	t.Run("handles nil callback", func(t *testing.T) {
		checkbox := &Checkbox{}
		// Should not panic
		checkbox.OnHover(HoverEvent{Type: HoverEnter})
	})
}
