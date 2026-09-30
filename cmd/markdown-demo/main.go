// Command markdown-demo runs the same demo used by terma-demos.
package main

import (
	"flag"

	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/markdowndemo"
)

func main() {
	probe := flag.Bool("probe", false, "run the focused widget verification demo")
	flag.Parse()
	var demo demokit.Demo
	if *probe {
		demo = markdowndemo.NewProbe()
	} else {
		demo = markdowndemo.New()
	}
	demokit.Run(demo)
}
