package terma

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withRunningApp makes the package behave as if an app were running, so
// terminal writes are queued, and clears the queue afterwards.
func withRunningApp(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	setAppRuntimeState(ctx, newDispatchQueue())
	takeTerminalWrites()
	resetClipboardReads()
	t.Cleanup(func() {
		cancel()
		clearAppRuntimeState()
		takeTerminalWrites()
		resetClipboardReads()
	})
}

func TestClipboardSelectionConstants(t *testing.T) {
	assert.EqualValues(t, ansi.SystemClipboard, SystemClipboard)
	assert.EqualValues(t, ansi.PrimaryClipboard, PrimaryClipboard)
}

func TestSetClipboard_QueuesOSC52WriteForNextFrame(t *testing.T) {
	withRunningApp(t)

	SetClipboard(SystemClipboard, "hello")
	SetClipboard(PrimaryClipboard, "world")

	assert.Equal(t, []string{
		ansi.SetClipboard(SystemClipboard, "hello"),
		ansi.SetClipboard(PrimaryClipboard, "world"),
	}, takeTerminalWrites())
	assert.Empty(t, takeTerminalWrites(), "writes are taken once")
}

func TestSetClipboard_WithoutRunningAppWritesNothing(t *testing.T) {
	takeTerminalWrites()
	SetClipboard(SystemClipboard, "hello")
	assert.Empty(t, takeTerminalWrites())
}

func TestWriteTerminal_IgnoresEmptySequence(t *testing.T) {
	withRunningApp(t)
	WriteTerminal("")
	assert.Empty(t, takeTerminalWrites())
}

func TestReadClipboard_RequestsAndDeliversRepliesInOrder(t *testing.T) {
	withRunningApp(t)

	var got []string
	ReadClipboard(SystemClipboard, func(s string) { got = append(got, "first:"+s) })
	ReadClipboard(PrimaryClipboard, func(s string) { got = append(got, "second:"+s) })

	assert.Equal(t, []string{
		ansi.RequestClipboard(SystemClipboard),
		ansi.RequestClipboard(PrimaryClipboard),
	}, takeTerminalWrites())

	assert.True(t, deliverClipboard("a"))
	assert.True(t, deliverClipboard("b"))
	assert.False(t, deliverClipboard("unrequested"), "a reply nobody asked for is dropped")
	assert.Equal(t, []string{"first:a", "second:b"}, got)
}

func TestReadClipboard_WithoutRunningAppDoesNothing(t *testing.T) {
	resetClipboardReads()
	called := false
	ReadClipboard(SystemClipboard, func(string) { called = true })
	assert.False(t, deliverClipboard("x"))
	assert.False(t, called)
}

// isolateClipboard keeps every test in the package away from the developer's
// real clipboard: no environment, no tools, and a runner that refuses to run.
func isolateClipboard() {
	currentClipboardSystem.Store(&clipboardSystem{
		goos:     "isolated",
		getenv:   func(string) string { return "" },
		lookPath: func(name string) (string, error) { return "", exec.ErrNotFound },
		run: func(cmd clipboardCommand, capture bool) ([]byte, error) {
			panic("test ran a real clipboard command: " + strings.Join(cmd.argv, " "))
		},
	})
}

type ranCommand struct {
	argv  string
	stdin string
}

// fakeClipboard is a clipboard system with a chosen OS, environment and set of
// tools, recording what it runs.
type fakeClipboard struct {
	goos  string
	env   map[string]string
	tools []string
	// fail lists tools that exit non-zero.
	fail []string
	// output is what each paste tool prints.
	output map[string]string
	// delay holds a command whose stdin matches the key before it returns.
	delay map[string]time.Duration
	// failCalls makes the nth call (from 1) fail, whatever the tool.
	failCalls []int

	mu  sync.Mutex
	ran []ranCommand
}

func (f *fakeClipboard) system() *clipboardSystem {
	return &clipboardSystem{
		goos:   f.goos,
		getenv: func(k string) string { return f.env[k] },
		lookPath: func(name string) (string, error) {
			if slices.Contains(f.tools, name) {
				return "/fake/" + name, nil
			}
			return "", exec.ErrNotFound
		},
		run: f.run,
	}
}

