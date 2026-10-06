// Command inline-example is a chat composer that runs inline, below the shell
// prompt, rather than taking over the screen. Messages and replies are printed
// above the live region and stay in the terminal's scrollback.
//
//	go run ./cmd/inline-example         # keep the final frame on exit
//	go run ./cmd/inline-example -clear  # erase the region on exit
//	go run ./cmd/inline-example -mouse  # report the mouse
//
// Type "/" for commands. ctrl+c exits.
package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	t "github.com/darrenburns/terma"
)

var commands = []t.Suggestion{
	{Label: "/run", Value: "/run", Description: "Run a shell command"},
	{Label: "/shout", Value: "/shout", Description: "Print a wide banner"},
	{Label: "/quit", Value: "/quit", Description: "Leave the app"},
}

type App struct {
	input    *t.TextInputState
	complete *t.AutocompleteState
	spinner  *t.SpinnerState
	shimmer  *t.ShimmerState
	sent     t.Signal[int]
	thinking t.Signal[bool]
	partial  t.Signal[string]
}

func NewApp() *App {
	complete := t.NewAutocompleteState()
	complete.SetSuggestions(commands)
	return &App{
		input:    t.NewTextInputState(""),
		complete: complete,
		spinner:  t.NewSpinnerState(t.SpinnerDots),
		shimmer:  t.NewShimmerState(1500 * time.Millisecond),
		sent:     t.NewSignal(0),
		thinking: t.NewSignal(false),
		partial:  t.NewSignal(""),
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	thinking := a.thinking.Get()

	return t.Column{
		Style: t.Style{Padding: t.EdgeInsetsXY(1, 0)},
		Children: []t.Widget{
			t.ShowWhen(thinking, t.Row{
				Spacing: 1,
				Children: []t.Widget{
					t.Spinner{State: a.spinner, Style: t.Style{ForegroundColor: theme.Accent}},
					t.Text{
						Content: "Thinking...",
						Style: t.Style{ForegroundColor: t.Shimmer{
							State: a.shimmer, Base: theme.TextMuted, Highlight: theme.Text,
						}},
					},
				},
			}),
			t.ShowWhen(thinking && a.partial.Get() != "", t.Text{
				Content: a.partial.Get(),
				Wrap:    t.WrapSoft,
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			}),
			t.Autocomplete{
				ID:            "commands",
				State:         a.complete,
				TriggerChars:  []rune{'/'},
				AnchorToInput: true,
				PopupWidth:    t.Cells(40),
				Child: t.TextInput{
					ID:          "prompt",
					State:       a.input,
					Placeholder: "Send a message, or / for commands",
					OnSubmit:    a.submit,
					Style: t.Style{
						Width:   t.Flex(1),
						Padding: t.EdgeInsetsXY(1, 0),
						Border:  t.Border{Style: t.BorderRounded, Color: theme.Border},
					},
				},
			},
			t.Text{
				Content: fmt.Sprintf("%d sent · ctrl+c to exit", a.sent.Get()),
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func (a *App) submit(text string) {
	text = strings.TrimSpace(text)
	a.input.SetText("")
	switch {
	case text == "":
		return
	case text == "/quit":
		t.Quit()
		return
	case text == "/run":
		_ = t.RunExternal(exec.Command("sh", "-c", "echo hello from external"))
		return
	case text == "/shout":
		t.PrintAbove(banner{})
		return
	}
	a.sent.Update(func(n int) int { return n + 1 })
	t.PrintAbove(message{who: "You", text: text, user: true})
	go a.reply(text)
}

// reply streams a made-up answer word by word, growing the live region
// while it does, then prints it above and lets the region shrink back.
func (a *App) reply(prompt string) {
	a.thinking.Set(true)
	a.spinner.Start()
	a.shimmer.Start()
	time.Sleep(600 * time.Millisecond)
	answer := fmt.Sprintf("You said %q. Here is a longer reply that streams in a word at a time, "+
		"so the live region grows as it wraps onto more lines and shrinks once it is printed above.", prompt)
	var streamed []string
	for _, word := range strings.Fields(answer) {
		streamed = append(streamed, word)
		a.partial.Set(strings.Join(streamed, " "))
		time.Sleep(60 * time.Millisecond)
	}
	t.PrintAbove(message{who: "Bot", text: answer})
	a.spinner.Stop()
	a.shimmer.Stop()
	a.partial.Set("")
	a.thinking.Set(false)
}

type message struct {
	who, text string
	user      bool
}

func (m message) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	color := theme.Secondary
	if m.user {
		color = theme.Primary
	}
	return t.Column{
		Style: t.Style{Padding: t.EdgeInsets{Left: 1, Right: 1, Bottom: 1}},
		Children: []t.Widget{
			t.Text{Content: m.who, Style: t.Style{ForegroundColor: color, Bold: true}},
			t.Text{Content: m.text, Wrap: t.WrapSoft, Style: t.Style{ForegroundColor: theme.Text}},
		},
	}
}

// banner fills the terminal's width, to show printed lines as wide as the
// terminal take exactly one row.
type banner struct{}

func (banner) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Row{
		Width:     t.Flex(1),
		MainAlign: t.MainAxisCenter,
		Style:     t.Style{BackgroundColor: theme.Primary},
		Children: []t.Widget{t.Text{
			Content: "LOUD NOISES",
			Style:   t.Style{ForegroundColor: theme.TextOnPrimary, Bold: true},
		}},
	}
}

func main() {
	clear := flag.Bool("clear", false, "erase the live region on exit")
	mouse := flag.Bool("mouse", false, "enable mouse reporting")
	flag.Parse()

	opts := t.InlineOptions{Mouse: *mouse}
	if *clear {
		opts.OnExit = t.InlineExitClear
	}
	if err := t.RunInline(NewApp(), opts); err != nil {
		log.Fatal(err)
	}
}
