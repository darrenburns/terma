package terma

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/colorprofile"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// Replacements are not transitive in Go modules. Check the actual linked
// Ultraviolet implementation so consumers missing the fork get safe fallback.
var kittyColorEncodingSafe = func() bool {
	a := uv.Style{Fg: ansi.IndexedColor(0), UnderlineColor: ansi.IndexedColor(0)}
	b := uv.Style{Fg: ansi.IndexedColor(16), UnderlineColor: ansi.IndexedColor(16)}
	return !a.Equal(&b) && strings.Contains(b.Diff(&a), "38;5;16") && strings.Contains(b.Diff(&a), "58;5;16")
}()

const imageProbeID = 2147483647
const imageDetectionTimeout = 350 * time.Millisecond

type imageDetector struct {
	requested                                                   string
	started, done                                               bool
	deadline                                                    time.Time
	graphics, sixel, syncOutput, knownPlaceholders, multiplexer bool
	profile                                                     colorprofile.Profile
}

func newImageDetector(profile colorprofile.Profile) *imageDetector {
	protocol := strings.ToLower(os.Getenv("TERMA_IMAGE_PROTOCOL"))
	switch protocol {
	case "kitty", "sixel", "blocks":
	default:
		protocol = "auto"
	}
	term := os.Getenv("TERM")
	return &imageDetector{requested: protocol, profile: profile, multiplexer: os.Getenv("TMUX") != "" || os.Getenv("STY") != "" || strings.HasPrefix(term, "screen") || strings.HasPrefix(term, "tmux")}
}
func (d *imageDetector) query(now time.Time) string {
	if d.started {
		return ""
	}
	d.started = true
	d.deadline = now.Add(imageDetectionTimeout)
	if d.requested == "blocks" || d.requested == "auto" && d.multiplexer {
		d.done = true
		return ""
	}
	// DA1 is deliberately last: terminals reply in order; it completes the
	// probe without waiting for unsupported queries to time out.
	return fmt.Sprintf("\x1b_Ga=q,i=%d,t=d,f=24,s=1,v=1;AAAA\x1b\\", imageProbeID) + requestCellSize + "\x1b[14t\x1b[>q" + ansi.RequestModeSynchronizedOutput + "\x1b[c"
}

var ghosttyIdentity = regexp.MustCompile(`(?i)^ghostty(?:[ (]|$)`)
var kittyVersion = regexp.MustCompile(`(?i)^kitty[ (]+([0-9]+)\.([0-9]+)(?:\.[0-9]+)?`)

func (d *imageDetector) handle(event uv.Event) {
	switch e := event.(type) {
	case uv.KittyGraphicsEvent:
		if e.Options.ID == imageProbeID && string(e.Payload) == "OK" {
			d.graphics = true
		}
	case uv.TerminalVersionEvent:
		name := strings.ToLower(e.Name)
		if strings.Contains(name, "tmux") || strings.Contains(name, "screen") {
			d.multiplexer = true
		}
		d.knownPlaceholders = ghosttyIdentity.MatchString(e.Name)
		if m := kittyVersion.FindStringSubmatch(e.Name); m != nil {
			major, _ := strconv.Atoi(m[1])
			minor, _ := strconv.Atoi(m[2])
			d.knownPlaceholders = major > 0 || minor >= 28
		}
	case uv.PrimaryDeviceAttributesEvent:
		for _, a := range e {
			if a == 4 {
				d.sixel = true
			}
		}
		if d.started {
			d.done = true
		}
	case uv.ModeReportEvent:
		if e.Mode == ansi.ModeSynchronizedOutput {
			d.syncOutput = !e.Value.IsNotRecognized() && !e.Value.IsPermanentlyReset()
		}
	}
}
func (d *imageDetector) protocol(now time.Time, geometry bool) string {
	if d.started && !d.done && !now.Before(d.deadline) {
		d.done = true
	}
	if !d.done || !geometry {
		return "blocks"
	}
	switch d.requested {
	case "blocks":
		return "blocks"
	case "kitty":
		if d.profile >= colorprofile.ANSI256 && kittyColorEncodingSafe {
			return "kitty"
		}
		return "blocks"
	case "sixel":
		return "sixel"
	}
	if d.multiplexer {
		return "blocks"
	}
	if d.graphics && d.knownPlaceholders && d.profile >= colorprofile.ANSI256 && kittyColorEncodingSafe {
		return "kitty"
	}
	if d.sixel {
		return "sixel"
	}
	return "blocks"
}
func explicitImageCellSize() (int, int) {
	parts := strings.Split(os.Getenv("TERMA_IMAGE_CELL_SIZE"), "x")
	if len(parts) != 2 {
		return 0, 0
	}
	w, e1 := strconv.Atoi(parts[0])
	h, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || w <= 0 || h <= 0 || w > 4096 || h > 4096 {
		return 0, 0
	}
	return w, h
}
