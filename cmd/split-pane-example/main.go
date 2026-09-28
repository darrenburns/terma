// Command split-pane-example runs the SplitPane demo on its own. It also appears
// in the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/splitpanedemo"
)

func main() {
	demokit.Run(splitpanedemo.New())
}
