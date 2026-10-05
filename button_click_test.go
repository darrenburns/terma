package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
)

// buttonClickApp lays out a toggle for disabling a "save" button, the save
// button itself (at x 0-5, y 1) and a plain button (at x 7-13, y 1).
type buttonClickApp struct {
	disabled Signal[bool]
	presses  *int
	clicks   *[]uv.MouseButton
}

func (a buttonClickApp) Build(ctx BuildContext) Widget {
	return Column{Children: []Widget{
		Text{Content: "header"},
		Row{Spacing: 1, Children: []Widget{
			DisabledWhen(a.disabled.Get(), Button{
				ID:      "save",
				Label:   "Save",
				OnPress: func() { *a.presses++ },
				Click:   func(e MouseEvent) { *a.clicks = append(*a.clicks, e.Button) },
			}),
			Button{ID: "other", Label: "Other"},
		}},
	}}
}

func newButtonClickScene(t *testing.T, disabled bool) (*Pilot, buttonClickApp) {
	app := buttonClickApp{disabled: NewSignal(disabled), presses: new(int), clicks: &[]uv.MouseButton{}}
	return NewPilot(t, app, 30, 5), app
}

func TestButton_LeftClickPresses(t *testing.T) {
	p, app := newButtonClickScene(t, false)

	p.ClickAt(2, 1)

	assert.Equal(t, 1, *app.presses, "a left click presses the button")
	assert.Equal(t, []uv.MouseButton{uv.MouseLeft}, *app.clicks, "the Click callback still runs")
	assert.Equal(t, "save", p.FocusedID(), "the clicked button takes focus")

	// A second click after a retained re-render presses it again.
	p.ClickAt(2, 1)
	assert.Equal(t, 2, *app.presses)
}

func TestButton_OtherMouseButtonsDoNotPress(t *testing.T) {
	p, app := newButtonClickScene(t, false)

	p.MouseDown(2, 1, uv.MouseRight, 0)
	p.MouseUp(2, 1, uv.MouseRight, 0)
	p.MouseDown(2, 1, uv.MouseMiddle, 0)
	p.MouseUp(2, 1, uv.MouseMiddle, 0)

	assert.Equal(t, 0, *app.presses, "only a left click presses the button")
	assert.Equal(t, []uv.MouseButton{uv.MouseRight, uv.MouseMiddle}, *app.clicks,
		"the Click callback receives every button")
}

func TestButton_DisabledIgnoresClicks(t *testing.T) {
	// Clicked on the first frame (full render) and after a retained re-render.
	p, app := newButtonClickScene(t, true)
	p.ClickAt(8, 1) // focus the other button
	p.ClickAt(2, 1)

	assert.Equal(t, 0, *app.presses, "a disabled button is not pressed")
	assert.Empty(t, *app.clicks, "a disabled button's Click callback doesn't run")
	assert.Equal(t, "other", p.FocusedID(), "a disabled button doesn't take focus")

	// Enabling it through a signal re-renders incrementally; clicks work again.
	app.disabled.Set(false)
	p.settle()
	p.ClickAt(2, 1)
	assert.Equal(t, 1, *app.presses)
	assert.Equal(t, "save", p.FocusedID())

	// And disabling it again stops them.
	app.disabled.Set(true)
	p.settle()
	p.ClickAt(2, 1)
	assert.Equal(t, 1, *app.presses)
}
