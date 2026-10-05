// Command hyperlink-demo runs the Hyperlinks demo on its own. It also appears
// in the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/hyperlinkdemo"
)

func main() {
	demokit.Run(hyperlinkdemo.New())
}
