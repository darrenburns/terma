package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/docs/widget-examples/collections"
	"github.com/darrenburns/terma/docs/widget-examples/demo"
	"github.com/darrenburns/terma/docs/widget-examples/display"
	"github.com/darrenburns/terma/docs/widget-examples/inputs"
	"github.com/darrenburns/terma/docs/widget-examples/layout"
)

func examples() []demo.Example {
	all := append(display.Examples(), inputs.Examples()...)
	all = append(all, collections.Examples()...)
	all = append(all, layout.Examples()...)
	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	return all
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name := flag.String("widget", "text", "example to run")
	list := flag.Bool("list", false, "list examples")
	render := flag.Bool("render", false, "write SVG images")
	check := flag.Bool("check", false, "compare SVG images with current rendering")
	assets := flag.String("assets", "docs/assets/widgets", "SVG directory")
	flag.Parse()
	all := examples()
	if *list {
		for _, example := range all {
			fmt.Println(example.Name)
		}
		return nil
	}
	if *render || *check {
		if *render && *check {
			return fmt.Errorf("choose -render or -check")
		}
		if *render {
			if err := os.MkdirAll(*assets, 0755); err != nil {
				return err
			}
		}
		for _, example := range all {
			svg := []byte(t.Snapshot(example.Widget(), example.Width, example.Height))
			path := filepath.Join(*assets, example.Name+".svg")
			if *check {
				existing, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if !bytes.Equal(existing, svg) {
					return fmt.Errorf("%s differs; regenerate with -render", path)
				}
			} else if err := os.WriteFile(path, svg, 0644); err != nil {
				return err
			}
			fmt.Println(path)
		}
		return nil
	}
	for _, example := range all {
		if example.Name == *name {
			return t.Run(example.Widget())
		}
	}
	return fmt.Errorf("unknown widget %q; use -list", *name)
}
