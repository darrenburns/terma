// Command scroll-example runs the Scroll demo on its own. It also appears in
// the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/scrolldemo"
)

func main() {
	demokit.Run(scrolldemo.New())
}
