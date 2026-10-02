package main

import (
	"log"

	t "github.com/darrenburns/terma"
)

func main() {
	if err := t.Run(t.Text{Content: "Hello, Terma!"}); err != nil {
		log.Fatal(err)
	}
}