func (f *fakeClipboard) run(cmd clipboardCommand, capture bool) ([]byte, error) {
	time.Sleep(f.delay[string(cmd.stdin)])
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ran = append(f.ran, ranCommand{strings.Join(cmd.argv, " "), string(cmd.stdin)})
	if slices.Contains(f.fail, cmd.argv[0]) || slices.Contains(f.failCalls, len(f.ran)) {
		return nil, errors.New("exit status 1")
	}
	return []byte(f.output[cmd.argv[0]]), nil
}

func (f *fakeClipboard) commands() []ranCommand {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ran)
}

// useClipboard swaps in fake for the test and restores the isolated system
// and the default method afterwards.
func useClipboard(t *testing.T, fake *fakeClipboard, method ClipboardMethod) {
	t.Helper()
	currentClipboardSystem.Store(fake.system())
	SetClipboardMethod(method)
	t.Cleanup(func() {
		waitClipboardIdle(t)
		SetClipboardMethod(ClipboardAuto)
		isolateClipboard()
	})
}

func waitClipboardIdle(t *testing.T) {
	t.Helper()
	require.Eventually(t, func() bool {
		clipboardJobs.mu.Lock()
		defer clipboardJobs.mu.Unlock()
		return !clipboardJobs.running
	}, 5*time.Second, time.Millisecond)
}

