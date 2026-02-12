package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// RenderedLineKind is the visual category of a rendered line.
type RenderedLineKind int

const (
	RenderedLineFileHeader RenderedLineKind = iota
	RenderedLineHunkHeader
	RenderedLineContext
	RenderedLineAdd
	RenderedLineRemove
	RenderedLineMeta
)

// TokenRole is a semantic token role used to map to theme styles.
type TokenRole int

const (
	TokenRoleOldLineNumber TokenRole = iota
	TokenRoleNewLineNumber
	TokenRoleDiffPrefixAdd
	TokenRoleDiffPrefixRemove
	TokenRoleDiffPrefixContext
	TokenRoleDiffFileHeader
	TokenRoleDiffHunkHeader
	TokenRoleDiffMeta
	TokenRoleSyntaxPlain
	TokenRoleSyntaxKeyword
	TokenRoleSyntaxType
	TokenRoleSyntaxFunction
	TokenRoleSyntaxString
	TokenRoleSyntaxNumber
	TokenRoleSyntaxComment
	TokenRoleSyntaxPunctuation
)

// HighlightToken identifies a byte range in rendered text and its semantic role.
type HighlightToken struct {
	StartByte int
	EndByte   int
	Role      TokenRole
}

// RenderedFile is the display model for one file diff.
type RenderedFile struct {
	Text      string
	Tokens    []HighlightToken
	LineKinds []RenderedLineKind
}

type renderLine struct {
	Kind    RenderedLineKind
	Prefix  string
	Content string
	OldLine int
	NewLine int
}

type syntaxToken struct {
	Start int
	End   int
	Role  TokenRole
}

func buildRenderedFile(file *DiffFile) *RenderedFile {
	if file == nil {
		return nil
	}

	lines := buildRenderLines(file)
	if len(lines) == 0 {
		lines = []renderLine{{Kind: RenderedLineMeta, Prefix: " ", Content: "No changes to render"}}
	}

	oldWidth, newWidth := lineNumberWidths(lines)
	lexer := chooseLexer(file)

	var builder strings.Builder
	tokens := make([]HighlightToken, 0, len(lines)*3)
	lineKinds := make([]RenderedLineKind, 0, len(lines))

	for lineIdx, line := range lines {
		oldText := lineNumberText(line.OldLine, oldWidth)
		newText := lineNumberText(line.NewLine, newWidth)

		oldStart := builder.Len()
		builder.WriteString(oldText)
		oldEnd := builder.Len()

		builder.WriteByte(' ')

		newStart := builder.Len()
		builder.WriteString(newText)
		newEnd := builder.Len()

		builder.WriteByte(' ')

		prefixStart := builder.Len()
		builder.WriteString(line.Prefix)
		prefixEnd := builder.Len()

		builder.WriteByte(' ')

		contentStart := builder.Len()
		builder.WriteString(line.Content)
		contentEnd := builder.Len()

		if lineIdx < len(lines)-1 {
			builder.WriteByte('\n')
		}

		lineKinds = append(lineKinds, line.Kind)

		if strings.TrimSpace(oldText) != "" {
			tokens = append(tokens, HighlightToken{StartByte: oldStart, EndByte: oldEnd, Role: TokenRoleOldLineNumber})
		}
		if strings.TrimSpace(newText) != "" {
			tokens = append(tokens, HighlightToken{StartByte: newStart, EndByte: newEnd, Role: TokenRoleNewLineNumber})
		}

		if prefixRole, ok := prefixRoleForLine(line.Kind); ok {
			tokens = append(tokens, HighlightToken{StartByte: prefixStart, EndByte: prefixEnd, Role: prefixRole})
		}

		switch line.Kind {
		case RenderedLineFileHeader:
			tokens = append(tokens, HighlightToken{StartByte: contentStart, EndByte: contentEnd, Role: TokenRoleDiffFileHeader})
		case RenderedLineHunkHeader:
			tokens = append(tokens, HighlightToken{StartByte: contentStart, EndByte: contentEnd, Role: TokenRoleDiffHunkHeader})
		case RenderedLineMeta:
			tokens = append(tokens, HighlightToken{StartByte: contentStart, EndByte: contentEnd, Role: TokenRoleDiffMeta})
		case RenderedLineContext, RenderedLineAdd, RenderedLineRemove:
			if lexer != nil && line.Content != "" {
				for _, token := range lexLine(lexer, line.Content) {
					tokens = append(tokens, HighlightToken{
						StartByte: contentStart + token.Start,
						EndByte:   contentStart + token.End,
						Role:      token.Role,
					})
				}
			}
		}

	}

	return &RenderedFile{
		Text:      builder.String(),
		Tokens:    tokens,
		LineKinds: lineKinds,
	}
}

