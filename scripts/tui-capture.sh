#!/bin/bash
# Run a Terma program in a detached tmux session, send it keys, and capture the
# screen after startup and after each key as ANSI, SVG and (if rsvg-convert is
# installed) PNG.
#
#   scripts/tui-capture.sh [-s WxH] [-o DIR] [-d DELAY] <binary> [key...]
#
# Keys use tmux send-keys syntax: Tab, Enter, Up, C-b, M-x, or literal text.
# Extra environment for the program can be passed with TUI_ENV, e.g.
#   TUI_ENV=TERMA_DEBUG_OVERLAY=1 scripts/tui-capture.sh ./demo C-b
set -euo pipefail

size=100x30 out=snapshot-output/tui delay=0.6
while getopts "s:o:d:" opt; do
	case $opt in
	s) size=$OPTARG ;;
	o) out=$OPTARG ;;
	d) delay=$OPTARG ;;
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

# A private tmux server keeps the user's sessions untouched. TERM must not start
# with "tmux" or colour detection ignores COLORTERM and falls back to 256 colours.
tmux -L "$socket" -f /dev/null new-session -d -s "$session" -x "$width" -y "$height" \
	"env TERM=xterm-256color COLORTERM=truecolor ${TUI_ENV:-} $bin"
trap 'tmux -L "$socket" kill-server 2>/dev/null || true' EXIT

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
	echo "$name: ${2:-startup}"
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
