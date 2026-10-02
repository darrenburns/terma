package demo

import t "github.com/darrenburns/terma"

type Example struct {
	Name          string
	Widget        func() t.Widget
	Width, Height int
}
