// Command shimmer-example runs the Shimmer demo on its own. It also appears in
// the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/shimmerdemo"
)

func main() {
	demokit.Run(shimmerdemo.New())
}
