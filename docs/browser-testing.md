# Testing apps in a browser

`terma-browser` runs an existing app locally in a pseudo-terminal and displays it
in a browser using [xterm.js](https://xtermjs.org/). Apps need no code changes.
Keyboard, paste, mouse clicks, dragging, scrolling, and terminal resizing use the
same terminal input/output path as normal Terma sessions. The Go process runs on
the host; this is not a WebAssembly build.

## Launch an app

From the Terma repository:

```sh
go run ./cmd/terma-browser -- go run ./cmd/todo-app
```

Open the URL printed by the launcher, including its `#` fragment. By default it
chooses an available port. To choose one explicitly:

```sh
go run ./cmd/terma-browser -port 8080 -- go run ./cmd/list-demo
```

For use from another app's repository, install the launcher:

```sh
go install github.com/darrenburns/terma/cmd/terma-browser@latest
terma-browser -- go run .
```

The command inherits the launcher's working directory and environment, except
terminal capabilities are set for the browser: `TERM=xterm-256color`,
`COLORTERM=truecolor`, Kitty keyboard disabled, and `NO_COLOR` removed. Browser
assets are embedded in the Go executable: no Node installation or CDN connection
is needed. The host must be macOS, Linux, or BSD.

Each browser tab starts a separate process. Reloading or clicking **Restart app**
starts a fresh process. Closing a tab terminates its process group, including
children such as those started by `go run`. Ctrl+C in the launcher's terminal
stops the server and its sessions. A restart resets in-memory state only; apps
still have their usual access to files and services. Use test data when testing
apps that persist changes.

## Computer-use workflow

1. Start the launcher with the desired app command.
2. Open its printed URL in the computer-use browser.
3. For repeatable screenshots, insert `?cols=100&rows=30` before the `#` fragment.
   Both dimensions are required. Supported sizes are 2–500 columns and 2–300 rows.
   Without these parameters, the terminal fits the browser viewport.
4. Type into **Terminal input**, send keys such as Tab/Enter/arrows, and click or
   scroll at the app's visible coordinates. Browser-reserved shortcuts may be
   intercepted by the browser.
5. Inspect the screenshot or accessibility tree after the app updates. Terminal
   rows are exposed as text, but individual Terma widgets are not HTML buttons or
   inputs. Use screen coordinates for widget mouse interactions.
6. Restart the app for another session, or close the tab when finished.

The launcher binds only to `127.0.0.1`. A random URL token, exact Host validation,
and same-origin WebSocket checks restrict access to sessions. It is a local
development tool, not a public hosting service or an app sandbox.

## Bundled dependencies

Browser files in `cmd/terma-browser/web/vendor` come from `@xterm/xterm` 5.5.0
(`lib/xterm.js`, `css/xterm.css`, and `LICENSE` in the npm package). Keep the license
alongside the assets when upgrading. The server uses `creack/pty` for the terminal
and `gorilla/websocket` for transport.
