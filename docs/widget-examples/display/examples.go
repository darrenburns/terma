package display

import (
	"fmt"
	"image"
	"image/color"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/docs/widget-examples/demo"
)

// --8<-- [start:text]
func Text() t.Widget {
	return t.Column{Spacing: 1, Children: []t.Widget{
		t.Text{Spans: []t.Span{
			t.BoldSpan("Deploy complete"),
			t.PlainSpan("  v0.4.0"),
		}},
		t.Text{
			Content: "Your terminal UI starts with a widget.",
			Wrap:    t.WrapSoft,
			Style:   t.Style{Width: t.Cells(28)},
		},
	}}
}

// --8<-- [end:text]

// --8<-- [start:presentedtext]
func PresentedText() t.Widget {
	count := t.NewSignal(0)
	label := t.SignalText(count, func(n int) string {
		return fmt.Sprintf("Count: %d", n)
	})
	label.LayoutStyle.Width = t.Auto
	return t.Column{Spacing: 1, Children: []t.Widget{
		label,
		t.Button{ID: "increment", Label: "Add one", OnPress: func() {
			count.Set(count.Peek() + 1)
		}},
	}}
}

// --8<-- [end:presentedtext]

// --8<-- [start:progressbar]
func ProgressBar() t.Widget {
	return t.Column{Spacing: 1, Children: []t.Widget{
		t.Text{Content: "Uploading 65%"},
		t.ProgressBar{
			Progress: 0.65,
			Style:    t.Style{Width: t.Cells(32)},
		},
	}}
}

// --8<-- [end:progressbar]

// --8<-- [start:sparkline]
func Sparkline() t.Widget {
	return t.Column{Spacing: 1, Children: []t.Widget{
		t.Text{Content: "Requests per minute"},
		t.Sparkline{
			Values:       []float64{2, 4, 3, 8, 5, 12, 9, 15, 11, 6, 8, 14},
			ColorByValue: true,
			Style:        t.Style{Width: t.Cells(32)},
		},
	}}
}

// --8<-- [end:sparkline]

// --8<-- [start:spinner]
func Spinner() t.Widget {
	state := t.NewSpinnerState(t.SpinnerDots)
	state.Start()
	return t.Row{Spacing: 1, Children: []t.Widget{
		t.Spinner{State: state},
		t.Text{Content: "Fetching updates"},
	}}
}

// --8<-- [end:spinner]

// --8<-- [start:toasts]
func Toasts() t.Widget {
	toasts := t.NewToastState(t.ToastOptions{})
	toasts.Notify(t.Toast{Title: "Saved", Message: "Wrote notes.md", Severity: t.ToastSuccess})
	toasts.Error("Copy failed")
	return t.Column{Children: []t.Widget{
		t.Button{ID: "save", Label: "Save", OnPress: func() { toasts.Success("Saved") }},
		t.Toasts{State: toasts, Width: 28},
	}}
}

// --8<-- [end:toasts]

// --8<-- [start:image]
func Image() t.Widget {
	pixels := image.NewNRGBA(image.Rect(0, 0, 64, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			pixels.SetNRGBA(x, y, color.NRGBA{uint8(x * 4), uint8(y * 8), 180, 255})
		}
	}
	source, err := t.NewImageResource(pixels)
	if err != nil {
		panic(err)
	}
	return t.Image{
		Source: source,
		Style:  t.Style{Width: t.Cells(32), Height: t.Cells(8)},
	}
}

// --8<-- [end:image]

// --8<-- [start:custom-widgets]
type Greeting struct {
	Name string
}

func (g Greeting) Build(ctx t.BuildContext) t.Widget {
	return t.Text{Content: "Hello, " + g.Name + "!"}
}

func CustomWidget() t.Widget {
	return Greeting{Name: "Terma"}
}

// --8<-- [end:custom-widgets]

func Examples() []demo.Example {
	return []demo.Example{
		{Name: "text", Widget: Text, Width: 38, Height: 5},
		{Name: "presentedtext", Widget: PresentedText, Width: 32, Height: 5},
		{Name: "progressbar", Widget: ProgressBar, Width: 34, Height: 4},
		{Name: "sparkline", Widget: Sparkline, Width: 34, Height: 4},
		{Name: "spinner", Widget: Spinner, Width: 24, Height: 2},
		{Name: "toasts", Widget: Toasts, Width: 40, Height: 10},
		{Name: "image", Widget: Image, Width: 32, Height: 8},
		{Name: "custom-widgets", Widget: CustomWidget, Width: 20, Height: 2},
	}
}
