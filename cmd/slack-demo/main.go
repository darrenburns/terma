package main

import (
	"fmt"
	"log"
	"time"

	t "github.com/darrenburns/terma"
)

type Channel struct {
	Name   string
	Unread int
}

type Message struct {
	Author    string
	Content   string
	Timestamp time.Time
}

type SlackApp struct {
	channels      *t.ListState[Channel]
	messages      t.AnySignal[[]Message]
	inputState    *t.TextInputState
	scrollState   *t.ScrollState
	activeChannel t.Signal[string]
}

func (a *SlackApp) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	activeChannel := a.activeChannel.Get()
	messages := a.messages.Get()

	return t.Row{
		Height: t.Flex(1),
		Children: []t.Widget{
			// Sidebar with gradient background
			a.buildSidebar(theme),
			// Main content area
			a.buildMainContent(theme, activeChannel, messages),
		},
	}
}

func (a *SlackApp) buildSidebar(theme t.ThemeData) t.Widget {
	return t.Column{
		Width: t.Cells(28),
		Style: t.Style{
			BackgroundColor: t.NewGradient(
				t.Hex("#4c1d95"), // Deep violet
				t.Hex("#1e1b4b"), // Dark indigo
			).WithAngle(180),
			Padding: t.EdgeInsets{Top: 1, Left: 1, Right: 1},
		},
		Children: []t.Widget{
			// Workspace header
			t.Text{
				Content: "Terma HQ",
				Style: t.Style{
					Bold:            true,
					ForegroundColor: t.White,
					Padding:         t.EdgeInsets{Bottom: 1},
				},
			},
			t.Text{
				Content: "Channels",
				Style: t.Style{
					ForegroundColor: t.Hex("#a5b4fc"),
					Padding:         t.EdgeInsets{Bottom: 1},
				},
			},
			// Channel list
			t.List[Channel]{
				ID:    "channel-list",
				State: a.channels,
				OnSelect: func(ch Channel) {
					a.activeChannel.Set(ch.Name)
				},
				RenderItem: func(ch Channel, active bool, selected bool) t.Widget {
					bg := t.Hex("#000000").WithAlpha(0)
					if active {
						bg = t.Hex("#6366f1").WithAlpha(0.4)
					}

					content := "# " + ch.Name
					if ch.Unread > 0 {
						content = fmt.Sprintf("# %s (%d)", ch.Name, ch.Unread)
					}

					return t.Text{
						Content: content,
						Style: t.Style{
							ForegroundColor: t.White,
							BackgroundColor: bg,
							Padding:         t.EdgeInsets{Left: 1, Right: 1},
						},
					}
				},
			},
		},
	}
}

func (a *SlackApp) buildMainContent(theme t.ThemeData, activeChannel string, messages []Message) t.Widget {
	return t.Column{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Children: []t.Widget{
			// Channel header
			t.Row{
				Height: t.Cells(3),
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsets{Left: 2, Top: 1},
				},
				Children: []t.Widget{
					t.Text{
						Content: "# " + activeChannel,
						Style: t.Style{
							Bold:            true,
							ForegroundColor: theme.Text,
						},
					},
				},
			},
			// Messages area
			t.Scrollable{
				State:  a.scrollState,
				Height: t.Flex(1),
				Child: t.Column{
					Style: t.Style{
						Padding: t.EdgeInsets{Left: 2, Right: 2, Top: 1},
					},
					Children: a.renderMessages(theme, messages),
				},
			},
			// Input area
			t.Row{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsXY(2, 1),
				},
				Children: []t.Widget{
					t.TextInput{
						ID:          "message-input",
						State:       a.inputState,
						Placeholder: fmt.Sprintf("Message #%s", activeChannel),
						Width:       t.Flex(1),
						Style: t.Style{
							BackgroundColor: theme.Background,
							ForegroundColor: theme.Text,
							Padding:         t.EdgeInsets{Left: 1, Right: 1},
						},
						OnSubmit: func(text string) {
							if text != "" {
								a.sendMessage(text)
							}
						},
					},
				},
			},
		},
	}
}

func (a *SlackApp) renderMessages(theme t.ThemeData, messages []Message) []t.Widget {
	widgets := make([]t.Widget, 0, len(messages))

	for _, msg := range messages {
		widgets = append(widgets, t.Column{
			Style: t.Style{
				Padding: t.EdgeInsets{Bottom: 1},
			},
			Children: []t.Widget{
				t.Row{
					Spacing: 2,
					Children: []t.Widget{
						t.Text{
							Content: msg.Author,
							Style: t.Style{
								Bold:            true,
								ForegroundColor: theme.Primary,
							},
						},
						t.Text{
							Content: msg.Timestamp.Format("3:04 PM"),
							Style: t.Style{
								ForegroundColor: theme.TextMuted,
							},
						},
					},
				},
				t.Text{
					Content: msg.Content,
					Wrap:    t.WrapSoft,
					Style: t.Style{
						ForegroundColor: theme.Text,
					},
				},
			},
		})
	}

	return widgets
}

func (a *SlackApp) sendMessage(text string) {
	a.messages.Update(func(msgs []Message) []Message {
		return append(msgs, Message{
			Author:    "You",
			Content:   text,
			Timestamp: time.Now(),
		})
	})
	a.inputState.SetText("")
	// Scroll to bottom (large value gets clamped to max)
	a.scrollState.SetOffset(999999)
}

func (a *SlackApp) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "ctrl+q", Name: "Quit", Action: func() { t.Quit() }},
	}
}

func main() {
	_ = t.InitLogger()

	channels := []Channel{
		{Name: "general", Unread: 3},
		{Name: "random", Unread: 0},
		{Name: "engineering", Unread: 12},
		{Name: "design", Unread: 0},
		{Name: "announcements", Unread: 1},
	}

	messages := []Message{
		{Author: "Alice", Content: "Hey everyone! Welcome to the new Terma chat.", Timestamp: time.Now().Add(-2 * time.Hour)},
		{Author: "Bob", Content: "This is looking great! Love the gradient sidebar.", Timestamp: time.Now().Add(-1 * time.Hour)},
		{Author: "Charlie", Content: "The reactive signals make state management so clean.", Timestamp: time.Now().Add(-30 * time.Minute)},
		{Author: "Alice", Content: "Agreed! And the layout system is really intuitive.", Timestamp: time.Now().Add(-15 * time.Minute)},
		{Author: "Bob", Content: "Has anyone tried the List widget with multi-select yet?", Timestamp: time.Now().Add(-5 * time.Minute)},
	}

	scrollState := t.NewScrollState()
	scrollState.PinToBottom = true
	scrollState.ScrollToBottom()
	app := &SlackApp{
		channels:      t.NewListState(channels),
		messages:      t.NewAnySignal(messages),
		inputState:    t.NewTextInputState(""),
		scrollState:   scrollState,
		activeChannel: t.NewSignal("general"),
	}

	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
