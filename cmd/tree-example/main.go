// Command tree-example runs the Tree demo on its own. It also appears in the
// cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/treedemo"
)

func main() {
	demokit.Run(treedemo.New())
}
