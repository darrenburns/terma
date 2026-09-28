// Command tab-example runs the Tabs demo on its own. It also appears in the
// cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/tabsdemo"
)

func main() {
	demokit.Run(tabsdemo.New())
}
