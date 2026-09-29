package terma

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func TestSpanStreamingMatchesReference(t *testing.T) {
	styles := []SpanStyle{{}, {Bold: true}, {Italic: true, Foreground: Red}}
	corpus := []string{"", "a", "abc def", "   abc  def  ", "abc\n\ndef\n", "你好世界", "e\u0301e\u0301", "👨‍👩‍👧‍👦👨‍👩‍👧‍👦", "\t\r\n", strings.Repeat("x", 5000), strings.Repeat(" ", 200) + "abc"}
	rng := rand.New(rand.NewSource(91))
	tokens := []string{"a", "bc", " ", "  ", "\n", "你", "e\u0301", "👨‍👩‍👧‍👦", "\t"}
	for i := 0; i < 150; i++ {
		var s strings.Builder
		for j := 0; j < 60; j++ {
			s.WriteString(tokens[rng.Intn(len(tokens))])
		}
		corpus = append(corpus, s.String())
	}
	for _, wrap := range []WrapMode{WrapNone, WrapHard, WrapSoft} {
		for _, width := range []int{-1, 0, 1, 2, 3, 7, 80} {
			for _, height := range []int{-1, 0, 1, 2, 5, 24} {
				for i, content := range corpus {
					// Split spans at independent grapheme/token boundaries and change styles.
					parts := strings.SplitAfter(content, " ")
					spans := make([]Span, len(parts))
					for j, part := range parts {
						spans[j] = Span{Text: part, Style: styles[j%len(styles)]}
					}
					widget := Text{Spans: spans, Wrap: wrap}
					got, want := widget.collectSpanLines(width, height), widget.referenceCollectSpanLines(width, height)
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("wrap=%v width=%d height=%d corpus=%d text=%q\ngot: %#v\nwant:%#v", wrap, width, height, i, content, spanLineSummary(got), spanLineSummary(want))
					}
				}
			}
		}
	}
}

func TestSpanStreamingVisiblePrefixAllocations(t *testing.T) {
	for _, wrap := range []WrapMode{WrapNone, WrapHard, WrapSoft} {
		text := Text{Spans: []Span{{Text: strings.Repeat("x", 1<<20), Style: SpanStyle{Bold: true}}}, Wrap: wrap}
		allocations := testing.AllocsPerRun(5, func() {
			lines := text.collectSpanLines(80, 1)
			if len(lines) != 1 || lines[0].width != 80 {
				t.Fatalf("unexpected visible prefix: %#v", lines)
			}
		})
		if allocations > 40 {
			t.Fatalf("wrap=%v allocations=%f, want bounded visible-prefix allocation", wrap, allocations)
		}
	}
}

func BenchmarkSpanStreamingVisiblePrefix(b *testing.B) {
	for _, wrap := range []WrapMode{WrapNone, WrapHard, WrapSoft} {
		for _, size := range []int{4096, 1 << 20} {
			b.Run(fmt.Sprintf("wrap=%d/bytes=%d", wrap, size), func(b *testing.B) {
				widget := Text{Spans: []Span{{Text: strings.Repeat("x", size), Style: SpanStyle{Bold: true}}}, Wrap: wrap}
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					textPerfSpanLines = widget.collectSpanLines(80, 1)
				}
			})
		}
	}
}

func spanLineSummary(lines []lineData) []string {
	var out []string
	for _, line := range lines {
		var s strings.Builder
		for _, seg := range line.segments {
			s.WriteString(seg.span.Text)
		}
		out = append(out, fmt.Sprintf("%d:%q", line.width, s.String()))
	}
	return out
}
