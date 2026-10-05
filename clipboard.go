package terma

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf16"

	"github.com/charmbracelet/x/ansi"
)

// ClipboardSelection identifies which terminal clipboard selection to target.
//
// Common values are [SystemClipboard] and [PrimaryClipboard].
type ClipboardSelection = byte

const (
	// SystemClipboard targets the terminal's system clipboard.
	SystemClipboard ClipboardSelection = ansi.SystemClipboard
	// PrimaryClipboard targets the terminal's primary selection.
	PrimaryClipboard ClipboardSelection = ansi.PrimaryClipboard
)

// ClipboardMethod chooses how [SetClipboard] and [ReadClipboard] reach the
// clipboard.
type ClipboardMethod int32

const (
	// ClipboardAuto uses the system clipboard tool in a local session, then
	// the tmux buffer, then OSC 52. It is the default.
	ClipboardAuto ClipboardMethod = iota
	// ClipboardNative uses only the system clipboard tool and the tmux
	// buffer, and never OSC 52.
	ClipboardNative
	// ClipboardTerminal uses only OSC 52, wrapped for tmux or screen when
	// running inside one.
	ClipboardTerminal
)

var clipboardMethod atomic.Int32

// SetClipboardMethod chooses how [SetClipboard] and [ReadClipboard] reach the
// clipboard, for apps that let users override [ClipboardAuto].
func SetClipboardMethod(m ClipboardMethod) {
	clipboardMethod.Store(int32(m))
}

// SetClipboard copies content to the selected clipboard. It is safe to call
// from any goroutine and returns without waiting; copies apply in call order.
//
// In a local session (not over SSH) with a clipboard tool for the platform,
// the tool gets the content: pbcopy on macOS, clip.exe on Windows and WSL,
// wl-copy under Wayland, and xclip or xsel under X11. If no tool applies or
// it fails, SetClipboard falls back to the terminal: inside tmux it loads a
// tmux buffer with `tmux load-buffer -w`, which tmux 3.2+ forwards to the
// outer terminal's clipboard, and also sends OSC 52 through tmux passthrough;
// inside screen it sends OSC 52 through screen passthrough; elsewhere it
// sends OSC 52 directly. OSC 52 is written between frames and needs a running
// app; terminals without OSC 52 support ignore it.
//
// macOS and Windows have no primary selection, so [PrimaryClipboard] goes
// straight to the terminal there. [SetClipboardMethod] can restrict the
// order to the tool or to the terminal. Failures are logged with [Log].
func SetClipboard(selection ClipboardSelection, content string) {
	sys := clipboardSys()
	plan := planClipboardWrite(sys, ClipboardMethod(clipboardMethod.Load()), selection, content)
	queueClipboardJob(plan.native == nil && plan.tmux == nil, func() { plan.apply(sys.run) })
}

// clipboardReads holds ReadClipboard callbacks waiting for the terminal's
// reply, oldest first. Terminals answer OSC 52 queries in order.
var clipboardReads struct {
	mu      sync.Mutex
	pending []func(string)
}

// ReadClipboard reads the selected clipboard and calls fn with its contents
// on the event loop. It returns without waiting and does nothing unless an
// app is running.
//
// In a local session with a paste tool (pbpaste, wl-paste, xclip, xsel, or
// PowerShell's Get-Clipboard on Windows and WSL), the tool's output is
// delivered. Otherwise ReadClipboard asks the terminal with an OSC 52 query,
// wrapped for tmux or screen. Reading through the terminal needs OSC 52 read
// support, which many terminals leave off or ask the user to allow, so fn may
// never be called. Don't block on it.
func ReadClipboard(selection ClipboardSelection, fn func(content string)) {
	if fn == nil || currentAppContext() == nil {
		return
	}
	sys := clipboardSys()
	plan := planClipboardRead(sys, ClipboardMethod(clipboardMethod.Load()), selection)
	queueClipboardJob(plan.native == nil, func() { plan.apply(sys.run, fn) })
}

// deliverClipboard passes a clipboard reply to the oldest waiting reader.
// It reports whether a reader was waiting.
func deliverClipboard(content string) bool {
	clipboardReads.mu.Lock()
	if len(clipboardReads.pending) == 0 {
		clipboardReads.mu.Unlock()
		return false
	}
	fn := clipboardReads.pending[0]
	clipboardReads.pending = clipboardReads.pending[1:]
	clipboardReads.mu.Unlock()
	fn(content)
	return true
}

// resetClipboardReads drops readers left waiting by a previous app run.
func resetClipboardReads() {
	clipboardReads.mu.Lock()
	clipboardReads.pending = nil
	clipboardReads.mu.Unlock()
}

// clipboardCommand is one run of a clipboard tool.
type clipboardCommand struct {
	argv  []string
	stdin []byte
}

