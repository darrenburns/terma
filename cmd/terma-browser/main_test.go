//go:build darwin || linux

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/colorprofile"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func testServer(t *testing.T, command ...string) (*httptest.Server, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	server := httptest.NewUnstartedServer(nil)
	handler := browserHandler(ctx, server.Listener.Addr().String(), "test-token", command)
	server.Config.Handler = handler
	server.Start()
	t.Cleanup(func() { cancel(); server.Close(); handler.sessions.Wait() })
	return server, cancel
}

func connect(t *testing.T, server *httptest.Server) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/session?token=test-token", nil)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	require.NoError(t, conn.WriteJSON(inputMessage{Type: "resize", Cols: 80, Rows: 24}))
	return conn
}

func readUntil(t *testing.T, conn *websocket.Conn, want string) string {
	t.Helper()
	var output strings.Builder
	for !strings.Contains(output.String(), want) {
		_, data, err := conn.ReadMessage()
		require.NoError(t, err, output.String())
		output.Write(data)
	}
	return output.String()
}

func TestBrowserSessionInputAndResize(t *testing.T) {
	server, _ := testServer(t, "sh", "-c", "stty -echo; stty size; read value; stty size; printf 'received:%s\\n' \"$value\"")
	conn := connect(t, server)
	readUntil(t, conn, "24 80")
	require.NoError(t, conn.WriteJSON(inputMessage{Type: "resize", Cols: 100, Rows: 40}))
	require.NoError(t, conn.WriteJSON(inputMessage{Type: "input", Data: "hello\n"}))
	output := readUntil(t, conn, "received:hello")
	require.Contains(t, output, "40 100")
	readUntil(t, conn, "App exited.")
}

func TestBrowserAccessChecks(t *testing.T) {
	server, _ := testServer(t, "sh")
	for _, tc := range []struct{ name, token, origin string }{
		{"missing token", "", ""},
		{"wrong token", "wrong", ""},
		{"cross origin", "test-token", "https://example.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			header := http.Header{}
			if tc.origin != "" {
				header.Set("Origin", tc.origin)
			}
			_, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/session?token="+tc.token, header)
			require.Error(t, err)
			require.Equal(t, http.StatusForbidden, response.StatusCode)
			response.Body.Close()
		})
	}
	request, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)
	request.Host = "evil.example"
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusForbidden, response.StatusCode)
}

func TestBrowserAssetsBundled(t *testing.T) {
	server, _ := testServer(t, "sh")
	for _, path := range []string{"/", "/terminal.js", "/vendor/xterm.js", "/vendor/xterm.css"} {
		response, err := http.Get(server.URL + path)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
		data, err := io.ReadAll(response.Body)
		response.Body.Close()
		require.NoError(t, err)
		require.NotEmpty(t, data)
	}
}

func TestBrowserCancellation(t *testing.T) {
	server, cancel := testServer(t, "sh", "-c", "printf ready; sleep 60")
	conn := connect(t, server)
	readUntil(t, conn, "ready")
	cancel()
	_, _, err := conn.ReadMessage()
	require.Error(t, err)
}

func TestBrowserStartFailure(t *testing.T) {
	server, _ := testServer(t, "/nonexistent/terma-app")
	conn := connect(t, server)
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)
	var message map[string]string
	require.NoError(t, json.Unmarshal(data, &message))
	require.Contains(t, message["message"], "Could not start app")
}

func TestBrowserRejectsInvalidSize(t *testing.T) {
	for _, size := range [][2]int{{0, 24}, {80, 0}, {-1, 24}, {501, 24}, {80, 301}} {
		require.False(t, validSize(size[0], size[1]))
	}
	require.True(t, validSize(80, 24))
}

// Run the same profile detection as Ultraviolet inside the browser PTY.
func TestBrowserColorProfileHelper(t *testing.T) {
	if os.Getenv("TERMA_TEST_COLOR_HELPER") != "1" {
		return
	}
	fmt.Printf("profile:%s\n", colorprofile.Detect(os.Stdout, os.Environ()))
	os.Exit(0)
}

func TestBrowserSessionTrueColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERMA_TEST_COLOR_HELPER", "1")
	executable, err := os.Executable()
	require.NoError(t, err)
	server, _ := testServer(t, executable, "-test.run=^TestBrowserColorProfileHelper$")
	conn := connect(t, server)
	output := readUntil(t, conn, "profile:")
	require.Contains(t, output, "profile:TrueColor")
}
