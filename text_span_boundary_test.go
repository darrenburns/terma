package terma

import (
	"reflect"
	"testing"
)

func TestSpanBoundaryPreservation(t *testing.T) {
	for _, parts := range [][]string{{"🇬", "🇧"}, {"a", "\u0301"}, {"👨\u200d", "👩"}, {"1", "\ufe0f\u20e3"}, {"🇬", "🇧", "🇺", "🇸"}} {
		for _, wrap := range []WrapMode{WrapNone, WrapHard, WrapSoft} {
			for _, width := range []int{1, 2, 3, 8} {
				spans := make([]Span, len(parts))
				for i, s := range parts {
					spans[i].Text = s
				}
				widget := Text{Spans: spans, Wrap: wrap}
				got, want := widget.collectSpanLines(width, 5), widget.referenceCollectSpanLines(width, 5)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("parts=%q wrap=%d width=%d got=%v want=%v", parts, wrap, width, spanLineSummary(got), spanLineSummary(want))
				}
			}
		}
	}
}
