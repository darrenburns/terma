// Command menu-example runs the Menu demo on its own. It also appears in the
// cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/menudemo"
)

func main() {
	demokit.Run(menudemo.New())
}
