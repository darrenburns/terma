// Command command-palette-example runs the Command Palette demo on its own. It
// also appears in the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/palettedemo"
)

func main() {
	demokit.Run(palettedemo.New())
}