func buildRenderLines(file *DiffFile) []renderLine {
	lines := make([]renderLine, 0, len(file.Headers)+len(file.Hunks)*8)
	for _, header := range file.Headers {
		if header == "" {
			continue
		}
		lines = append(lines, renderLine{
			Kind:    RenderedLineFileHeader,
			Prefix:  " ",
			Content: header,
		})
	}

	for _, hunk := range file.Hunks {
		lines = append(lines, renderLine{
			Kind:    RenderedLineHunkHeader,
			Prefix:  " ",
			Content: hunk.Header,
		})

		for _, line := range hunk.Lines {
			rendered := renderLine{
				Prefix:  " ",
				Content: line.Content,
				OldLine: line.OldLine,
				NewLine: line.NewLine,
			}
			switch line.Kind {
			case DiffLineContext:
				rendered.Kind = RenderedLineContext
				rendered.Prefix = " "
			case DiffLineAdd:
				rendered.Kind = RenderedLineAdd
				rendered.Prefix = "+"
			case DiffLineRemove:
				rendered.Kind = RenderedLineRemove
				rendered.Prefix = "-"
			default:
				rendered.Kind = RenderedLineMeta
				rendered.Prefix = " "
			}
			lines = append(lines, rendered)
		}
	}

	if len(lines) == 0 {
		lines = append(lines, renderLine{
			Kind:    RenderedLineMeta,
			Prefix:  " ",
			Content: "No displayable content",
		})
	}
	return lines
}

func lineNumberWidths(lines []renderLine) (oldWidth int, newWidth int) {
	maxOld := 1
	maxNew := 1
	for _, line := range lines {
		if line.OldLine > maxOld {
			maxOld = line.OldLine
		}
		if line.NewLine > maxNew {
			maxNew = line.NewLine
		}
	}
	return len(strconv.Itoa(maxOld)), len(strconv.Itoa(maxNew))
}

func lineNumberText(number int, width int) string {
	if number <= 0 {
		return strings.Repeat(" ", width)
	}
	return fmt.Sprintf("%*d", width, number)
}

func prefixRoleForLine(kind RenderedLineKind) (TokenRole, bool) {
	switch kind {
	case RenderedLineAdd:
		return TokenRoleDiffPrefixAdd, true
	case RenderedLineRemove:
		return TokenRoleDiffPrefixRemove, true
	case RenderedLineContext:
		return TokenRoleDiffPrefixContext, true
	default:
		return 0, false
	}
}

func chooseLexer(file *DiffFile) chroma.Lexer {
	path := file.NewPath
	if path == "" {
		path = file.OldPath
	}
	if path == "" {
		return nil
	}
	lexer := lexers.Match(path)
	if lexer == nil {
		return nil
	}
	return chroma.Coalesce(lexer)
}

func lexLine(lexer chroma.Lexer, content string) []syntaxToken {
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return nil
	}

	result := make([]syntaxToken, 0, 8)
	offset := 0
	for token := iterator(); token != chroma.EOF; token = iterator() {
		length := len(token.Value)
		if length == 0 {
			continue
		}
		role := tokenRoleFromChroma(token.Type)
		result = append(result, syntaxToken{Start: offset, End: offset + length, Role: role})
		offset += length
	}

	if offset < len(content) {
		result = append(result, syntaxToken{Start: offset, End: len(content), Role: TokenRoleSyntaxPlain})
	}

	return result
}

func tokenRoleFromChroma(token chroma.TokenType) TokenRole {
	switch {
	case token.InCategory(chroma.Comment):
		return TokenRoleSyntaxComment
	case token.InCategory(chroma.Keyword):
		return TokenRoleSyntaxKeyword
	case token.InCategory(chroma.LiteralString):
		return TokenRoleSyntaxString
	case token.InCategory(chroma.LiteralNumber):
		return TokenRoleSyntaxNumber
	case token.InSubCategory(chroma.NameFunction):
		return TokenRoleSyntaxFunction
	case token == chroma.NameClass || token.InSubCategory(chroma.NameBuiltin) || token == chroma.KeywordType:
		return TokenRoleSyntaxType
	case token == chroma.Punctuation || token.InCategory(chroma.Operator) || token == chroma.TextPunctuation:
		return TokenRoleSyntaxPunctuation
	default:
		return TokenRoleSyntaxPlain
	}
}
