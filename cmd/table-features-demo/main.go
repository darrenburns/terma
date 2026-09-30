// Command table-features-demo runs the same demo used by terma-demos.
package main

import (
	"flag"

	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/tablefeaturesdemo"
)

func main() {
	probe := flag.Bool("probe", false, "run the focused widget verification demo")
	flag.Parse()
	var demo demokit.Demo
	if *probe {
		demo = tablefeaturesdemo.NewProbe()
	} else {
		demo = tablefeaturesdemo.New()
	}
	demokit.Run(demo)
}
