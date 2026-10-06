#!/bin/bash
# Run a Terma program in a detached tmux session, send it keys, and capture the
# screen after startup and after each key as ANSI, SVG and (if rsvg-convert is
# installed) PNG.
#
#   scripts/tui-capture.sh [-s WxH] [-o DIR] [-d DELAY] [-b] <binary> [key...]
#
# -b captures the scrollback above the screen too, as a separate
# NN.scrollback.ansi/.svg/.png per step (for inline apps, see RunInline).
# The binary may be a shell command line, e.g. "sh -c 'seq 5; exec ./demo'".
#
# Keys use tmux send-keys syntax: Tab, Enter, Up, C-b, M-x, or literal text.
# Extra environment for the program can be passed with TUI_ENV, e.g.
#   TUI_ENV=TERMA_DEBUG_OVERLAY=1 scripts/tui-capture.sh ./demo C-b
set -euo pipefail

size=100x30 out=snapshot-output/tui delay=0.6 scrollback=0
while getopts "s:o:d:b" opt; do
	case $opt in
	s) size=$OPTARG ;;
	o) out=$OPTARG ;;
	d) delay=$OPTARG ;;
	b) scrollback=1 ;;
	*) exit 2 ;;
	esac
done
shift $((OPTIND - 1))
bin=$1
shift
width=${size%x*} height=${size#*x}
root=$(cd "$(dirname "$0")/.." && pwd)
session="terma-capture-$$"
socket="terma-capture-$$"

mkdir -p "$out"
converter="$out/.ansi-to-svg"
(cd "$root" && go build -o "$converter" ./cmd/ansi-to-svg)

# Programs run with a throwaway home and XDG directories, so anything they save
# (settings, app state) never touches the user's real files.
sandbox=$(mktemp -d)
mkdir -p "$sandbox/config" "$sandbox/state" "$sandbox/data" "$sandbox/cache"

# A private tmux server keeps the user's sessions untouched. TERM must not start
# with "tmux" or colour detection ignores COLORTERM and falls back to 256 colours.
# NO_COLOR is dropped so captures always show the real colours.
tmux -L "$socket" -f /dev/null new-session -d -s "$session" -x "$width" -y "$height" \
	"env -u NO_COLOR TERM=xterm-256color COLORTERM=truecolor HOME=$sandbox XDG_CONFIG_HOME=$sandbox/config XDG_STATE_HOME=$sandbox/state XDG_DATA_HOME=$sandbox/data XDG_CACHE_HOME=$sandbox/cache ${TUI_ENV:-} $bin"
trap 'tmux -L "$socket" kill-server 2>/dev/null || true; rm -rf "$sandbox"' EXIT

capture() {
	local name
	name=$(printf "%s/%02d" "$out" "$1")
	# -N keeps trailing spaces; without it, cells whose only content is a
	# background colour are dropped and render as black.
	tmux -L "$socket" capture-pane -p -e -N -t "$session" >"$name.ansi"
	"$converter" -w "$width" -h "$height" <"$name.ansi" >"$name.svg"
	if command -v rsvg-convert >/dev/null; then
		rsvg-convert -z 1.5 "$name.svg" -o "$name.png"
	fi
	if [ "$scrollback" = 1 ]; then
		tmux -L "$socket" capture-pane -p -e -N -S - -t "$session" >"$name.scrollback.ansi"
		local rows
		rows=$(wc -l <"$name.scrollback.ansi" | tr -d ' ')
		"$converter" -w "$width" -h "$rows" <"$name.scrollback.ansi" >"$name.scrollback.svg"
		if command -v rsvg-convert >/dev/null; then
			rsvg-convert -z 1.5 "$name.scrollback.svg" -o "$name.scrollback.png"
		fi
	fi
	echo "$name: ${2:-startup} (cursor at $(tmux -L "$socket" display -p -t "$session" '#{cursor_x},#{cursor_y}'))"
}

sleep 1.5
capture 0
step=0
for key in "$@"; do
	step=$((step + 1))
	tmux -L "$socket" send-keys -t "$session" "$key"
	sleep "$delay"
	capture "$step" "$key"
done
