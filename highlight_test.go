package terma

import (
	"regexp"
	"testing"
)

// --- Unit Tests for Highlight Helper Functions ---

func TestBuildHighlightMap(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		result := buildHighlightMap(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("single highlight", func(t *testing.T) {
		highlights := []TextHighlight{
			{Start: 2, End: 5, Style: SpanStyle{Bold: true}},
		}
		result := buildHighlightMap(highlights)
		if len(result) != 3 {
			t.Errorf("expected 3 entries, got %d", len(result))
		}
		for i := 2; i < 5; i++ {
			if !result[i].Bold {
				t.Errorf("expected Bold at index %d", i)
			}
		}
	})

	t.Run("overlapping highlights - later wins", func(t *testing.T) {
		highlights := []TextHighlight{
			{Start: 0, End: 5, Style: SpanStyle{Bold: true}},
			{Start: 3, End: 8, Style: SpanStyle{Italic: true}},
		}
		result := buildHighlightMap(highlights)
		// 0-2 should be bold only
		for i := 0; i < 3; i++ {
			if !result[i].Bold || result[i].Italic {
				t.Errorf("index %d: expected Bold only, got Bold=%v Italic=%v", i, result[i].Bold, result[i].Italic)
			}
		}
		// 3-4 should be italic (later highlight overrides)
		for i := 3; i < 5; i++ {
			if result[i].Bold || !result[i].Italic {
				t.Errorf("index %d: expected Italic only, got Bold=%v Italic=%v", i, result[i].Bold, result[i].Italic)
			}
		}
	})
}

func TestBuildLineHighlightMap(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		result := buildLineHighlightMap(nil, 5)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("single line", func(t *testing.T) {
		highlights := []LineHighlight{
			{StartLine: 2, EndLine: 3, Style: Style{Bold: true}},
		}
		result := buildLineHighlightMap(highlights, 5)
		if len(result) != 1 {
			t.Errorf("expected 1 entry, got %d", len(result))
		}
		if !result[2].Bold {
			t.Errorf("expected Bold at line 2")
		}
	})

	t.Run("range to end (-1)", func(t *testing.T) {
		highlights := []LineHighlight{
			{StartLine: 2, EndLine: -1, Style: Style{Bold: true}},
		}
		result := buildLineHighlightMap(highlights, 5)
		if len(result) != 3 { // lines 2, 3, 4
			t.Errorf("expected 3 entries, got %d", len(result))
		}
		for i := 2; i < 5; i++ {
			if !result[i].Bold {
				t.Errorf("expected Bold at line %d", i)
			}
		}
	})
}

func TestApplySpanStyle(t *testing.T) {
	base := Style{
		ForegroundColor: RGB(255, 255, 255),
		Bold:            false,
	}

	span := SpanStyle{
		Foreground: RGB(255, 0, 0),
		Bold:       true,
		Italic:     true,
	}

	result := applySpanStyle(base, span)

	if !result.Bold {
		t.Error("expected Bold to be true")
	}
	if !result.Italic {
		t.Error("expected Italic to be true")
	}
	// Foreground should be overridden
	if result.ForegroundColor == nil || !result.ForegroundColor.IsSet() {
		t.Error("expected ForegroundColor to be set")
	}
}

// --- HighlighterFunc Tests ---

func TestHighlighterFunc(t *testing.T) {
	fn := HighlighterFunc(func(text string, graphemes []string) []TextHighlight {
		return []TextHighlight{
			{Start: 0, End: 1, Style: SpanStyle{Bold: true}},
		}
	})

	result := fn.Highlight("test", []string{"t", "e", "s", "t"})
	if len(result) != 1 {
		t.Errorf("expected 1 highlight, got %d", len(result))
	}
}

// --- Test Highlighter Implementation ---

// hashtagHighlighter highlights #hashtags with a given style.
type hashtagHighlighter struct {
	style SpanStyle
}

func (h *hashtagHighlighter) Highlight(text string, graphemes []string) []TextHighlight {
	pattern := regexp.MustCompile(`#\w+`)
	matches := pattern.FindAllStringIndex(text, -1)

	var highlights []TextHighlight
	for _, match := range matches {
		// Convert byte indices to grapheme indices
		startByte := match[0]
		endByte := match[1]

		// Find grapheme indices
		bytePos := 0
		startGrapheme := -1
		endGrapheme := -1
		for i, g := range graphemes {
			if bytePos == startByte {
				startGrapheme = i
			}
			bytePos += len(g)
			if bytePos == endByte {
				endGrapheme = i + 1
				break
			}
		}
		if startGrapheme >= 0 && endGrapheme > startGrapheme {
			highlights = append(highlights, TextHighlight{
				Start: startGrapheme,
				End:   endGrapheme,
				Style: h.style,
			})
		}
	}
	return highlights
}

// --- Snapshot Tests ---

func TestSnapshot_TextInput_Highlighting(t *testing.T) {
	state := NewTextInputState("hello #world today")
	state.CursorIndex.Set(0)

	highlighter := &hashtagHighlighter{
		style: SpanStyle{
			Foreground: RGB(0, 150, 255), // Blue
			Bold:       true,
		},
	}

	widget := TextInput{
		ID:          "textinput-highlight",
		State:       state,
		Highlighter: highlighter,
		Width:       Cells(25),
	}

	AssertSnapshot(t, widget, 25, 1,
		"TextInput with #world highlighted in blue bold.")
}

func TestSnapshot_TextInput_MultipleHighlights(t *testing.T) {
	state := NewTextInputState("check #tag1 and #tag2 now")
	state.CursorIndex.Set(0)

	highlighter := &hashtagHighlighter{
		style: SpanStyle{
			Foreground: RGB(255, 100, 0), // Orange
			Italic:     true,
		},
	}

	widget := TextInput{
		ID:          "textinput-multi-highlight",
		State:       state,
		Highlighter: highlighter,
		Width:       Cells(30),
	}

	AssertSnapshot(t, widget, 30, 1,
		"TextInput with two hashtags highlighted in orange italic.")
}

func TestSnapshot_TextInput_HighlightWithScroll(t *testing.T) {
	state := NewTextInputState("prefix #highlighted suffix text")
	// Move cursor to end to force scroll
	state.CursorIndex.Set(len(splitGraphemes(state.GetText())))

	highlighter := &hashtagHighlighter{
		style: SpanStyle{
			Foreground: RGB(0, 255, 0), // Green
		},
	}

	widget := TextInput{
		ID:          "textinput-highlight-scroll",
		State:       state,
		Highlighter: highlighter,
		Width:       Cells(15), // Narrow to force scroll
	}

	AssertSnapshot(t, widget, 15, 1,
		"TextInput scrolled right with highlight partially/fully visible.")
}

func TestSnapshot_TextInput_HighlightAtCursor(t *testing.T) {
	state := NewTextInputState("hello #tag world")
	// Position cursor in the middle of the hashtag
	state.CursorIndex.Set(8) // On 'a' of #tag

	highlighter := &hashtagHighlighter{
		style: SpanStyle{
			Foreground: RGB(255, 0, 255), // Magenta
			Bold:       true,
		},
	}

	widget := TextInput{
		ID:          "textinput-highlight-cursor",
		State:       state,
		Highlighter: highlighter,
		Width:       Cells(20),
	}

	AssertSnapshot(t, widget, 20, 1,
		"TextInput with cursor on highlighted text. Cursor (reverse) takes precedence over highlight.")
}

func TestSnapshot_TextArea_Highlighting(t *testing.T) {
	state := NewTextAreaState("hello #world\nthis is a #test")
	state.WrapMode.Set(WrapSoft)
	state.CursorIndex.Set(0)

	highlighter := &hashtagHighlighter{
		style: SpanStyle{
			Foreground: RGB(0, 150, 255), // Blue
			Bold:       true,
		},
	}

	widget := TextArea{
		ID:          "textarea-highlight",
		State:       state,
		Highlighter: highlighter,
		Width:       Cells(20),
		Height:      Cells(3),
	}

	AssertSnapshot(t, widget, 20, 3,
		"TextArea with #world and #test highlighted in blue bold.")
}

func TestSnapshot_TextArea_LineHighlight(t *testing.T) {
	state := NewTextAreaState("line 0\nline 1\nline 2\nline 3")
	state.WrapMode.Set(WrapSoft)
	state.CursorIndex.Set(0)

	widget := TextArea{
		ID:     "textarea-line-highlight",
		State:  state,
		Width:  Cells(15),
		Height: Cells(4),
		LineHighlights: []LineHighlight{
			{StartLine: 1, EndLine: 2, Style: Style{BackgroundColor: RGB(50, 50, 100)}},
		},
	}

	AssertSnapshot(t, widget, 15, 4,
		"TextArea with line 1 highlighted with blue background.")
}

func TestSnapshot_TextArea_LineHighlightRange(t *testing.T) {
	state := NewTextAreaState("line 0\nline 1\nline 2\nline 3\nline 4")
	state.WrapMode.Set(WrapSoft)
	state.CursorIndex.Set(0)

	widget := TextArea{
		ID:     "textarea-line-highlight-range",
		State:  state,
		Width:  Cells(15),
		Height: Cells(5),
		LineHighlights: []LineHighlight{
			{StartLine: 1, EndLine: 4, Style: Style{BackgroundColor: RGB(80, 40, 40)}},
		},
	}

	AssertSnapshot(t, widget, 15, 5,
		"TextArea with lines 1-3 highlighted with red-ish background.")
}

func TestSnapshot_TextArea_CombinedHighlights(t *testing.T) {
	state := NewTextAreaState("check #tag here\nerror line\nnormal line")
	state.WrapMode.Set(WrapSoft)
	state.CursorIndex.Set(0)

	highlighter := &hashtagHighlighter{
		style: SpanStyle{
			Foreground: RGB(100, 200, 255), // Light blue
			Bold:       true,
		},
	}

	widget := TextArea{
		ID:          "textarea-combined-highlight",
		State:       state,
		Highlighter: highlighter,
		Width:       Cells(20),
		Height:      Cells(3),
		LineHighlights: []LineHighlight{
			// Error line with red background
			{StartLine: 1, EndLine: 2, Style: Style{BackgroundColor: RGB(100, 30, 30)}},
		},
	}

	AssertSnapshot(t, widget, 20, 3,
		"TextArea with #tag text highlighted AND line 1 with red background (error line).")
}

func TestSnapshot_TextArea_HighlightWithSelection(t *testing.T) {
	state := NewTextAreaState("select #highlighted text")
	state.WrapMode.Set(WrapSoft)
	state.SetSelectionAnchor(7)  // Start of #highlighted
	state.CursorIndex.Set(19)    // End of #highlighted

	highlighter := &hashtagHighlighter{
		style: SpanStyle{
			Foreground: RGB(255, 200, 0), // Yellow
			Bold:       true,
		},
	}

	widget := TextArea{
		ID:          "textarea-highlight-selection",
		State:       state,
		Highlighter: highlighter,
		Width:       Cells(30),
		Height:      Cells(2),
	}

	AssertSnapshot(t, widget, 30, 2,
		"TextArea with selection over highlighted text. Selection background takes precedence.")
}

func TestSnapshot_TextArea_HighlightWithScroll(t *testing.T) {
	state := NewTextAreaState("line 0\nline 1\nline 2 #tag\nline 3\nline 4")
	state.WrapMode.Set(WrapSoft)
	// Scroll down by positioning cursor at the end
	state.CursorIndex.Set(len(splitGraphemes(state.GetText())))

	highlighter := &hashtagHighlighter{
		style: SpanStyle{
			Foreground: RGB(0, 255, 100), // Green
		},
	}

	widget := TextArea{
		ID:          "textarea-highlight-scroll",
		State:       state,
		Highlighter: highlighter,
		Width:       Cells(15),
		Height:      Cells(3), // Only 3 lines visible
	}

	AssertSnapshot(t, widget, 15, 3,
		"TextArea scrolled to show bottom lines with #tag highlighted.")
}

func TestSnapshot_TextArea_LineHighlightToEnd(t *testing.T) {
	state := NewTextAreaState("line 0\nline 1\nline 2\nline 3")
	state.WrapMode.Set(WrapSoft)
	state.CursorIndex.Set(0)

	widget := TextArea{
		ID:     "textarea-line-highlight-to-end",
		State:  state,
		Width:  Cells(15),
		Height: Cells(4),
		LineHighlights: []LineHighlight{
			// EndLine -1 means highlight to end
			{StartLine: 2, EndLine: -1, Style: Style{BackgroundColor: RGB(40, 80, 40)}},
		},
	}

	AssertSnapshot(t, widget, 15, 4,
		"TextArea with lines 2 onwards highlighted with green background (EndLine=-1).")
}