func TestPlanClipboardWrite(t *testing.T) {
	ssh := map[string]string{"SSH_CONNECTION": "10.0.0.1 22 10.0.0.2 22"}
	raw := ansi.SetClipboard(SystemClipboard, "héllo")
	tmuxLoad := &clipboardCommand{argv: []string{"tmux", "load-buffer", "-w", "-"}, stdin: []byte("héllo")}
	tests := []struct {
		name      string
		fake      *fakeClipboard
		selection ClipboardSelection
		method    ClipboardMethod
		want      clipboardWritePlan
	}{
		{
			name: "darwin local uses pbcopy",
			fake: &fakeClipboard{goos: "darwin", tools: []string{"pbcopy"}},
			want: clipboardWritePlan{native: &clipboardCommand{argv: []string{"pbcopy"}, stdin: []byte("héllo")}, terminal: raw},
		},
		{
			name: "darwin over SSH uses OSC 52",
			fake: &fakeClipboard{goos: "darwin", env: ssh, tools: []string{"pbcopy"}},
			want: clipboardWritePlan{terminal: raw},
		},
		{
			name: "darwin over SSH_TTY alone uses OSC 52",
			fake: &fakeClipboard{goos: "darwin", env: map[string]string{"SSH_TTY": "/dev/ttys001"}, tools: []string{"pbcopy"}},
			want: clipboardWritePlan{terminal: raw},
		},
		{
			name: "linux Wayland uses wl-copy",
			fake: &fakeClipboard{goos: "linux", env: map[string]string{"WAYLAND_DISPLAY": "wayland-0", "DISPLAY": ":0"}, tools: []string{"wl-copy", "xclip"}},
			want: clipboardWritePlan{native: &clipboardCommand{argv: []string{"wl-copy"}, stdin: []byte("héllo")}, terminal: raw},
		},
		{
			name: "linux X11 prefers xclip",
			fake: &fakeClipboard{goos: "linux", env: map[string]string{"DISPLAY": ":0"}, tools: []string{"xclip", "xsel"}},
			want: clipboardWritePlan{native: &clipboardCommand{argv: []string{"xclip", "-selection", "clipboard"}, stdin: []byte("héllo")}, terminal: raw},
		},
		{
			name: "linux X11 with only xsel",
			fake: &fakeClipboard{goos: "linux", env: map[string]string{"DISPLAY": ":0"}, tools: []string{"xsel"}},
			want: clipboardWritePlan{native: &clipboardCommand{argv: []string{"xsel", "--clipboard", "--input"}, stdin: []byte("héllo")}, terminal: raw},
		},
		{
			name: "linux with no display uses OSC 52",
			fake: &fakeClipboard{goos: "linux", tools: []string{"xclip", "wl-copy"}},
			want: clipboardWritePlan{terminal: raw},
		},
		{
			name: "WSL uses clip.exe with UTF-16LE and a BOM",
			fake: &fakeClipboard{goos: "linux", env: map[string]string{"WSL_DISTRO_NAME": "Ubuntu", "DISPLAY": ":0"}, tools: []string{"clip.exe", "xclip"}},
			want: clipboardWritePlan{native: &clipboardCommand{argv: []string{"clip.exe"}, stdin: []byte{0xFF, 0xFE, 'h', 0, 0xE9, 0, 'l', 0, 'l', 0, 'o', 0}}, terminal: raw},
		},
		{
			name: "WSL detected by WSL_INTEROP",
			fake: &fakeClipboard{goos: "linux", env: map[string]string{"WSL_INTEROP": "/run/WSL/1_interop"}, tools: []string{"clip.exe"}},
			want: clipboardWritePlan{native: &clipboardCommand{argv: []string{"clip.exe"}, stdin: []byte{0xFF, 0xFE, 'h', 0, 0xE9, 0, 'l', 0, 'l', 0, 'o', 0}}, terminal: raw},
		},
		{
			name: "windows uses clip.exe",
			fake: &fakeClipboard{goos: "windows", tools: []string{"clip.exe"}},
			want: clipboardWritePlan{native: &clipboardCommand{argv: []string{"clip.exe"}, stdin: []byte{0xFF, 0xFE, 'h', 0, 0xE9, 0, 'l', 0, 'l', 0, 'o', 0}}, terminal: raw},
		},
		{
			name:      "primary selection on darwin uses OSC 52",
			fake:      &fakeClipboard{goos: "darwin", tools: []string{"pbcopy"}},
			selection: PrimaryClipboard,
			want:      clipboardWritePlan{terminal: ansi.SetClipboard(PrimaryClipboard, "héllo")},
		},
		{
			name:      "primary selection on Wayland uses wl-copy --primary",
			fake:      &fakeClipboard{goos: "linux", env: map[string]string{"WAYLAND_DISPLAY": "wayland-0"}, tools: []string{"wl-copy"}},
			selection: PrimaryClipboard,
			want:      clipboardWritePlan{native: &clipboardCommand{argv: []string{"wl-copy", "--primary"}, stdin: []byte("héllo")}, terminal: ansi.SetClipboard(PrimaryClipboard, "héllo")},
		},
		{
			name: "tmux with no tool loads a buffer and wraps OSC 52",
			fake: &fakeClipboard{goos: "linux", env: map[string]string{"TMUX": "/tmp/tmux-501/default,1,0", "TERM": "screen-256color"}},
			want: clipboardWritePlan{tmux: tmuxLoad, terminal: ansi.TmuxPassthrough(raw)},
		},
		{
			name: "tmux over SSH skips the local tool",
			fake: &fakeClipboard{goos: "darwin", env: map[string]string{"TMUX": "/tmp/tmux", "SSH_CLIENT": "10.0.0.1 5000 22"}, tools: []string{"pbcopy"}},
			want: clipboardWritePlan{tmux: tmuxLoad, terminal: ansi.TmuxPassthrough(raw)},
		},
		{
			name:      "tmux buffer is not used for the primary selection",
			fake:      &fakeClipboard{goos: "linux", env: map[string]string{"TMUX": "/tmp/tmux"}},
			selection: PrimaryClipboard,
			want:      clipboardWritePlan{terminal: ansi.TmuxPassthrough(ansi.SetClipboard(PrimaryClipboard, "héllo"))},
		},
		{
			name: "screen wraps OSC 52",
			fake: &fakeClipboard{goos: "linux", env: map[string]string{"STY": "1234.pts-0.host"}},
			want: clipboardWritePlan{terminal: ansi.ScreenPassthrough(raw, 768)},
		},
		{
			name: "screen detected by TERM",
			fake: &fakeClipboard{goos: "linux", env: map[string]string{"TERM": "screen.xterm-256color"}},
			want: clipboardWritePlan{terminal: ansi.ScreenPassthrough(raw, 768)},
		},
		{
			name:   "native method never plans OSC 52",
			fake:   &fakeClipboard{goos: "darwin", env: map[string]string{"TMUX": "/tmp/tmux"}, tools: []string{"pbcopy"}},
			method: ClipboardNative,
			want:   clipboardWritePlan{native: &clipboardCommand{argv: []string{"pbcopy"}, stdin: []byte("héllo")}, tmux: tmuxLoad},
		},
		{
			name:   "native method over SSH outside tmux plans nothing",
			fake:   &fakeClipboard{goos: "darwin", env: ssh, tools: []string{"pbcopy"}},
			method: ClipboardNative,
			want:   clipboardWritePlan{},
		},
		{
			name:   "terminal method never plans a tool",
			fake:   &fakeClipboard{goos: "darwin", env: map[string]string{"TMUX": "/tmp/tmux"}, tools: []string{"pbcopy"}},
			method: ClipboardTerminal,
			want:   clipboardWritePlan{terminal: ansi.TmuxPassthrough(raw)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selection := tt.selection
			if selection == 0 {
				selection = SystemClipboard
			}
			got := planClipboardWrite(*tt.fake.system(), tt.method, selection, "héllo")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPlanClipboardWrite_ScreenChunksLongPayloads(t *testing.T) {
	fake := &fakeClipboard{goos: "linux", env: map[string]string{"STY": "1234.pts-0.host"}}
	content := strings.Repeat("0123456789", 300)
	plan := planClipboardWrite(*fake.system(), ClipboardAuto, SystemClipboard, content)

	chunks := strings.Split(strings.TrimSuffix(plan.terminal, "\x1b\\"), "\x1b\\")
	require.Greater(t, len(chunks), 1, "a long payload is split")
	var joined strings.Builder
	for _, chunk := range chunks {
		require.True(t, strings.HasPrefix(chunk, "\x1bP"), "each chunk is its own DCS string")
		payload := strings.TrimPrefix(chunk, "\x1bP")
		assert.LessOrEqual(t, len(payload), 768)
		joined.WriteString(payload)
	}
	assert.Equal(t, ansi.SetClipboard(SystemClipboard, content), joined.String())
}

func TestSetClipboard_NativeSuccessSendsNoOSC52(t *testing.T) {
	withRunningApp(t)
	fake := &fakeClipboard{goos: "darwin", tools: []string{"pbcopy"}}
	useClipboard(t, fake, ClipboardAuto)

	SetClipboard(SystemClipboard, "hello")
	waitClipboardIdle(t)

	assert.Equal(t, []ranCommand{{"pbcopy", "hello"}}, fake.commands())
	assert.Empty(t, takeTerminalWrites())
}

func TestSetClipboard_NativeFailureFallsBackToOSC52(t *testing.T) {
	withRunningApp(t)
	fake := &fakeClipboard{goos: "darwin", tools: []string{"pbcopy"}, fail: []string{"pbcopy"}}
	useClipboard(t, fake, ClipboardAuto)

	SetClipboard(SystemClipboard, "hello")
	waitClipboardIdle(t)

	assert.Equal(t, []ranCommand{{"pbcopy", "hello"}}, fake.commands())
	assert.Equal(t, []string{ansi.SetClipboard(SystemClipboard, "hello")}, takeTerminalWrites())
}

func TestSetClipboard_InTmuxLoadsBufferAndSendsOnlyWrappedOSC52(t *testing.T) {
	withRunningApp(t)
	fake := &fakeClipboard{goos: "linux", env: map[string]string{"TMUX": "/tmp/tmux"}, fail: []string{"tmux"}}
	useClipboard(t, fake, ClipboardAuto)

	SetClipboard(SystemClipboard, "hello")
	waitClipboardIdle(t)

	assert.Equal(t, []ranCommand{{"tmux load-buffer -w -", "hello"}}, fake.commands())
	writes := takeTerminalWrites()
	assert.Equal(t, []string{ansi.TmuxPassthrough(ansi.SetClipboard(SystemClipboard, "hello"))}, writes)
	assert.NotContains(t, writes, ansi.SetClipboard(SystemClipboard, "hello"))
}

func TestSetClipboard_NativeMethodWithNoToolWritesNothing(t *testing.T) {
	withRunningApp(t)
	fake := &fakeClipboard{goos: "linux", env: map[string]string{"STY": "1.pts"}}
	useClipboard(t, fake, ClipboardNative)

	SetClipboard(SystemClipboard, "hello")
	waitClipboardIdle(t)

	assert.Empty(t, fake.commands())
	assert.Empty(t, takeTerminalWrites())
}

func TestSetClipboard_TerminalMethodRunsNoTool(t *testing.T) {
	withRunningApp(t)
	fake := &fakeClipboard{goos: "darwin", tools: []string{"pbcopy"}}
	useClipboard(t, fake, ClipboardTerminal)

	SetClipboard(SystemClipboard, "hello")
	waitClipboardIdle(t)

	assert.Empty(t, fake.commands())
	assert.Equal(t, []string{ansi.SetClipboard(SystemClipboard, "hello")}, takeTerminalWrites())
}

func TestSetClipboard_NativeWorksWithoutRunningApp(t *testing.T) {
	takeTerminalWrites()
	fake := &fakeClipboard{goos: "darwin", tools: []string{"pbcopy"}}
	useClipboard(t, fake, ClipboardAuto)

	SetClipboard(SystemClipboard, "hello")
	waitClipboardIdle(t)

	assert.Equal(t, []ranCommand{{"pbcopy", "hello"}}, fake.commands())
	assert.Empty(t, takeTerminalWrites())
}

func TestSetClipboard_CopiesApplyInCallOrder(t *testing.T) {
	withRunningApp(t)
	// The first copy is slow; run concurrently, it would finish last.
	fake := &fakeClipboard{goos: "darwin", tools: []string{"pbcopy"}, delay: map[string]time.Duration{"first": 50 * time.Millisecond}}
	useClipboard(t, fake, ClipboardAuto)

	SetClipboard(SystemClipboard, "first")
	SetClipboard(SystemClipboard, "second")
	waitClipboardIdle(t)

	assert.Equal(t, []ranCommand{{"pbcopy", "first"}, {"pbcopy", "second"}}, fake.commands())
}

func TestSetClipboard_OSC52FallbackStaysBehindSlowNativeCopy(t *testing.T) {
	withRunningApp(t)
	fake := &fakeClipboard{goos: "darwin", tools: []string{"pbcopy"}, delay: map[string]time.Duration{"first": 50 * time.Millisecond}}
	useClipboard(t, fake, ClipboardAuto)

	SetClipboard(SystemClipboard, "first")
	SetClipboardMethod(ClipboardTerminal)
	SetClipboard(SystemClipboard, "second")
	assert.Empty(t, takeTerminalWrites(), "the OSC 52 copy waits for the native copy ahead of it")
	waitClipboardIdle(t)

	assert.Equal(t, []ranCommand{{"pbcopy", "first"}}, fake.commands())
	assert.Equal(t, []string{ansi.SetClipboard(SystemClipboard, "second")}, takeTerminalWrites())
}

func TestRunClipboardCommand_DoesNotWaitForForkedChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("needs a POSIX shell")
	}
	dir := t.TempDir()
	tool := filepath.Join(dir, "fake-copy")
	pidFile := filepath.Join(dir, "child.pid")
	// Like xclip, keep a background child holding the inherited stdout/stderr.
	script := "#!/bin/sh\ncat >/dev/null\nsleep 30 &\necho $! > \"$1\"\n"
	require.NoError(t, os.WriteFile(tool, []byte(script), 0o755))
	t.Cleanup(func() {
		if pid, err := os.ReadFile(pidFile); err == nil {
			_ = exec.Command("kill", strings.TrimSpace(string(pid))).Run()
		}
	})

	start := time.Now()
	_, err := runClipboardCommand(clipboardCommand{argv: []string{tool, pidFile}, stdin: []byte("hello")}, false)

	require.NoError(t, err)
	assert.Less(t, time.Since(start), time.Second)
}

func TestReadClipboard_NativeAnswerIsDeliveredOnTheLoop(t *testing.T) {
	withRunningApp(t)
	fake := &fakeClipboard{goos: "linux", env: map[string]string{"WAYLAND_DISPLAY": "wayland-0"}, tools: []string{"wl-paste"}, output: map[string]string{"wl-paste": "pasted"}}
	useClipboard(t, fake, ClipboardAuto)

	var got []string
	ReadClipboard(PrimaryClipboard, func(s string) { got = append(got, s) })
	waitClipboardIdle(t)
	assert.Empty(t, got, "the callback waits for the event loop")

	drainPendingDispatches()
	assert.Equal(t, []string{"pasted"}, got)
	assert.Equal(t, []ranCommand{{"wl-paste --no-newline --primary", ""}}, fake.commands())
	assert.Empty(t, takeTerminalWrites())
}

func TestReadClipboard_PowerShellTrailingNewlineIsTrimmed(t *testing.T) {
	withRunningApp(t)
	fake := &fakeClipboard{goos: "windows", tools: []string{"powershell.exe"}, output: map[string]string{"powershell.exe": "line one\r\nline two\r\n"}}
	useClipboard(t, fake, ClipboardAuto)

	var got string
	ReadClipboard(SystemClipboard, func(s string) { got = s })
	waitClipboardIdle(t)
	drainPendingDispatches()

	assert.Equal(t, "line one\r\nline two", got)
}

func TestReadClipboard_FallsBackToOSC52QueryWhenToolFails(t *testing.T) {
	withRunningApp(t)
	fake := &fakeClipboard{goos: "darwin", env: map[string]string{"TMUX": "/tmp/tmux"}, tools: []string{"pbpaste"}, fail: []string{"pbpaste"}}
	useClipboard(t, fake, ClipboardAuto)

	var got string
	ReadClipboard(SystemClipboard, func(s string) { got = s })
	waitClipboardIdle(t)

	assert.Equal(t, []ranCommand{{"pbpaste", ""}}, fake.commands())
	assert.Equal(t, []string{ansi.TmuxPassthrough(ansi.RequestClipboard(SystemClipboard))}, takeTerminalWrites())
	assert.True(t, deliverClipboard("from terminal"))
	assert.Equal(t, "from terminal", got)
}

func TestReadClipboard_TerminalRepliesSkipNativelyAnsweredReads(t *testing.T) {
	withRunningApp(t)
	// The first and third reads fail natively and wait on the terminal.
	fake := &fakeClipboard{goos: "darwin", tools: []string{"pbpaste"}, output: map[string]string{"pbpaste": "native"}, failCalls: []int{1, 3}}
	useClipboard(t, fake, ClipboardAuto)

	got := map[string]string{}
	for _, name := range []string{"first", "second", "third"} {
		ReadClipboard(SystemClipboard, func(s string) { got[name] = s })
	}
	waitClipboardIdle(t)
	drainPendingDispatches()

	assert.Equal(t, map[string]string{"second": "native"}, got)
	assert.Equal(t, []string{ansi.RequestClipboard(SystemClipboard), ansi.RequestClipboard(SystemClipboard)}, takeTerminalWrites())
	assert.True(t, deliverClipboard("reply one"))
	assert.True(t, deliverClipboard("reply two"))
	assert.False(t, deliverClipboard("extra"))
	assert.Equal(t, map[string]string{"first": "reply one", "second": "native", "third": "reply two"}, got)
}

func TestReadClipboard_NativeMethodNeverQueriesTheTerminal(t *testing.T) {
	withRunningApp(t)
	fake := &fakeClipboard{goos: "darwin", tools: []string{"pbpaste"}, fail: []string{"pbpaste"}}
	useClipboard(t, fake, ClipboardNative)

	ReadClipboard(SystemClipboard, func(string) { t.Error("no answer was available") })
	waitClipboardIdle(t)

	assert.Empty(t, takeTerminalWrites())
	assert.False(t, deliverClipboard("x"))
}
