// Command progressbar-example runs the Progress & Spinners demo on its own. It
// also appears in the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/progressdemo"
)

func main() {
	demokit.Run(progressdemo.New())
}
