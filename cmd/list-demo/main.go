// Command list-demo runs the List demo on its own. It also appears in the
// cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/listdemo"
)

func main() {
	demokit.Run(listdemo.New())
}
