// Command toast-example runs the Toasts demo on its own. It also appears in
// the cmd/terma-demos gallery.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/toastdemo"
)

func main() {
	demokit.Run(toastdemo.New())
}
