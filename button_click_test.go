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

func newButtonClickScene(t *testing.T, disabled bool) (*clickScene, buttonClickApp) {
	app := buttonClickApp{disabled: NewSignal(disabled), presses: new(int), clicks: &[]uv.MouseButton{}}
	return newClickScene(t, app, 30, 5), app
}

func pressAt(s *clickScene, x, y int, button uv.MouseButton) {
	s.router.press(uv.MouseClickEvent{X: x, Y: y, Button: button}, 0.5, 0.5, s.now)
	s.router.release(uv.MouseReleaseEvent{X: x, Y: y, Button: button}, 0.5, 0.5)
	s.draw()
}

func TestButton_LeftClickPresses(t *testing.T) {
	s, app := newButtonClickScene(t, false)

	s.click(2, 1, 0)

	assert.Equal(t, 1, *app.presses, "a left click presses the button")
	assert.Equal(t, []uv.MouseButton{uv.MouseLeft}, *app.clicks, "the Click callback still runs")
	assert.Equal(t, "save", s.focus.FocusedID(), "the clicked button takes focus")

	// A second click after a retained re-render presses it again.
	s.click(2, 1, 0)
	assert.Equal(t, 2, *app.presses)
}

func TestButton_OtherMouseButtonsDoNotPress(t *testing.T) {
	s, app := newButtonClickScene(t, false)

	pressAt(s, 2, 1, uv.MouseRight)
	pressAt(s, 2, 1, uv.MouseMiddle)

	assert.Equal(t, 0, *app.presses, "only a left click presses the button")
	assert.Equal(t, []uv.MouseButton{uv.MouseRight, uv.MouseMiddle}, *app.clicks,
		"the Click callback receives every button")
}

func TestButton_DisabledIgnoresClicks(t *testing.T) {
	// Clicked on the first frame (full render) and after a retained re-render.
	s, app := newButtonClickScene(t, true)
	s.click(8, 1, 0) // focus the other button
	s.click(2, 1, 0)

	assert.Equal(t, 0, *app.presses, "a disabled button is not pressed")
	assert.Empty(t, *app.clicks, "a disabled button's Click callback doesn't run")
	assert.Equal(t, "other", s.focus.FocusedID(), "a disabled button doesn't take focus")

	// Enabling it through a signal re-renders incrementally; clicks work again.
	app.disabled.Set(false)
	s.draw()
	s.click(2, 1, 0)
	assert.Equal(t, 1, *app.presses)
	assert.Equal(t, "save", s.focus.FocusedID())

	// And disabling it again stops them.
	app.disabled.Set(true)
	s.draw()
	s.click(2, 1, 0)
	assert.Equal(t, 1, *app.presses)
}
