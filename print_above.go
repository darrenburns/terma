package terma

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// printAboveItem is one PrintAbove or PrintAboveText call, waiting for the
// next frame of an inline app.
type printAboveItem struct {
	widget Widget
	text   string
}

// printAboveQueue holds what is waiting to be printed above the live region.
// open is set while an inline app runs and will print what is queued.
var printAboveQueue struct {
	mu      sync.Mutex
	open    bool
	pending []printAboveItem
}

// printAboveMaxRows bounds how tall a widget printed above can be. A widget
// is laid out in the terminal's height first, so short output stays cheap,
// and laid out again in this many rows only if it filled that height.
const printAboveMaxRows = 1000

// PrintAbove renders w at the terminal's width and writes it above the live
// region of an inline app (see RunInline), where it stays in the terminal's
// scrollback. w is laid out with its height loose: Auto heights follow its
// content (up to 1000 rows), so avoid Flex heights.
//
// It is safe to call from any goroutine; w is printed with the next frame.
// With no app running it prints w to stdout, like Print. A fullscreen app
// (Run) has nothing above it, so there it does nothing.
func PrintAbove(w Widget) {
	queuePrintAbove(printAboveItem{widget: w}, func() { _ = Print(w) })
}

// PrintAboveText writes text above the live region of an inline app, like
// PrintAbove. text may contain ANSI styling. Each line takes one row: lines
// wider than the terminal are cut off (print a wrapping Text with PrintAbove
// to wrap them), and a single trailing newline is ignored.
func PrintAboveText(text string) {
	text = strings.TrimSuffix(text, "\n")
	queuePrintAbove(printAboveItem{text: text}, func() { fmt.Fprintln(os.Stdout, text) })
}

func queuePrintAbove(item printAboveItem, direct func()) {
	printAboveQueue.mu.Lock()
	if printAboveQueue.open {
		printAboveQueue.pending = append(printAboveQueue.pending, item)
		printAboveQueue.mu.Unlock()
		scheduleRender()
		return
	}
	printAboveQueue.mu.Unlock()
	if currentAppContext() != nil {
		Log("PrintAbove ignored: the app isn't running inline")
		return
	}
	direct()
}

// openPrintAbove starts queueing printed output for an inline app.
func openPrintAbove() {
	printAboveQueue.mu.Lock()
	printAboveQueue.open = true
	printAboveQueue.mu.Unlock()
}

// takePrintAbove returns and clears what is queued. With closing set, nothing
// more is queued afterwards.
func takePrintAbove(closing bool) []printAboveItem {
	printAboveQueue.mu.Lock()
	defer printAboveQueue.mu.Unlock()
	pending := printAboveQueue.pending
	printAboveQueue.pending = nil
	if closing {
		printAboveQueue.open = false
	}
	return pending
}

// lines renders the item as the rows it takes in a terminal width cells wide
// and rows high.
func (p printAboveItem) lines(width, rows int) []string {
	if p.widget == nil {
		return strings.Split(p.text, "\n")
	}
	rows = max(rows, 1)
	buf, layoutWidth, layoutHeight := RenderToBufferWithSize(p.widget, width, rows)
	if layoutHeight >= rows && rows < printAboveMaxRows {
		buf, layoutWidth, layoutHeight = RenderToBufferWithSize(p.widget, width, printAboveMaxRows)
	}
	if layoutHeight == 0 {
		return nil
	}
	return strings.Split(BufferToANSI(buf, layoutWidth, layoutHeight), "\n")
}
