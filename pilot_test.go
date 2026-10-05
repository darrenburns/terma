package terma

import (
	"context"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pilotSignupApp is a small form: a name input, a submit button, a status
// line and a confirmation dialog that submitting opens.
type pilotSignupApp struct {
	name      *TextInputState
	status    Signal[string]
	confirmed Signal[bool]
	quitOnQ   bool
}

func newPilotSignupApp() *pilotSignupApp {
	return &pilotSignupApp{name: NewTextInputState(""), status: NewSignal("waiting"), confirmed: NewSignal(false)}
}

func (a *pilotSignupApp) submit() {
	a.status.Set("hello " + a.name.GetText())
	a.confirmed.Set(true)
}

func (a *pilotSignupApp) Keybinds() []Keybind {
	return []Keybind{
		{Key: "ctrl+y", Name: "Copy", Action: func() { SetClipboard(SystemClipboard, a.name.GetText()) }},
		{Key: "ctrl+p", Name: "Paste from clipboard", Action: func() {
			ReadClipboard(SystemClipboard, func(text string) { a.status.Set("clipboard " + text) })
		}},
		{Key: "q", Name: "Quit", Action: Quit},
	}
}

func (a *pilotSignupApp) Build(ctx BuildContext) Widget {
	return Column{Spacing: 1, Children: []Widget{
		TextInput{ID: "name", State: a.name, Placeholder: "Your name", OnSubmit: func(string) { a.submit() }, Style: Style{Width: Cells(20)}},
		Button{ID: "submit", Label: "Submit", OnPress: a.submit},
		Text{ID: "status", Content: a.status.Get()},
		Dialog{
			ID:        "done",
			Visible:   a.confirmed.Get(),
			Title:     "Signed up",
			Content:   Text{Content: "Welcome!"},
			Buttons:   []Button{{ID: "ok", Label: "OK", OnPress: func() { a.confirmed.Set(false) }}},
			OnDismiss: func() { a.confirmed.Set(false) },
		},
	}}
}

func TestPilot_TypeTabEnterOpensDialog(t *testing.T) {
	app := newPilotSignupApp()
	p := NewPilot(t, app, 40, 12)

	require.Equal(t, "name", p.FocusedID(), "the first focusable starts focused, as in Run")
	p.Type("Ada")
	p.Press("tab")
	require.Equal(t, "submit", p.FocusedID())
	p.Press("enter")

	assert.Equal(t, "hello Ada", app.status.Peek())
	assert.Equal(t, "hello Ada", p.TextOf("status"))
	assert.Equal(t, "ok", p.FocusedID(), "the opened modal takes focus")
	p.AssertSnapshot("dialog_open", "The form after typing Ada and submitting: status reads 'hello Ada', a 'Signed up' dialog is open over a dimmed backdrop with its OK button focused.")

	p.Press("escape")
	assert.False(t, app.confirmed.Peek(), "escape dismisses the dialog")
	assert.Equal(t, "submit", p.FocusedID(), "closing the modal restores focus")
}

func TestPilot_ClickIsHitTested(t *testing.T) {
	app := newPilotSignupApp()
	p := NewPilot(t, app, 40, 12)

	p.Click("submit")
	assert.Equal(t, "hello ", app.status.Peek(), "clicking the button presses it")
	assert.Equal(t, "ok", p.FocusedID())

	p.Click("submit")
	assert.Equal(t, "hello ", app.status.Peek())
	assert.True(t, app.confirmed.Peek(), "the modal backdrop takes the click meant for the covered button")

	p.Click("ok")
	assert.False(t, app.confirmed.Peek())
}

func TestPilot_PasteAndClipboard(t *testing.T) {
	app := newPilotSignupApp()
	p := NewPilot(t, app, 40, 12)

	p.Paste("Grace Hopper")
	assert.Equal(t, "Grace Hopper", app.name.GetText(), "a paste reaches the focused input in one piece")

	p.Press("ctrl+y")
	assert.Equal(t, "Grace Hopper", p.Clipboard(), "SetClipboard writes reach the pilot's clipboard")

	p.Press("ctrl+p")
	assert.Equal(t, "clipboard Grace Hopper", p.TextOf("status"), "ReadClipboard is answered from the same clipboard")
}

func TestPilot_QuitStopsTheApp(t *testing.T) {
	p := NewPilot(t, newPilotSignupApp(), 40, 12)
	p.Press("escape")
	require.False(t, p.Exited())
	p.Click("submit")
	p.Press("enter") // OK in the dialog
	p.Press("tab", "tab")
	require.NotEqual(t, "name", p.FocusedID())
	p.Press("q")
	assert.True(t, p.Exited(), "Quit ends the app")
}

func TestPilot_CtrlCQuits(t *testing.T) {
	p := NewPilot(t, newPilotSignupApp(), 40, 12)
	p.Press("ctrl+c")
	assert.True(t, p.Exited())
}

func TestPilot_Resize(t *testing.T) {
	p := NewPilot(t, Text{Content: "a long line of text", Wrap: WrapSoft}, 30, 3)
	assert.Equal(t, "a long line of text", strings.TrimSpace(strings.Split(p.ScreenText(), "\n")[0]))
	p.Resize(8, 3)
	assert.Equal(t, []string{"a long  ", "line of ", "text    "}, strings.Split(p.ScreenText(), "\n"))
}

func TestPilot_AdvanceDrivesAnimations(t *testing.T) {
	spinner := NewSpinnerState(SpinnerStyle{Frames: []string{"1", "2", "3"}, FrameTime: 100 * time.Millisecond})
	spinner.Start()
	p := NewPilot(t, Spinner{State: spinner}, 4, 1)

	frame := func() string { return strings.TrimSpace(p.ScreenText()) }
	assert.Equal(t, "1", frame())
	p.Advance(99 * time.Millisecond)
	assert.Equal(t, "1", frame(), "time is frozen between Advance calls")
	p.Advance(time.Millisecond)
	assert.Equal(t, "2", frame())
	p.Advance(200 * time.Millisecond)
	assert.Equal(t, "1", frame(), "the spinner loops")
}

func TestPilot_AdvanceBlinksCursor(t *testing.T) {
	SetCursorBlink(true)
	t.Cleanup(func() { SetCursorBlink(false) })
	p := NewPilot(t, TextInput{ID: "in", State: NewTextInputState(""), Style: Style{Width: Cells(10)}}, 10, 1)

	cursorCell := func() uv.Style { return p.Buffer().CellAt(0, 0).Style }
	shown := cursorCell()
	p.Advance(cursorBlinkInterval - time.Millisecond)
	assert.Equal(t, shown, cursorCell(), "the cursor stays until the blink interval ends")
	p.Advance(time.Millisecond)
	assert.NotEqual(t, shown, cursorCell(), "the cursor blinks off after the blink interval")
	p.Advance(cursorBlinkInterval)
	assert.Equal(t, shown, cursorCell(), "and back on after another")
	p.Advance(cursorBlinkInterval)
	p.Type("x")
	assert.Equal(t, "x", strings.TrimSpace(p.ScreenText()))
	assert.Equal(t, shown, p.Buffer().CellAt(1, 0).Style, "typing shows the cursor again, after the new character")
}

func TestPilot_WaitUntilSeesBackgroundWork(t *testing.T) {
	task := NewTask[string]()
	root := pilotBuilder(func(BuildContext) Widget {
		if task.Phase.Get() == TaskSuccess {
			return Column{Children: []Widget{Text{ID: "out", Content: task.Value.Get()}}}
		}
		return Column{Children: []Widget{Text{ID: "out", Content: "loading"}}}
	})
	p := NewPilot(t, root, 20, 1)
	release := make(chan struct{})
	task.Start(func(context.Context) (string, error) {
		<-release
		return "done", nil
	})
	assert.Equal(t, "loading", p.TextOf("out"))
	close(release)
	p.WaitUntil(func() bool { return p.TextOf("out") == "done" }, 0)
}

type pilotBuilder func(BuildContext) Widget

func (b pilotBuilder) Build(ctx BuildContext) Widget { return b(ctx) }

func TestPilot_KeyParsing(t *testing.T) {
	cases := map[string]uv.Key{
		"tab":        {Code: uv.KeyTab},
		"enter":      {Code: uv.KeyEnter},
		"esc":        {Code: uv.KeyEscape},
		"escape":     {Code: uv.KeyEscape},
		"space":      {Code: uv.KeySpace, Text: " "},
		"pgdown":     {Code: uv.KeyPgDown},
		"f5":         {Code: uv.KeyF5},
		"shift+down": {Code: uv.KeyDown, Mod: uv.ModShift},
		"ctrl+s":     {Code: 's', Mod: uv.ModCtrl},
		"ctrl++":     {Code: '+', Mod: uv.ModCtrl},
		"+":          {Code: '+', Text: "+"},
		"?":          {Code: '?', Text: "?"},
		"shift+a":    {Code: 'a', Mod: uv.ModShift, Text: "A", ShiftedCode: 'A'},
	}
	for name, want := range cases {
		got, err := keyPress(name)
		require.NoError(t, err, name)
		assert.Equal(t, want, uv.Key(got), name)
		if !strings.Contains(name, "+") || strings.Contains(name[:len(name)-1], "+") && !strings.HasSuffix(name, "++") {
			assert.True(t, uv.Key(got).MatchString(name), "%q must match its own keybind", name)
		}
	}
	for _, bad := range []string{"bogus", "ctl+s"} {
		_, err := keyPress(bad)
		assert.Error(t, err, bad)
	}
}
