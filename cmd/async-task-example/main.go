package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	t "github.com/darrenburns/terma"
)

type apiResult struct {
	Message  string `json:"message"`
	Endpoint string `json:"endpoint"`
	DelayMS  int    `json:"delay_ms"`
	ServedAt string `json:"served_at"`
}

type AsyncTaskDemo struct {
	baseURL string

	loadTask *t.Task[*apiResult]
	spinner  *t.SpinnerState
	status   t.Signal[string]

	// These are plain Go fields, not signals. They are only updated from the UI
	// thread via terma.Dispatch or from input handlers, which also run on the UI thread.
	currentRequestID int
	nextRequestID    int
	callbackCount    int
}

func NewAsyncTaskDemo(baseURL string) *AsyncTaskDemo {
	return &AsyncTaskDemo{
		baseURL:  baseURL,
		loadTask: t.NewTask[*apiResult](),
		spinner:  t.NewSpinnerState(t.SpinnerDots),
		status:   t.NewSignal("Ready"),
	}
}

func (a *AsyncTaskDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "1", Name: "Fast success", Action: a.fetchFastSuccess},
		{Key: "2", Name: "Slow success", Action: a.fetchSlowSuccess},
		{Key: "3", Name: "Server error", Action: a.fetchError},
		{Key: "c", Name: "Cancel", Action: a.cancelRequest},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *AsyncTaskDemo) fetchFastSuccess() {
	a.startRequest("/user?name=Ada&delay_ms=800", "Fetching user from /user...")
}

func (a *AsyncTaskDemo) fetchSlowSuccess() {
	a.startRequest("/user?name=Grace&delay_ms=3000", "Fetching slow user from /user...")
}

func (a *AsyncTaskDemo) fetchError() {
	a.startRequest("/fail?delay_ms=1500", "Calling /fail to simulate an error...")
}

func (a *AsyncTaskDemo) cancelRequest() {
	if !a.loadTask.Running.Peek() {
		return
	}
	a.currentRequestID = 0
	a.loadTask.Cancel()
	a.spinner.Stop()
	a.status.Set("Cancelled request")
}

func (a *AsyncTaskDemo) startRequest(path string, status string) {
	a.nextRequestID++
	requestID := a.nextRequestID
	a.currentRequestID = requestID
	a.status.Set(status)
	a.spinner.Start()

	url := a.baseURL + path
	started := time.Now()

	a.loadTask.Start(func(ctx context.Context) (*apiResult, error) {
		result, err := fetchJSON(ctx, url)
		duration := time.Since(started).Round(10 * time.Millisecond)

		t.Dispatch(func() {
			a.callbackCount++
			if a.currentRequestID != requestID {
				return
			}

			a.spinner.Stop()
			if ctx.Err() != nil {
				a.status.Set("Cancelled request")
				return
			}
			if err != nil {
				a.status.Set(fmt.Sprintf("Request failed after %s", duration))
				return
			}

			a.status.Set(fmt.Sprintf("Loaded response in %s", duration))
		})

		return result, err
	})
}

func (a *AsyncTaskDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	running := a.loadTask.Running.Get()
	phase := a.loadTask.Phase.Get()

	children := []t.Widget{
		t.Text{
			Content: "Async Task Example",
			Style: t.Style{
				ForegroundColor: theme.TextOnPrimary,
				BackgroundColor: theme.Primary,
				Padding:         t.EdgeInsetsXY(2, 0),
				Bold:            true,
			},
		},
		t.Text{
			Content: "Real HTTP requests run in the background. The UI stays interactive while they are in flight.",
			Style: t.Style{
				ForegroundColor: theme.TextMuted,
			},
		},
		t.Row{
			Spacing: 1,
			Children: []t.Widget{
				t.DisabledWhen(running, t.Button{
					Label:   "Fast Success",
					Variant: t.ButtonPrimary,
					OnPress: a.fetchFastSuccess,
				}),
				t.DisabledWhen(running, t.Button{
					Label:   "Slow Success",
					Variant: t.ButtonAccent,
					OnPress: a.fetchSlowSuccess,
				}),
				t.DisabledWhen(running, t.Button{
					Label:   "Server Error",
					Variant: t.ButtonError,
					OnPress: a.fetchError,
				}),
				t.DisabledWhen(!running, t.Button{
					Label:   "Cancel",
					Variant: t.ButtonWarning,
					OnPress: a.cancelRequest,
				}),
			},
		},
		t.Row{
			Spacing: 1,
			Children: []t.Widget{
				t.Label("Phase: "+taskPhaseLabel(phase), phaseVariant(phase), theme),
				t.Text{
					Content: "Status: " + a.status.Get(),
				},
			},
		},
		t.ShowWhen(running, t.Row{
			Spacing: 1,
			Children: []t.Widget{
				t.Spinner{State: a.spinner},
				t.Text{
					Content: "The request is running off the UI thread.",
				},
			},
		}),
		t.Text{
			Content: fmt.Sprintf("UI-thread callbacks applied: %d", a.callbackCount),
			Style: t.Style{
				ForegroundColor: theme.TextMuted,
			},
		},
		a.buildResult(theme, phase),
	}

	return t.Dock{
		Top: []t.Widget{
			t.Column{
				Style: t.Style{
					Padding: t.EdgeInsetsAll(2),
				},
				Spacing:  1,
				Children: children,
			},
		},
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
		},
		Body: t.Spacer{},
	}
}

