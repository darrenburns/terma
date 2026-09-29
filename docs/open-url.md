# Opening URLs

Use `terma.OpenURL` to open a web URL in the user's default browser. Apps can
replace their own platform switches and shell wrappers with this call:

```go
err := terma.OpenURL("https://example.com/docs")
if err != nil {
	// Show or log the error using your app's normal error handling.
}
```

The URL must be absolute, use HTTP or HTTPS, and contain a host. Encode spaces in
URLs as `%20`. The helper passes the URL as a single process argument and does
not invoke a shell. It uses `open` on macOS, `xdg-open` on Linux, and
`rundll32.exe url.dll,FileProtocolHandler` on Windows. Linux requires `xdg-open`
and a desktop session with a configured browser.

`OpenURL` waits for the launcher to exit and returns its error, including a
missing launcher or nonzero exit status. A successful call means the launcher
accepted the request; it cannot confirm that a browser loaded the page. The
helper does not suspend the terminal UI. Call it from a [background task](async.md)
to keep input and rendering responsive while the launcher runs.
