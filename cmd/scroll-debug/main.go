package main

import (
	"log"

	t "github.com/darrenburns/terma"
)


type ScrollDebug struct {
	scrollState *t.ScrollState
}

func (s *ScrollDebug) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		ID:     "root",
		Height: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			t.Text{Content: "Scroll Debug - Press j/k to scroll, G to go to bottom"},
			&t.Scrollable{
				ID:     "scroller",
				State:  s.scrollState,
				Height: t.Cells(10),
				Width:  t.Flex(1),
				Style: t.Style{
					Border:  t.RoundedBorder(theme.Primary, t.BorderTitle("Wrapped Text")),
					Padding: t.EdgeInsetsAll(1),
				},
				Child: t.Text{
					Content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. END OF TEXT",
					Width:   t.Flex(1),
				},
			},
		},
	}
}

func main() {
	app := &ScrollDebug{
		scrollState: t.NewScrollState(),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
