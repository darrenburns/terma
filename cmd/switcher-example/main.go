// Command switcher-example runs the Switcher demo on its own. It also appears
// in the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/switcherdemo"
)

func main() {
	demokit.Run(switcherdemo.New())
}
