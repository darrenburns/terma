package main

import (
	"log"

	t "terma"
)

type App struct {
	state *t.TreeState[string]
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	return t.Tree[string]{State: a.state}
}

func main() {
	app := &App{
		state: t.NewTreeState([]t.TreeNode[string]{
			{Data: "Fruits", Children: []t.TreeNode[string]{
				{Data: "Apple", Children: []t.TreeNode[string]{}},
				{Data: "Banana", Children: []t.TreeNode[string]{}},
			}},
			{Data: "Vegetables", Children: []t.TreeNode[string]{
				{Data: "Carrot", Children: []t.TreeNode[string]{}},
			}},
		}),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
