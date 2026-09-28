package terma

import (
	"reflect"
	"testing"

	"github.com/darrenburns/terma/layout"
)

func TestTextWrappingPreservesContent(t *testing.T) {
	for _, test := range []struct {
		name    string
		content string
		width   int
		lines   []string
	}{
		{"ASCII", "abcdefgh", 3, []string{"abc", "def", "gh"}},
		{"spaces", "a   bc  ", 3, []string{"a  ", " bc", "  "}},
		{"newlines", "abc\n\ndef\n", 2, []string{"ab", "c", "", "de", "f", ""}},
		{"wide", "你好世界", 3, []string{"你", "好", "世", "界"}},
		{"combining", "e\u0301e\u0301e\u0301", 2, []string{"e\u0301e\u0301", "e\u0301"}},
		{"emoji", "👨‍👩‍👧‍👦👨‍👩‍👧‍👦", 2, []string{"👨‍👩‍👧‍👦", "👨‍👩‍👧‍👦"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := wrapText(test.content, test.width, WrapHard); !reflect.DeepEqual(got, test.lines) {
				t.Fatalf("wrapText() = %q, want %q", got, test.lines)
			}
			_, height := layout.MeasureText(test.content, layout.WrapChar, test.width)
			if height != len(test.lines) {
				t.Fatalf("layout height = %d, want %d", height, len(test.lines))
			}
		})
	}
}

func TestTextSpanCollectionAllocations(t *testing.T) {
	text := Text{Spans: []Span{{Text: "A status message with enough characters to expose per-character string copies.", Style: SpanStyle{Bold: true}}}}
	allocations := testing.AllocsPerRun(10, func() {
		lines := text.collectSpanLines(80, 1)
		if len(lines) != 1 || len(lines[0].segments) != 1 || lines[0].segments[0].span != text.Spans[0] {
			t.Fatal("span content or style changed")
		}
	})
	if allocations > 30 {
		t.Fatalf("span collection allocated %.0f times; expected at most 30, without a string allocation per character", allocations)
	}
}

func TestTextWrappingNarrowWideGrapheme(t *testing.T) {
	for _, wrap := range []WrapMode{WrapHard, WrapSoft} {
		got := wrapText("你好", 1, wrap)
		want := []string{"你", "好"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("wrapText() = %q, want %q", got, want)
		}
		_, height := layout.MeasureText("你好", toLayoutWrapMode(wrap), 1)
		if height != len(want) {
			t.Fatalf("layout height = %d, want %d", height, len(want))
		}
	}
}
