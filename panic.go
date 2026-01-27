package terma

import (
	"fmt"
	"os"
	"sync"
)

// Panic represents a framework-level panic with structured information.
type Panic struct {
	Message    string
	StackTrace string
}

var (
	panicStore   []Panic
	panicStoreMu sync.Mutex
)

// recordPanic stores a panic for later rendering after terminal shutdown.
func recordPanic(p Panic) {
	panicStoreMu.Lock()
	defer panicStoreMu.Unlock()
	panicStore = append(panicStore, p)
}

// drainPanics returns all recorded panics and clears the store.
func drainPanics() []Panic {
	panicStoreMu.Lock()
	defer panicStoreMu.Unlock()
	panics := panicStore
	panicStore = nil
	return panics
}

// renderPanics prints recorded panics to stderr. Called after the terminal
// has been restored to normal mode so output is visible.
func renderPanics() {
	for _, p := range drainPanics() {
		if p.StackTrace != "" {
			fmt.Fprintf(os.Stderr, "%s\n", p.StackTrace)
		}
		fmt.Fprintf(os.Stderr, "terma: panic: %s\n", p.Message)
	}
}
