package terma

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unicode"

	uv "github.com/charmbracelet/ultraviolet"
)

// embedSocketEnv switches Run into embedded mode: instead of driving the
// terminal, the app renders into memory, writes each changed frame to stdout
// as one JSON line, and takes input events posted to an HTTP server on the
// Unix socket the variable names. A host that draws cells itself (a Claude
// Code pane, an editor panel) runs the program as a child process.
const embedSocketEnv = "TERMA_EMBED_SOCKET"

// embedFrame is one screen as stdout carries it. Each line is a list of runs
// of equal style, [text, style index], with styles listed once per frame.
type embedFrame struct {
	Cols   int          `json:"cols"`
	Rows   int          `json:"rows"`
	Styles []embedStyle `json:"styles"`
	Lines  [][][2]any   `json:"lines"`
	Focus  string       `json:"focus,omitempty"`
	index  map[embedStyle]int
}

type embedStyle struct {
	Fg     string `json:"fg,omitempty"`
	Bg     string `json:"bg,omitempty"`
	Bold   bool   `json:"b,omitempty"`
	Faint  bool   `json:"d,omitempty"`
	Italic bool   `json:"i,omitempty"`
	Under  bool   `json:"u,omitempty"`
	Strike bool   `json:"s,omitempty"`
	Invert bool   `json:"r,omitempty"`
}

