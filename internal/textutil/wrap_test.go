package textutil

import (
	"reflect"
	"strings"
	"testing"
)

func TestHardWrap(t *testing.T) {
	for _, test := range []struct {
		name    string
		content string
		width   int
		want    []string
	}{
		{"empty", "", 4, []string{""}},
		{"unbounded", "abcd", 0, []string{"abcd"}},
		{"narrow_wide", "你好", 1, []string{"你", "好"}},
		{"wide_after_ASCII", "ab你c", 3, []string{"ab", "你c"}},
		{"ANSI", "\x1b[31mabcdef\x1b[0m", 3, []string{"\x1b[31mabc", "def\x1b[0m"}},
		{"ANSI_at_boundary", "ab\x1b[31mcd\x1b[0m", 2, []string{"ab\x1b[31m", "cd\x1b[0m"}},
		{"control", "ab\tcd", 2, []string{"ab\t", "cd"}},
		{"keycap", "1️⃣2️⃣", 2, []string{"1️⃣", "2️⃣"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := HardWrap(test.content, test.width)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("HardWrap() = %q, want %q", got, test.want)
			}
			if strings.Join(got, "") != test.content {
				t.Fatal("wrapping changed the input bytes")
			}
		})
	}
}

func FuzzHardWrap(f *testing.F) {
	for _, seed := range []string{"hello world", "你好", "e\u0301👨‍👩‍👧‍👦", "1️⃣2️⃣", "\x1b[31mred\x1b[0m", "\x1b[", "\x1b[\xffabc", "\xff\xfe"} {
		f.Add(seed, uint8(4))
	}
	f.Fuzz(func(t *testing.T, content string, width uint8) {
		lines := HardWrap(content, int(width)+1)
		if strings.Join(lines, "") != content {
			t.Fatal("wrapping changed the input bytes")
		}
	})
}