// clipboardWritePlan is what SetClipboard does, in order: run native, and
// stop if it succeeds; run tmux, ignoring failure; write terminal.
type clipboardWritePlan struct {
	native   *clipboardCommand
	tmux     *clipboardCommand
	terminal string
}

// clipboardReadPlan is what ReadClipboard does: run native and deliver its
// output, or send the terminal query if it is missing or fails.
type clipboardReadPlan struct {
	native   *clipboardCommand
	trimCRLF bool
	terminal string
}

func (p clipboardWritePlan) apply(run clipboardRunner) {
	if p.native != nil {
		_, err := run(*p.native, false)
		if err == nil {
			return
		}
		Log("Clipboard: %s failed: %v", p.native.argv[0], err)
	}
	if p.tmux != nil {
		if _, err := run(*p.tmux, false); err != nil {
			Log("Clipboard: tmux load-buffer failed: %v", err)
		}
	}
	if p.tmux == nil && p.terminal == "" {
		Log("Clipboard: copy dropped; no clipboard method applies")
	}
	WriteTerminal(p.terminal)
}

func (p clipboardReadPlan) apply(run clipboardRunner, fn func(string)) {
	if p.native != nil {
		out, err := run(*p.native, true)
		if err == nil {
			content := string(out)
			if p.trimCRLF {
				content = strings.TrimSuffix(content, "\r\n")
			}
			Dispatch(func() { fn(content) })
			return
		}
		Log("Clipboard: %s failed: %v", p.native.argv[0], err)
	}
	if p.terminal == "" || currentAppContext() == nil {
		return
	}
	clipboardReads.mu.Lock()
	clipboardReads.pending = append(clipboardReads.pending, fn)
	clipboardReads.mu.Unlock()
	WriteTerminal(p.terminal)
}

// clipboardTool is a platform's copy and paste commands for one selection.
type clipboardTool struct {
	copy, paste []string
	utf16Input  bool // clip.exe reads stdin in the ANSI code page unless given UTF-16LE with a BOM
	trimCRLF    bool // PowerShell ends its output with a newline
}

// clipboardTools lists the tools for a local session, in preference order.
// It is empty when the platform has no tool for the selection.
func clipboardTools(goos string, getenv func(string) string, selection ClipboardSelection) []clipboardTool {
	primary := selection == PrimaryClipboard
	switch {
	case goos == "darwin":
		if primary {
			return nil
		}
		return []clipboardTool{{copy: []string{"pbcopy"}, paste: []string{"pbpaste"}}}
	case goos == "windows" || (goos == "linux" && (getenv("WSL_DISTRO_NAME") != "" || getenv("WSL_INTEROP") != "")):
		if primary {
			return nil
		}
		return []clipboardTool{{
			copy: []string{"clip.exe"},
			paste: []string{"powershell.exe", "-NoProfile", "-Command",
				"[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; Get-Clipboard -Raw"},
			utf16Input: true,
			trimCRLF:   true,
		}}
	case getenv("WAYLAND_DISPLAY") != "":
		if primary {
			return []clipboardTool{{copy: []string{"wl-copy", "--primary"}, paste: []string{"wl-paste", "--no-newline", "--primary"}}}
		}
		return []clipboardTool{{copy: []string{"wl-copy"}, paste: []string{"wl-paste", "--no-newline"}}}
	case getenv("DISPLAY") != "":
		xclipSel, xselSel := "clipboard", "--clipboard"
		if primary {
			xclipSel, xselSel = "primary", "--primary"
		}
		return []clipboardTool{
			{copy: []string{"xclip", "-selection", xclipSel}, paste: []string{"xclip", "-o", "-selection", xclipSel}},
			{copy: []string{"xsel", xselSel, "--input"}, paste: []string{"xsel", "--output", xselSel}},
		}
	}
	return nil
}

// clipboardHost is where the app runs, as seen from its environment.
type clipboardHost struct {
	remote bool
	tmux   bool
	screen bool
}

func detectClipboardHost(getenv func(string) string) clipboardHost {
	tmux := getenv("TMUX") != ""
	return clipboardHost{
		remote: getenv("SSH_CONNECTION") != "" || getenv("SSH_CLIENT") != "" || getenv("SSH_TTY") != "",
		tmux:   tmux,
		screen: !tmux && (getenv("STY") != "" || strings.HasPrefix(getenv("TERM"), "screen")),
	}
}

// wrap prepares an OSC 52 sequence for the multiplexer between the app and
// the terminal. tmux and screen drop a raw OSC 52 from an app.
func (h clipboardHost) wrap(seq string) string {
	switch {
	case h.tmux:
		return ansi.TmuxPassthrough(seq)
	case h.screen:
		return ansi.ScreenPassthrough(seq, 768)
	}
	return seq
}

