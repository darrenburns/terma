// Command checkbox-demo runs the Checkbox demo on its own. It also appears in
// the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/checkboxdemo"
)

func main() {
	demokit.Run(checkboxdemo.New())
}
