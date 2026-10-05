// Command terma-demos brings every widget demo together in one app. It opens
// on a home page listing the demos; press ctrl+g anywhere to switch demo.
package main

import (
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/darrenburns/terma/cmd/internal/demos/checkboxdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/dialogdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/filepickerdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/formdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/jumpdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/listdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/markdowndemo"
	"github.com/darrenburns/terma/cmd/internal/demos/menudemo"
	"github.com/darrenburns/terma/cmd/internal/demos/palettedemo"
	"github.com/darrenburns/terma/cmd/internal/demos/progressdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/scrolldemo"
	"github.com/darrenburns/terma/cmd/internal/demos/selectdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/shimmerdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/splitpanedemo"
	"github.com/darrenburns/terma/cmd/internal/demos/switcherdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/tabledemo"
	"github.com/darrenburns/terma/cmd/internal/demos/tablefeaturesdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/tabsdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/textareademo"
	"github.com/darrenburns/terma/cmd/internal/demos/textinputdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/toastdemo"
	"github.com/darrenburns/terma/cmd/internal/demos/treedemo"
)

// demos lists the gallery's demos in the order they're shown.
var demos = []Entry{
	{Info: selectdemo.Info, Command: "./cmd/select-demo", New: selectdemo.New},
	{Info: formdemo.Info, Command: "./cmd/form-demo", New: formdemo.New},
	{Info: markdowndemo.Info, Command: "./cmd/markdown-demo", New: markdowndemo.New},
	{Info: tablefeaturesdemo.Info, Command: "./cmd/table-features-demo", New: tablefeaturesdemo.New},
	{Info: filepickerdemo.Info, Command: "./cmd/filepicker-demo", New: filepickerdemo.New},

	{Info: listdemo.Info, Command: "./cmd/list-demo", New: listdemo.New},
	{Info: tabledemo.Info, Command: "./cmd/table-demo", New: tabledemo.New},
	{Info: treedemo.Info, Command: "./cmd/tree-example", New: treedemo.New},
	{Info: textinputdemo.Info, Command: "./cmd/text-input-example", New: textinputdemo.New},
	{Info: textareademo.Info, Command: "./cmd/text-area-demo", New: textareademo.New},
	{Info: checkboxdemo.Info, Command: "./cmd/checkbox-demo", New: checkboxdemo.New},
	{Info: tabsdemo.Info, Command: "./cmd/tab-example", New: tabsdemo.New},
	{Info: switcherdemo.Info, Command: "./cmd/switcher-example", New: switcherdemo.New},
	{Info: scrolldemo.Info, Command: "./cmd/scroll-example", New: scrolldemo.New},
	{Info: splitpanedemo.Info, Command: "./cmd/split-pane-example", New: splitpanedemo.New},
	{Info: menudemo.Info, Command: "./cmd/menu-example", New: menudemo.New},
	{Info: palettedemo.Info, Command: "./cmd/command-palette-example", New: palettedemo.New},
	{Info: dialogdemo.Info, Command: "./cmd/dialog-example", New: dialogdemo.New},
	{Info: toastdemo.Info, Command: "./cmd/toast-example", New: toastdemo.New},
	{Info: progressdemo.Info, Command: "./cmd/progressbar-example", New: progressdemo.New},
	{Info: shimmerdemo.Info, Command: "./cmd/shimmer-example", New: shimmerdemo.New},
	{Info: jumpdemo.Info, Command: "./cmd/jump-example", New: jumpdemo.New},
}

func main() {
	demokit.SwitchHint = "[b $Accent]" + switchKey + "[/] [$TextMuted]demos[/]"
	demokit.Run(NewGallery(demos))
}
