// Command table-demo runs the Table demo on its own. It also appears in the
// cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/tabledemo"
)

func main() {
	demokit.Run(tabledemo.New())
}
