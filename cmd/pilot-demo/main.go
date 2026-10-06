// Command pilot-demo runs the Pilot demo on its own: a sign-up app next to the
// test that drives it. It also appears in the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/pilotdemo"
)

func main() {
	demokit.Run(pilotdemo.New())
}