func (a *AsyncTaskDemo) buildResult(theme t.ThemeData, phase t.TaskPhase) t.Widget {
	switch phase {
	case t.TaskSuccess:
		result := a.loadTask.Value.Get()
		if result == nil {
			return t.EmptyWidget{}
		}

		return t.Column{
			Spacing: 1,
			Children: []t.Widget{
				t.Label("Response", t.LabelSuccess, theme),
				t.Text{Content: "Message: " + result.Message},
				t.Text{Content: "Endpoint: " + result.Endpoint},
				t.Text{Content: fmt.Sprintf("Server delay: %dms", result.DelayMS)},
				t.Text{Content: "Served at: " + result.ServedAt},
			},
		}
	case t.TaskError:
		err := a.loadTask.Err.Get()
		if err == nil {
			return t.EmptyWidget{}
		}
		return t.Column{
			Spacing: 1,
			Children: []t.Widget{
				t.Label("Error", t.LabelError, theme),
				t.Text{Content: err.Error()},
			},
		}
	case t.TaskCancelled:
		return t.Label("The most recent request was cancelled.", t.LabelWarning, theme)
	default:
		return t.Text{
			Content: "Press 1/2/3 or click a button to start a request.",
			Style: t.Style{
				ForegroundColor: theme.TextMuted,
			},
		}
	}
}

func taskPhaseLabel(phase t.TaskPhase) string {
	switch phase {
	case t.TaskRunning:
		return "Running"
	case t.TaskSuccess:
		return "Success"
	case t.TaskError:
		return "Error"
	case t.TaskCancelled:
		return "Cancelled"
	default:
		return "Idle"
	}
}

func phaseVariant(phase t.TaskPhase) t.LabelVariant {
	switch phase {
	case t.TaskRunning:
		return t.LabelInfo
	case t.TaskSuccess:
		return t.LabelSuccess
	case t.TaskError:
		return t.LabelError
	case t.TaskCancelled:
		return t.LabelWarning
	default:
		return t.LabelSecondary
	}
}

func fetchJSON(ctx context.Context, url string) (*apiResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var payload map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&payload)
		message := payload["error"]
		if message == "" {
			message = resp.Status
		}
		return nil, fmt.Errorf("%s: %s", resp.Status, message)
	}

	var result apiResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func newExampleServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		delay := requestDelay(r, 1200*time.Millisecond)
		if !waitForDelay(r.Context(), delay) {
			return
		}

		name := r.URL.Query().Get("name")
		if name == "" {
			name = "World"
		}

		writeJSON(w, http.StatusOK, apiResult{
			Message:  "Hello " + name,
			Endpoint: r.URL.Path,
			DelayMS:  int(delay / time.Millisecond),
			ServedAt: time.Now().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		delay := requestDelay(r, 1200*time.Millisecond)
		if !waitForDelay(r.Context(), delay) {
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "simulated server failure",
		})
	})

	return httptest.NewServer(mux)
}

func requestDelay(r *http.Request, fallback time.Duration) time.Duration {
	raw := r.URL.Query().Get("delay_ms")
	if raw == "" {
		return fallback
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms < 0 {
		return fallback
	}
	return time.Duration(ms) * time.Millisecond
}

func waitForDelay(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func main() {
	server := newExampleServer()
	defer server.Close()

	app := NewAsyncTaskDemo(server.URL)
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
