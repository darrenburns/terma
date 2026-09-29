package terma

import "sync"

// terminalOutput holds raw sequences waiting to be written to the terminal.
// The renderer owns the terminal while an app runs, so writes are queued and
// the event loop writes them just before the next frame is flushed.
var terminalOutput struct {
	mu      sync.Mutex
	pending []string
}

// WriteTerminal queues a raw escape sequence (an OSC, a mode change, a bell)
// to be written to the terminal between frames. It is safe to call from any
// goroutine. Nothing is written if no app is running.
func WriteTerminal(seq string) {
	if seq == "" || currentAppContext() == nil {
		return
	}
	terminalOutput.mu.Lock()
	terminalOutput.pending = append(terminalOutput.pending, seq)
	terminalOutput.mu.Unlock()
	scheduleRender()
}

// takeTerminalWrites returns and clears the queued terminal writes.
func takeTerminalWrites() []string {
	terminalOutput.mu.Lock()
	defer terminalOutput.mu.Unlock()
	pending := terminalOutput.pending
	terminalOutput.pending = nil
	return pending
}
