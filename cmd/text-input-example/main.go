// Command text-input-example runs the TextInput demo on its own. It also
// appears in the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/textinputdemo"
)

func main() {
	demokit.Run(textinputdemo.New())
}