// embedEvent is one input event a host posts to /events (one object, or a
// JSON array of them).
type embedEvent struct {
	Type   string `json:"t"` // key, down, up, move, wheel, resize
	Key    string `json:"key,omitempty"`
	Ctrl   bool   `json:"ctrl,omitempty"`
	Shift  bool   `json:"shift,omitempty"`
	Meta   bool   `json:"meta,omitempty"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Button string `json:"button,omitempty"`
	Dir    string `json:"dir,omitempty"`
	Cols   int    `json:"cols,omitempty"`
	Rows   int    `json:"rows,omitempty"`
}

func hexColor(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

func encodeFrame(buf *uv.Buffer, width, height int) embedFrame {
	f := embedFrame{Cols: width, Rows: height, Styles: []embedStyle{}, index: map[embedStyle]int{}}
	f.Lines = make([][][2]any, height)
	for y := 0; y < height; y++ {
		runs := [][2]any{}
		var text strings.Builder
		current := -1
		flush := func() {
			if text.Len() > 0 {
				runs = append(runs, [2]any{text.String(), current})
				text.Reset()
			}
		}
		for x := 0; x < width; {
			cell := buf.CellAt(x, y)
			content, step := " ", 1
			var style embedStyle
			if cell != nil {
				if cell.Content != "" {
					content = cell.Content
				}
				if cell.Width > 1 {
					step = cell.Width
				}
				s := cell.Style
				style = embedStyle{
					Fg:     hexColor(s.Fg),
					Bg:     hexColor(s.Bg),
					Bold:   s.Attrs&uv.AttrBold != 0,
					Faint:  s.Attrs&uv.AttrFaint != 0,
					Italic: s.Attrs&uv.AttrItalic != 0,
					Under:  s.Underline != uv.UnderlineNone,
					Strike: s.Attrs&uv.AttrStrikethrough != 0,
					Invert: s.Attrs&uv.AttrReverse != 0,
				}
			}
			idx, ok := f.index[style]
			if !ok {
				idx = len(f.Styles)
				f.Styles = append(f.Styles, style)
				f.index[style] = idx
			}
			if idx != current {
				flush()
				current = idx
			}
			text.WriteString(content)
			x += step
		}
		flush()
		f.Lines[y] = runs
	}
	return f
}

// embedKeys turns a host's key into the events a terminal would have sent.
// Hosts name special keys (up, return, pageup, ...) and send printable keys as
// the text typed, which is several characters when they arrive in one read.
func embedKeys(e embedEvent) []uv.KeyPressEvent {
	var k uv.Key
	if e.Ctrl {
		k.Mod |= uv.ModCtrl
	}
	if e.Shift {
		k.Mod |= uv.ModShift
	}
	if e.Meta {
		k.Mod |= uv.ModAlt
	}
	special := map[string]rune{
		"up": uv.KeyUp, "down": uv.KeyDown, "left": uv.KeyLeft, "right": uv.KeyRight,
		"return": uv.KeyEnter, "enter": uv.KeyEnter, "tab": uv.KeyTab,
		"backspace": uv.KeyBackspace, "delete": uv.KeyDelete,
		"pageup": uv.KeyPgUp, "pagedown": uv.KeyPgDown, "home": uv.KeyHome, "end": uv.KeyEnd,
		"escape": uv.KeyEscape, "space": uv.KeySpace,
	}
	if code, ok := special[e.Key]; ok {
		k.Code = code
		if code == uv.KeySpace && k.Mod&^uv.ModShift == 0 {
			k.Text = " "
		}
		return []uv.KeyPressEvent{uv.KeyPressEvent(k)}
	}
	var keys []uv.KeyPressEvent
	for _, r := range e.Key {
		key := uv.Key{Mod: k.Mod, Code: unicode.ToLower(r)}
		if unicode.IsUpper(r) {
			key.ShiftedCode = r
			key.Mod |= uv.ModShift
		}
		if key.Mod&(uv.ModCtrl|uv.ModAlt) == 0 {
			key.Text = string(r)
		}
		keys = append(keys, uv.KeyPressEvent(key))
	}
	return keys
}

func embedButton(name string) uv.MouseButton {
	switch name {
	case "right":
		return uv.MouseRight
	case "middle":
		return uv.MouseMiddle
	}
	return uv.MouseLeft
}

// serveEmbedEvents accepts events on the Unix socket at path and sends them on
// events until ctx ends.
func serveEmbedEvents(ctx context.Context, path string, events chan<- embedEvent) error {
	_ = os.Remove(path)
	listener, err := net.Listen("unix", path)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/events", func(w http.ResponseWriter, req *http.Request) {
		var batch []embedEvent
		dec := json.NewDecoder(req.Body)
		raw := json.RawMessage{}
		if err := dec.Decode(&raw); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if strings.HasPrefix(strings.TrimSpace(string(raw)), "[") {
			err = json.Unmarshal(raw, &batch)
		} else {
			var one embedEvent
			err = json.Unmarshal(raw, &one)
			batch = []embedEvent{one}
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, e := range batch {
			select {
			case events <- e:
			case <-ctx.Done():
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})
	server := &http.Server{Handler: mux}
	go func() {
		<-ctx.Done()
		_ = server.Close()
		_ = os.Remove(path)
	}()
	go func() { _ = server.Serve(listener) }()
	return nil
}

func embedInitialSize() (int, int) {
	width, height := 80, 24
	if size := os.Getenv("TERMA_EMBED_SIZE"); size != "" {
		if w, h, ok := strings.Cut(size, "x"); ok {
			if n, err := strconv.Atoi(w); err == nil && n > 0 {
				width = n
			}
			if n, err := strconv.Atoi(h); err == nil && n > 0 {
				height = n
			}
		}
	}
	return width, height
}

// runEmbedded is Run's event loop with the terminal replaced: frames go to
// stdout, events come from the socket.
func runEmbedded(root Widget, socketPath string) (runErr error) {
	ctx, cancel := context.WithCancel(context.Background())
	appCancel = cancel
	setAppRuntimeState(ctx, newDispatchQueue())

	animController := NewAnimationController(defaultFPS)
	currentController = animController
	swapRenderTrigger(make(chan struct{}, 1))

	events := make(chan embedEvent, 256)
	if err := serveEmbedEvents(ctx, socketPath, events); err != nil {
		cancel()
		return err
	}

	eventLoopDone := make(chan struct{})
	eventLoopStarted := false
	defer func() {
		if r := recover(); r != nil {
			recordPanic(Panic{Message: fmt.Sprint(r), StackTrace: string(debug.Stack())})
			runErr = ErrPanicked
		}
		cancel()
		if eventLoopStarted {
			<-eventLoopDone
		}
		appCancel = nil
		appRenderer = nil
		takeTerminalWrites()
		resetClipboardReads()
		swapRenderTrigger(nil)
		currentController = nil
		clearAppRuntimeState()
		animController.Stop()
		renderPanics()
	}()

	width, height := embedInitialSize()
	buf := uv.NewBuffer(width, height)

	focusManager := NewFocusManager()
	focusManager.SetRootWidget(root)
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	lastFocusedID := ""

	renderer := NewRenderer(buf, width, height, focusManager, focusedSignal, hoveredSignal)
	appRenderer = renderer

	updateFocusedSignal := func() bool {
		keybindsChanged := focusManager.syncKeybinds()
		focusedID := focusManager.FocusedID()
		if focusedID == lastFocusedID {
			return keybindsChanged
		}
		lastFocusedID = focusedID
		focusedSignal.Set(focusManager.Focused())
		return true
	}

	out := bufio.NewWriter(os.Stdout)
	lastFrame := ""
	writeFrame := func() {
		frame := encodeFrame(buf, width, height)
		frame.Focus = focusManager.FocusedID()
		data, err := json.Marshal(frame)
		if err != nil || string(data) == lastFrame {
			return
		}
		lastFrame = string(data)
		_, _ = out.Write(data)
		_ = out.WriteByte('\n')
		_ = out.Flush()
	}

	lastModalCount := 0
	mouse := newMouseRouter(renderer, focusManager, hoveredSignal)
	display := func() {
		drainPendingDispatches()
		updateFocusedSignal()
		focusables := renderer.Update(root)
		focusManager.SetFocusables(focusables)
		if updateFocusedSignal() {
			renderer.Update(root)
		}
		modalCount := renderer.ModalCount()
		for i := lastModalCount; i < modalCount; i++ {
			focusManager.SaveFocus()
		}
		for i := modalCount; i < lastModalCount; i++ {
			focusManager.RestoreFocus()
		}
		if pendingFocusID != "" {
			focusManager.FocusByID(pendingFocusID)
			pendingFocusID = ""
			if updateFocusedSignal() {
				renderer.Update(root)
			}
		}
		lastModalCount = modalCount
		if updateFocusedSignal() {
			renderer.Update(root)
		}
		if mouse.reconcileHover() {
			renderer.Update(root)
		}
		// Terminal sequences (OSC 52 clipboard writes) have nowhere to go.
		takeTerminalWrites()
		writeFrame()
	}

	renderInterval := time.Second / time.Duration(defaultFPS)
	scheduler := newFrameScheduler(renderInterval, display)
	renderNow := scheduler.renderNow
	requestRender := func() { scheduler.request() }

	blinkTicker := time.NewTicker(cursorBlinkInterval)
	defer blinkTicker.Stop()

	renderNow()

	eventLoopStarted = true
	go func() {
		defer close(eventLoopDone)
		defer scheduler.stopTimer()
		defer func() {
			if r := recover(); r != nil {
				recordPanic(Panic{Message: fmt.Sprint(r), StackTrace: string(debug.Stack())})
				cancel()
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-currentRenderTrigger():
				requestRender()
			case <-animController.Tick():
				animController.Update()
				requestRender()
			case <-scheduler.timerCh:
				if scheduler.pending {
					renderNow()
				}
			case <-blinkTicker.C:
				toggleCursorBlink()
			case e := <-events:
				switch e.Type {
				case "resize":
					if e.Cols <= 0 || e.Rows <= 0 || (e.Cols == width && e.Rows == height) {
						continue
					}
					width, height = e.Cols, e.Rows
					renderer.Resize(width, height)
					buf.Clear()
					requestRender()
				case "key":
					for _, ev := range embedKeys(e) {
						if ev.MatchString("ctrl+c") {
							cancel()
							return
						}
						if scheduler.pending {
							renderNow()
						}
						showCursorForInput()
						blinkTicker.Reset(cursorBlinkInterval)
						dispatchKey(renderer, focusManager, root, KeyEvent{event: ev})
						requestRender()
					}
				case "down":
					showCursorForInput()
					mouse.press(uv.MouseClickEvent{X: e.X, Y: e.Y, Button: embedButton(e.Button)}, 0.5, 0.5, time.Now())
					requestRender()
				case "up":
					mouse.release(uv.MouseReleaseEvent{X: e.X, Y: e.Y, Button: embedButton(e.Button)}, 0.5, 0.5)
					requestRender()
				case "move":
					button := uv.MouseNone
					if e.Button != "" {
						button = embedButton(e.Button)
					}
					if mouse.motion(uv.MouseMotionEvent{X: e.X, Y: e.Y, Button: button}, 0.5, 0.5) {
						requestRender()
					}
				case "wheel":
					button := uv.MouseWheelDown
					if e.Dir == "up" {
						button = uv.MouseWheelUp
					}
					if mouse.wheelAt(uv.MouseWheelEvent{X: e.X, Y: e.Y, Button: button}, 0.5, 0.5) {
						requestRender()
					}
				}
			}
		}
	}()

	<-ctx.Done()
	return runErr
}
