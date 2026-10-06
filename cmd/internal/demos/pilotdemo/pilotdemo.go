// Package pilotdemo is a small sign-up app shown next to the Pilot test that
// drives it (pilotdemo_test.go): typing, clicks, a background task, a spinner,
// a dialog and the clipboard, all checked without a terminal.
package pilotdemo

import (
	"context"
	_ "embed"
	"strings"
	"time"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "pilot",
	Title:       "Pilot",
	Description: "An app and the test that drives it with keys and clicks",
}

//go:embed pilotdemo_test.go
var testSource string

// Signup asks for a name, signs up in the background, then shows an invite
// code that can be copied.
type Signup struct {
	name    *t.TextInputState
	task    *t.Task[string]
	spinner *t.SpinnerState
	problem t.Signal[string]
	invite  t.Signal[string]
	// wait stands in for the network call that signs up.
	wait   func(context.Context) error
	source *t.ScrollState
}

// New creates the demo.
func New() demokit.Demo {
	return newSignup(func(ctx context.Context) error {
		select {
		case <-time.After(1200 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}

func newSignup(wait func(context.Context) error) *Signup {
	return &Signup{
		name:    t.NewTextInputState(""),
		task:    t.NewTask[string](),
		spinner: t.NewSpinnerState(t.SpinnerDots),
		problem: t.NewSignal(""),
		invite:  t.NewSignal(""),
		wait:    wait,
		source:  t.NewScrollState(),
	}
}

// InitialFocus starts in the name field.
func (s *Signup) InitialFocus() string { return "name" }

func (s *Signup) Keybinds() []t.Keybind {
	return []t.Keybind{{Key: "t", Name: "Theme", Action: demokit.NextTheme, Hidden: true}}
}

func (s *Signup) submit() {
	name := strings.TrimSpace(s.name.GetText())
	if name == "" {
		s.problem.Set("Enter a name first")
		t.RequestFocus("name")
		return
	}
	s.problem.Set("")
	s.spinner.Start()
	s.task.Start(func(ctx context.Context) (string, error) {
		if err := s.wait(ctx); err != nil {
			return "", err
		}
		code := "INVITE-" + strings.ToUpper(name)
		t.Dispatch(func() {
			s.spinner.Stop()
			s.invite.Set(code)
		})
		return code, nil
	})
}

func (s *Signup) closeWelcome() {
	s.invite.Set("")
	s.name.SetText("")
}

func (s *Signup) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	phase := s.task.Phase.Get()
	return t.Dock{
		Style:  t.Style{BackgroundColor: theme.Background},
		Top:    []t.Widget{demokit.Header{Title: "Pilot", Tagline: "The app on the left is tested by the code on the right"}},
		Bottom: []t.Widget{demokit.Footer(theme)},
		Body: t.Row{
			Spacing: 2,
			Style:   t.Style{Padding: t.EdgeInsetsXY(2, 1)},
			Children: []t.Widget{
				s.form(theme, phase),
				t.Scrollable{
					ID:        "source",
					State:     s.source,
					Focusable: true,
					Style:     demokit.PanelStyle(theme, "pilotdemo_test.go", ctx.IsFocused(t.Scrollable{ID: "source"})),
					Width:     t.Flex(3),
					Height:    t.Flex(1),
					Child:     t.Text{Content: flowTest(), Style: t.Style{ForegroundColor: theme.TextMuted}},
				},
			},
		},
	}
}

func (s *Signup) form(theme t.ThemeData, phase t.TaskPhase) t.Widget {
	status := t.Widget(t.Text{ID: "status", Content: s.problem.Get(), Style: t.Style{ForegroundColor: theme.Error}})
	if phase == t.TaskRunning {
		status = t.Row{ID: "status", Spacing: 1, Children: []t.Widget{
			t.Spinner{ID: "spinner", State: s.spinner, Style: t.Style{ForegroundColor: theme.Accent}},
			t.Text{Content: "Signing up...", Style: t.Style{ForegroundColor: theme.TextMuted}},
		}}
	}
	return t.Column{
		Spacing: 1,
		Width:   t.Flex(2),
		Style:   demokit.PanelStyle(theme, "Sign up", false),
		Children: []t.Widget{
			t.Text{Content: "Name", Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.TextInput{ID: "name", State: s.name, Placeholder: "Ada Lovelace", OnSubmit: func(string) { s.submit() },
				Style: t.Style{Width: t.Flex(1), BackgroundColor: theme.Surface}},
			t.Button{ID: "signup", Label: "Sign up", OnPress: s.submit},
			status,
			t.Dialog{
				ID:      "welcome",
				Visible: s.invite.Get() != "",
				Title:   "Welcome",
				Content: t.Column{Children: []t.Widget{t.Text{ID: "invite", Content: "Your invite code is " + s.invite.Get()}}},
				Buttons: []t.Button{
					{ID: "copy", Label: "Copy code", OnPress: func() { t.SetClipboard(t.SystemClipboard, s.invite.Peek()) }},
					{ID: "close", Label: "Close", OnPress: s.closeWelcome},
				},
				OnDismiss: s.closeWelcome,
			},
		},
	}
}

// flowTest is TestSignupFlow from the embedded test file, tabs expanded.
func flowTest() string {
	source := strings.ReplaceAll(testSource, "\t", "    ")
	start := strings.Index(source, "func TestSignupFlow")
	if start < 0 {
		return source
	}
	end := strings.Index(source[start:], "\n}\n")
	if end < 0 {
		return source[start:]
	}
	return source[start : start+end+2]
}
