// Command dialog-example runs the Dialog demo on its own. It also appears in the
// cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/dialogdemo"
)

func main() {
	demokit.Run(dialogdemo.New())
}
