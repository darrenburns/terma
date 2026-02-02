package main

import (
	"fmt"
	"log"
	"time"

	t "terma"
)

const maxEntries = 25

type KeyEntry struct {
	at                  time.Time
	key                 string
	text                string
	matchEnter          bool
	matchShiftEnter     bool
	matchCtrlEnter      bool
	matchCtrlShiftEnter bool
}

type KeyEventDebug struct {
	entries []KeyEntry
}

func (a *KeyEventDebug) IsFocusable() bool { return true }

func (a *KeyEventDebug) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
		{Key: "c", Name: "Clear", Action: a.clear},
	}
}

func (a *KeyEventDebug) clear() {
	a.entries = nil
}

func (a *KeyEventDebug) OnKey(event t.KeyEvent) bool {
	entry := KeyEntry{
		at:                  time.Now(),
		key:                 event.Key(),
		text:                event.Text(),
		matchEnter:          event.MatchString("enter"),
		matchShiftEnter:     event.MatchString("shift+enter"),
		matchCtrlEnter:      event.MatchString("ctrl+enter"),
		matchCtrlShiftEnter: event.MatchString("ctrl+shift+enter"),
	}
	a.entries = append(a.entries, entry)
	if len(a.entries) > maxEntries {
		a.entries = a.entries[len(a.entries)-maxEntries:]
	}
	return true
}

func (a *KeyEventDebug) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	var lastWidgets []t.Widget
	if len(a.entries) == 0 {
		lastWidgets = []t.Widget{
			t.Text{
				Content: "No key events yet.",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		}
	} else {
		last := a.entries[len(a.entries)-1]
		lastWidgets = []t.Widget{
			t.Text{Content: fmt.Sprintf("Key: %q", last.key)},
			t.Text{Content: fmt.Sprintf("Text: %q", last.text)},
			t.Text{
				Content: fmt.Sprintf("Match: enter=%t shift+enter=%t ctrl+enter=%t ctrl+shift+enter=%t",
					last.matchEnter, last.matchShiftEnter, last.matchCtrlEnter, last.matchCtrlShiftEnter),
				Wrap: t.WrapSoft,
			},
		}
	}

	lastSection := t.Column{
		Style: t.Style{
			Border:  t.RoundedBorder(theme.Border, t.BorderTitle("Last Event")),
			Padding: t.EdgeInsetsAll(1),
		},
		Children: lastWidgets,
	}

	var historyWidgets []t.Widget
	if len(a.entries) == 0 {
		historyWidgets = []t.Widget{
			t.Text{
				Content: "History is empty.",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		}
	} else {
		for i := len(a.entries) - 1; i >= 0; i-- {
			entry := a.entries[i]
			line := fmt.Sprintf("%s key=%q text=%q enter=%t shift+enter=%t ctrl+enter=%t",
				entry.at.Format("15:04:05.000"),
				entry.key,
				entry.text,
				entry.matchEnter,
				entry.matchShiftEnter,
				entry.matchCtrlEnter,
			)
			historyWidgets = append(historyWidgets, t.Text{
				Content: line,
				Wrap:    t.WrapSoft,
			})
		}
	}

	historySection := t.Column{
		Height: t.Flex(1),
		Style: t.Style{
			Border:  t.RoundedBorder(theme.Border, t.BorderTitle("History (latest first)")),
			Padding: t.EdgeInsetsAll(1),
		},
		Children: historyWidgets,
	}

	return t.Dock{
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsets{Left: 1, Right: 1},
				},
			},
		},
		Body: t.Column{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style: t.Style{
				BackgroundColor: theme.Background,
				Padding:         t.EdgeInsetsAll(1),
			},
			Children: []t.Widget{
				t.Text{
					Content: "Key Event Debugger (UV)",
					Style: t.Style{
						ForegroundColor: theme.Text,
						Bold:            true,
					},
				},
				t.Text{
					Content: "Press keys to inspect UV key strings. Shift+Enter should show key=\"shift+enter\" and match shift+enter=true when Kitty keyboard is active.",
					Style:   t.Style{ForegroundColor: theme.TextMuted},
					Wrap:    t.WrapSoft,
				},
				lastSection,
				historySection,
			},
		},
	}
}

func main() {
	app := &KeyEventDebug{}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
