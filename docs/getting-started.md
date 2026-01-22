# Getting Started

This guide will walk you through installing Terma and building your first terminal UI application.

## Chapter 1: Hello World

Welcome to Terma! In this first chapter you'll build the smallest possible app: a single line of text rendered in the terminal.

Concepts: Widget interface, Build pattern, Run()

Build: Static text display

### Define your app

Terma apps are widgets. If your type implements `Build(ctx BuildContext) Widget`, it satisfies the `Widget` interface and can be rendered as the root of your UI.

```go
type App struct{}

func (a *App) Build(ctx BuildContext) Widget {
    return Text{Content: "Hello, Terminal!"}
}

func main() { Run(&App{}) }
```

In `Build`, you return a widget tree that describes the UI for the current state. Here it's just a leaf widget (`Text`) with static content.

### Run it

`Run()` starts Terma's render loop and input handling, then blocks until the app exits. If you want a complete file you can run immediately:

```go
package main

import t "terma"

type App struct{}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.Text{Content: "Hello, Terminal!"}
}

func main() {
    t.Run(&App{})
}
```

Then run:

```bash
go run .
```

Teaches:
- The Widget interface and Build() method
- How Terma apps are structured
- Running with Run()
