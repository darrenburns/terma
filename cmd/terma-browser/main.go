// terma-browser runs a local command in a browser terminal.
package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed web
var assets embed.FS

type inputMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

func validSize(cols, rows int) bool { return cols >= 2 && cols <= 500 && rows >= 2 && rows <= 300 }

func main() {
	port := flag.Int("port", 0, "localhost port (0 chooses an available port)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: terma-browser [-port 8080] -- command [args...]")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	listener, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		log.Fatal(err)
	}
	token := hex.EncodeToString(secret)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	handler := browserHandler(ctx, listener.Addr().String(), token, flag.Args())
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	shutdownDone := make(chan struct{})
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
		close(shutdownDone)
	}()
	fmt.Printf("Open http://%s/#%s\nEach tab starts a new app; closing it stops that app.\n", listener.Addr(), token)
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	stop()
	<-shutdownDone
	handler.sessions.Wait()
}

type sessionHandler struct {
	http.Handler
	sessions sync.WaitGroup
}

func browserHandler(ctx context.Context, host, token string, command []string) *sessionHandler {
	handler := &sessionHandler{}
	mux := http.NewServeMux()
	files, _ := fs.Sub(assets, "web")
	mux.Handle("/", http.FileServer(http.FS(files)))
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		handler.sessions.Add(1)
		defer handler.sessions.Done()
		if r.URL.Query().Get("token") != token {
			http.Error(w, "invalid session token", http.StatusForbidden)
			return
		}
		// The default upgrader rejects cross-origin browser connections.
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
		defer stop()
		conn.SetReadLimit(64 * 1024)
		_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		var initial inputMessage
		if err := conn.ReadJSON(&initial); err != nil || initial.Type != "resize" || !validSize(initial.Cols, initial.Rows) {
			return
		}
		_ = conn.SetReadDeadline(time.Time{})
		runSession(conn, command, initial.Cols, initial.Rows)
	})
	handler.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != host {
			http.Error(w, "invalid host", http.StatusForbidden)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; frame-ancestors 'none'")
		mux.ServeHTTP(w, r)
	})
	return handler
}

func sendStatus(conn *websocket.Conn, message string) {
	data, _ := json.Marshal(map[string]string{"type": "status", "message": message})
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_ = conn.WriteMessage(websocket.TextMessage, data)
}
