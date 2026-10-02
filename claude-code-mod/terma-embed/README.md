# terma-embed (experiment)

A Claude Code mod that runs a terma app in a pane, live and interactive.

```bash
claude --plugin-dir claude-code-mod/terma-embed   # from the repo root
```

Then type `/terma` in Claude Code. It builds `./cmd/terma-demos`, starts it, and docks it beside the
transcript. `/terma ./cmd/list-demo` runs another package, and `/terma stop` ends it. Click the pane
to give it the keyboard. Escape never reaches a pane, so ctrl+] sends Escape.

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
