package terma

import (
	"image/color"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestEmbedKeys_SplitsTextReadInOneBurst(t *testing.T) {
	keys := embedKeys(embedEvent{Type: "key", Key: "Hi!"})
	if len(keys) != 3 {
		t.Fatalf("got %d keys, want 3", len(keys))
	}
	text := ""
	for _, k := range keys {
		text += k.Text
	}
	if text != "Hi!" {
		t.Fatalf("typed %q, want %q", text, "Hi!")
	}
	if !keys[0].MatchString("shift+h") {
		t.Fatalf("first key %q, want shift+h", keys[0].String())
	}
}

func TestEmbedKeys_NamedAndModifiedKeys(t *testing.T) {
	cases := map[string]embedEvent{
		"up":        {Key: "up"},
		"enter":     {Key: "return"},
		"shift+tab": {Key: "tab", Shift: true},
		"ctrl+c":    {Key: "c", Ctrl: true},
		"esc":       {Key: "escape"},
	}
	for want, e := range cases {
		keys := embedKeys(e)
		if len(keys) != 1 || !keys[0].MatchString(want) {
			t.Errorf("%+v gave %v, want one %s", e, keys, want)
		}
		if e.Ctrl && keys[0].Text != "" {
			t.Errorf("%s inserts text %q", want, keys[0].Text)
		}
	}
}

func TestEncodeFrame_RunsOfEqualStyle(t *testing.T) {
	buf := uv.NewBuffer(6, 1)
	bold := uv.Style{Attrs: uv.AttrBold, Fg: color.RGBA{R: 0xff, A: 0xff}}
	buf.SetCell(0, 0, &uv.Cell{Content: "a", Width: 1})
	buf.SetCell(1, 0, &uv.Cell{Content: "b", Width: 1})
	buf.SetCell(2, 0, &uv.Cell{Content: "世", Width: 2, Style: bold})
	buf.SetCell(4, 0, &uv.Cell{Content: "c", Width: 1, Style: bold})

	f := encodeFrame(buf, 6, 1)
	line := f.Lines[0]
	if len(line) != 3 {
		t.Fatalf("got runs %v, want 3", line)
	}
	if line[0][0] != "ab" || line[1][0] != "世c" || line[2][0] != " " {
		t.Fatalf("got runs %v", line)
	}
	style := f.Styles[line[1][1].(int)]
	if !style.Bold || style.Fg == "" {
		t.Fatalf("bold run has style %+v", style)
	}
}
