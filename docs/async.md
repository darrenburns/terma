# Async Tasks

Terma keeps rendering and input on a single UI loop. Slow work such as network requests, filesystem access, or subprocess calls should run in the background and publish results back to the UI.

Use `Task[T]` for one-shot async work:

- `Start(...)` runs the task in a goroutine
- `Running` and `Phase` expose loading state
- `Value` and `Err` expose the final result
- starting a new run cancels the old one and ignores stale completions

## Basic Pattern

If your UI-visible state is already signal-backed, you usually do not need `Dispatch`. The task signals themselves are enough to refresh the UI.

```go
type User struct {
	Name string `json:"name"`
}

type App struct {
	loadUser *terma.Task[*User]
}

func NewApp() *App {
	return &App{
		loadUser: terma.NewTask[*User](),
	}
}

func (a *App) fetchUser() {
	a.loadUser.Start(func(ctx context.Context) (*User, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/me", nil)
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var user User
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, err
		}

		return &user, nil
	})
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	children := []terma.Widget{
		terma.Button{
			Label:   "Load User",
			OnPress: a.fetchUser,
		},
	}

	if a.loadUser.Running.Get() {
		children = append(children, terma.Row{
			Children: []terma.Widget{
				terma.Text{Content: "Loading..."},
			},
		})
	}

	if a.loadUser.Phase.Get() == terma.TaskSuccess {
		user := a.loadUser.Value.Get()
		children = append(children, terma.Text{Content: "Hello " + user.Name})
	}

	if a.loadUser.Phase.Get() == terma.TaskError {
		children = append(children, terma.Text{Content: a.loadUser.Err.Get().Error()})
	}

	return terma.Column{Children: children}
}
```

## When To Use Dispatch

Use `Dispatch` when completion needs to touch plain Go fields or other UI-thread-only state.

```go
type App struct {
	loadUser *terma.Task[*User]
	status   terma.Signal[string]
	user     *User
}

func (a *App) fetchUser() {
	a.status.Set("Loading...")

	a.loadUser.Start(func(ctx context.Context) (*User, error) {
		user, err := fetchUserFromAPI(ctx)
		if err != nil {
			return nil, err
		}

		terma.Dispatch(func() {
			a.user = user
			a.status.Set("Loaded")
		})

		return user, nil
	})
}
```

Use this rule of thumb:

- If the UI reads signals, update the signals directly.
- If the UI reads plain struct fields, update those fields inside `Dispatch`.
- Do not block in `Build()`, `Render()`, or input handlers.

## Cancellation

Tasks receive a `context.Context`.

- Call `Cancel()` to stop the current run.
- Starting a new run cancels the previous one.
- If the app exits, app-bound task contexts are canceled too.
