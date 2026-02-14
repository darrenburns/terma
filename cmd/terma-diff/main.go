package main

import (
	"flag"
	"log"
	"os"

	t "terma"
)

func main() {
	var staged bool
	flag.BoolVar(&staged, "staged", false, "show staged diff (git diff --staged)")
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	provider := GitDiffProvider{WorkDir: cwd}
	app := NewDiffApp(provider, staged)
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
