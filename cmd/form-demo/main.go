// Command form-demo exercises form validation and input state.
package main

import (
	"flag"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/formdemo"
)

type modalDemo struct{ demokit.Demo }

func (d modalDemo) Build(ctx t.BuildContext) t.Widget {
	return t.Dialog{ID: "form-demo-dialog", Visible: true, Title: "Form in a dialog", Content: d.Demo, Style: t.Style{Width: t.Percent(90), Height: t.Percent(90)}}
}
func main() {
	dialog := flag.Bool("dialog", false, "show the form inside a modal dialog")
	flag.Parse()
	demo := formdemo.New()
	if *dialog {
		demo = modalDemo{demo}
	}
	demokit.Run(demo)
}