func planClipboardWrite(sys clipboardSystem, method ClipboardMethod, selection ClipboardSelection, content string) clipboardWritePlan {
	host := detectClipboardHost(sys.getenv)
	var plan clipboardWritePlan
	if method != ClipboardTerminal {
		if !host.remote {
			for _, tool := range clipboardTools(sys.goos, sys.getenv, selection) {
				if _, err := sys.lookPath(tool.copy[0]); err != nil {
					continue
				}
				stdin := []byte(content)
				if tool.utf16Input {
					stdin = utf16LEWithBOM(content)
				}
				plan.native = &clipboardCommand{argv: tool.copy, stdin: stdin}
				break
			}
		}
		// load-buffer -w always sets the outer terminal's system clipboard.
		if host.tmux && selection == SystemClipboard {
			plan.tmux = &clipboardCommand{argv: []string{"tmux", "load-buffer", "-w", "-"}, stdin: []byte(content)}
		}
	}
	if method != ClipboardNative {
		plan.terminal = host.wrap(ansi.SetClipboard(selection, content))
	}
	return plan
}

func planClipboardRead(sys clipboardSystem, method ClipboardMethod, selection ClipboardSelection) clipboardReadPlan {
	host := detectClipboardHost(sys.getenv)
	var plan clipboardReadPlan
	if method != ClipboardTerminal && !host.remote {
		for _, tool := range clipboardTools(sys.goos, sys.getenv, selection) {
			if _, err := sys.lookPath(tool.paste[0]); err != nil {
				continue
			}
			plan.native = &clipboardCommand{argv: tool.paste}
			plan.trimCRLF = tool.trimCRLF
			break
		}
	}
	if method != ClipboardNative {
		plan.terminal = host.wrap(ansi.RequestClipboard(selection))
	}
	return plan
}

func utf16LEWithBOM(s string) []byte {
	units := utf16.Encode([]rune(s))
	out := make([]byte, 2, 2+2*len(units))
	binary.LittleEndian.PutUint16(out, 0xFEFF)
	for _, u := range units {
		out = binary.LittleEndian.AppendUint16(out, u)
	}
	return out
}

// clipboardRunner runs a clipboard command, returning its stdout when capture
// is set.
type clipboardRunner func(cmd clipboardCommand, capture bool) ([]byte, error)

// clipboardSystem is the outside world the clipboard depends on. Tests
// replace it so they never touch the real clipboard.
type clipboardSystem struct {
	goos     string
	getenv   func(string) string
	lookPath func(string) (string, error)
	run      clipboardRunner
}

var currentClipboardSystem atomic.Pointer[clipboardSystem]

func init() {
	currentClipboardSystem.Store(&clipboardSystem{
		goos:     runtime.GOOS,
		getenv:   os.Getenv,
		lookPath: exec.LookPath,
		run:      runClipboardCommand,
	})
}

func clipboardSys() clipboardSystem { return *currentClipboardSystem.Load() }

func runClipboardCommand(c clipboardCommand, capture bool) ([]byte, error) {
	// PowerShell can take seconds to start.
	timeout := 2 * time.Second
	if capture {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.argv[0], c.argv[1:]...)
	cmd.Stdin = bytes.NewReader(c.stdin)
	cmd.WaitDelay = 500 * time.Millisecond
	if capture {
		return cmd.Output()
	}
	// xclip and wl-copy leave a child holding the clipboard. With Stdout and
	// Stderr nil it gets /dev/null, so Run doesn't wait for that child.
	return nil, cmd.Run()
}

// clipboardJobs runs clipboard work one job at a time, in call order, off the
// caller's goroutine.
var clipboardJobs struct {
	mu      sync.Mutex
	queue   []func()
	running bool
}

// queueClipboardJob runs job after every earlier job. A cheap job (no command
// to run) runs on the caller's goroutine when nothing is queued ahead of it.
func queueClipboardJob(cheap bool, job func()) {
	clipboardJobs.mu.Lock()
	if cheap && !clipboardJobs.running {
		defer clipboardJobs.mu.Unlock()
		job()
		return
	}
	clipboardJobs.queue = append(clipboardJobs.queue, job)
	if !clipboardJobs.running {
		clipboardJobs.running = true
		go drainClipboardJobs()
	}
	clipboardJobs.mu.Unlock()
}

func drainClipboardJobs() {
	for {
		clipboardJobs.mu.Lock()
		if len(clipboardJobs.queue) == 0 {
			clipboardJobs.running = false
			clipboardJobs.mu.Unlock()
			return
		}
		job := clipboardJobs.queue[0]
		clipboardJobs.queue = clipboardJobs.queue[1:]
		clipboardJobs.mu.Unlock()
		job()
	}
}
