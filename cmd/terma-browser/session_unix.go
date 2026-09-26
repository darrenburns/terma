//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

func runSession(conn *websocket.Conn, command []string, cols, rows int) {
	cmd := exec.Command(command[0], command[1:]...)
	// The browser is a known true-colour terminal, independent of the
	// launcher's environment (agent shells commonly set NO_COLOR).
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "NO_COLOR=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "TERM=xterm-256color", "COLORTERM=truecolor", "TERMA_ENABLE_KITTY_KEYBOARD=0")
	terminal, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	if err != nil {
		sendStatus(conn, fmt.Sprintf("Could not start app: %v", err))
		return
	}
	outputDone := make(chan struct{})
	defer func() {
		// StartWithSize creates a session/process group, including go run children.
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		_ = terminal.Close()
		_ = conn.Close()
		<-outputDone
	}()
	go func() {
		defer close(outputDone)
		defer conn.Close()
		buffer := make([]byte, 32*1024)
		for {
			n, err := terminal.Read(buffer)
			if n > 0 {
				_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if writeErr := conn.WriteMessage(websocket.BinaryMessage, buffer[:n]); writeErr != nil {
					_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
					break
				}
			}
			if err != nil {
				break
			}
		}
		err := cmd.Wait()
		message := "App exited. Reload to start a new session."
		if err != nil {
			message = fmt.Sprintf("App exited: %v. Reload to restart.", err)
		}
		sendStatus(conn, message)
	}()
	for {
		var message inputMessage
		if err := conn.ReadJSON(&message); err != nil {
			return
		}
		switch message.Type {
		case "input":
			if _, err := terminal.Write([]byte(message.Data)); err != nil {
				return
			}
		case "resize":
			if !validSize(message.Cols, message.Rows) {
				return
			}
			if err := pty.Setsize(terminal, &pty.Winsize{Cols: uint16(message.Cols), Rows: uint16(message.Rows)}); err != nil {
				return
			}
		default:
			return
		}
	}
}
