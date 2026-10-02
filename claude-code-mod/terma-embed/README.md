# terma-embed (experiment)

A Claude Code mod that runs a terma app in a pane, live and interactive.

```bash
claude --plugin-dir claude-code-mod/terma-embed   # from the repo root
```

Then type `/terma` in Claude Code. It builds `./cmd/terma-demos`, starts it, and docks it beside the
transcript. `/terma stop` ends it. Click the pane to give it the keyboard. Escape never reaches a
pane, so ctrl+] sends Escape.

`/terma` takes what to run, then any arguments for the app:

| Argument | Runs |
|----------|------|
| (none) | `./cmd/terma-demos` in the session's directory |
| `./cmd/list-demo` | another package in the session's directory |
| `~/code/posting/cmd/posting` | a main package anywhere on disk, built and run in that directory |
| `/path/to/binary` | a binary you built yourself with a terma that has embedded mode |

An app outside this repo usually depends on a terma without embedded mode. When it does, the mod
builds it with `GOWORK=/tmp/terma-embed.go.work`, a workspace that uses the app's module and
replaces terma with the session directory's terma. The app's own `go.mod` and `go.work` stay
untouched, so start Claude Code in a terma checkout that has `embed.go`.

## How it works

- `TERMA_EMBED_SOCKET` switches `terma.Run` into embedded mode (`embed.go`). The app renders into
  memory, writes each changed frame to stdout as one JSON line (runs of styled text), and serves
  `POST /events` (keys, mouse, resize) on that Unix socket.
- `hooks/register.tsx` builds and spawns the app, stores the latest frame in `$.state`, draws the
  pane, and relays input to the socket with `$.http.fetch({ socketPath })`.
- `hooks/screen.tsx` is the `Client` surface module. It draws the frame as `Text` runs and posts
  key, pointer and resize events. Each post repeats the undelivered tail of its outbox, because a
  post made in the same frame as another replaces it.

Checks: `claude plugin validate`, `claude plugin test` (from this folder), and
`go test -run 'TestEmbed|TestEncodeFrame' .` from the repo root.
