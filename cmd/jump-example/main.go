// Command jump-example runs the jump mode demo on its own. It also appears in
// the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/jumpdemo"
)

func main() {
	demokit.Run(jumpdemo.New())
}
