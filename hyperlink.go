package terma

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// terminalLink converts a span's URL into the cell link written as OSC 8.
// OSC 8 URIs may only contain printable ASCII, so spaces and non-ASCII bytes
// are percent-encoded. A URL with a control character is dropped: it could
// end the escape sequence early and inject terminal commands.
func terminalLink(url string) uv.Link {
	if url == "" {
		return uv.Link{}
	}
	clean := true
	for i := 0; i < len(url); i++ {
		switch b := url[i]; {
		case b < 0x20 || b == 0x7f:
			return uv.Link{}
		case b == ' ' || b >= 0x80:
			clean = false
		}
	}
	if clean {
		return uv.Link{URL: url}
	}
	var sb strings.Builder
	for i := 0; i < len(url); i++ {
		if b := url[i]; b == ' ' || b >= 0x80 {
			const hex = "0123456789ABCDEF"
			sb.WriteByte('%')
			sb.WriteByte(hex[b>>4])
			sb.WriteByte(hex[b&0xf])
		} else {
			sb.WriteByte(b)
		}
	}
	return uv.Link{URL: sb.String()}
}
