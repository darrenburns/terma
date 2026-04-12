package terma

import "testing"

func TestBuildTextAreaLayout_SoftWrapSpaceAtBoundaryKeepsCursorAtWrappedLineStart(t *testing.T) {
	graphemes := splitGraphemes("hello world")

	layout := buildTextAreaLayout(graphemes, WrapSoft, 5, 6)

	if len(layout.lines) != 2 {
		t.Fatalf("expected 2 wrapped lines, got %d", len(layout.lines))
	}
	if layout.cursorLine != 1 || layout.cursorCol != 0 {
		t.Fatalf("expected cursor at start of wrapped line, got line=%d col=%d", layout.cursorLine, layout.cursorCol)
	}
	if layout.lines[1].width != 5 {
		t.Fatalf("expected wrapped line width 5, got %d", layout.lines[1].width)
	}
}

func TestBuildTextAreaLayout_SoftWrapSpaceAtBoundaryDoesNotShiftCursorRight(t *testing.T) {
	graphemes := splitGraphemes("hello world")

	layout := buildTextAreaLayout(graphemes, WrapSoft, 5, 7)

	if layout.cursorLine != 1 || layout.cursorCol != 1 {
		t.Fatalf("expected cursor after \"w\" at line=1 col=1, got line=%d col=%d", layout.cursorLine, layout.cursorCol)
	}
}
