//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd && !dragonfly

package main

import "github.com/gorilla/websocket"

func runSession(conn *websocket.Conn, command []string, cols, rows int) {
	sendStatus(conn, "terma-browser requires a Unix host (macOS, Linux, or BSD).")
}
