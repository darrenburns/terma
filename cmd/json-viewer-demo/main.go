// Command json-viewer-demo runs the JSONViewer demo on its own.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/jsonviewerdemo"
)

func main() {
	demokit.Run(jsonviewerdemo.New())
}
