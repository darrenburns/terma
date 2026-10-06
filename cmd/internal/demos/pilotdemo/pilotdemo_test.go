package pilotdemo

import (
	"context"
	"testing"
	"time"

	t "github.com/darrenburns/terma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignupFlow(test *testing.T) {
	previous := t.CurrentThemeName()
	t.SetTheme(t.ThemeNameRosePine)
	test.Cleanup(func() { t.SetTheme(previous) })
	signedUp := make(chan struct{})
	app := newSignup(func(context.Context) error { <-signedUp; return nil })
	p := t.NewPilot(test, app, 100, 22)
	p.Click("name")

	// Signing up with no name explains why and returns to the field.
	p.Click("signup")
	assert.Equal(test, "Enter a name first", p.TextOf("status"))
	assert.Equal(test, "name", p.FocusedID())

	// Type a name and press enter: a spinner runs while the work does.
	p.Type("Ada")
	p.Press("enter")
	spinner := p.TextOf("spinner")
	p.Advance(80 * time.Millisecond)
	assert.NotEqual(test, spinner, p.TextOf("spinner"), "the spinner moves with the clock")

	// The background work finishes and a dialog takes focus.
	close(signedUp)
	p.WaitUntil(func() bool { return p.FocusedID() == "copy" }, time.Second)
	assert.Equal(test, "Your invite code is INVITE-ADA", p.TextOf("invite"))
	p.AssertSnapshot("welcome", "The welcome dialog shows INVITE-ADA with 'Copy code' focused")

	// Copy the code, then close the dialog with escape.
	p.Press("enter")
	assert.Equal(test, "INVITE-ADA", p.Clipboard())
	p.Press("escape")
	assert.Equal(test, "name", p.FocusedID(), "focus returns to where it was")
	assert.Empty(test, app.name.GetText(), "closing starts a fresh sign-up")
}

func TestSignupStartsInTheNameField(test *testing.T) {
	p := t.NewPilot(test, New(), 100, 22)
	require.Contains(test, p.ScreenText(), "Sign up")
	p.Press("tab")
	assert.Equal(test, "signup", p.FocusedID())
}
