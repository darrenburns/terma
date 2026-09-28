// Command text-area-demo runs the TextArea demo on its own. It also appears
// in the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/textareademo"
)

func main() {
	demokit.Run(textareademo.New())
}
