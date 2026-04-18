package vt

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestScrollbackRingBuffer(t *testing.T) {
	sb := NewScrollback(5)

	// Test initial state
	if sb.Len() != 0 {
		t.Errorf("expected empty scrollback, got %d lines", sb.Len())
	}
	if sb.MaxLines() != 5 {
		t.Errorf("expected maxLines=5, got %d", sb.MaxLines())
	}

	// Push lines until full
	for i := range 5 {
		line := uv.Line{{Content: string(rune('A' + i)), Width: 1}}
		sb.PushLine(line)
	}

	if sb.Len() != 5 {
		t.Errorf("expected 5 lines, got %d", sb.Len())
	}

	// Verify ring buffer overwrites oldest
	line6 := uv.Line{{Content: "F", Width: 1}}
	sb.PushLine(line6)

	if sb.Len() != 5 {
		t.Errorf("expected 5 lines after overflow, got %d", sb.Len())
	}

	// First line should now be 'B' (oldest 'A' was dropped)
	first := sb.Line(0)
	if first == nil || first[0].Content != "B" {
		t.Errorf("expected first line to be 'B', got %v", first)
	}

	// Last line should be 'F'
	last := sb.Line(4)
	if last == nil || last[0].Content != "F" {
		t.Errorf("expected last line to be 'F', got %v", last)
	}
}

func TestScrollbackSoftWrapping(t *testing.T) {
	sb := NewScrollback(10)

	// Push soft-wrapped line
	line1 := uv.Line{{Content: "A", Width: 1}}
	sb.PushLineWithWrap(line1, true)

	// Push hard break line
	line2 := uv.Line{{Content: "B", Width: 1}}
	sb.PushLineWithWrap(line2, false)

	if sb.Len() != 2 {
		t.Errorf("expected 2 lines, got %d", sb.Len())
	}

	// Verify lines were stored
	if l := sb.Line(0); l == nil || l[0].Content != "A" {
		t.Errorf("line 0 incorrect")
	}
	if l := sb.Line(1); l == nil || l[0].Content != "B" {
		t.Errorf("line 1 incorrect")
	}
}

func TestScrollbackClear(t *testing.T) {
	sb := NewScrollback(10)

	for i := range 5 {
		line := uv.Line{{Content: string(rune('A' + i)), Width: 1}}
		sb.PushLine(line)
	}

	sb.Clear()

	if sb.Len() != 0 {
		t.Errorf("expected empty after clear, got %d lines", sb.Len())
	}

	// Should be able to push after clear
	line := uv.Line{{Content: "X", Width: 1}}
	sb.PushLine(line)

	if sb.Len() != 1 {
		t.Errorf("expected 1 line after push, got %d", sb.Len())
	}
}

func TestScrollbackSetMaxLines(t *testing.T) {
	sb := NewScrollback(10)

	// Fill with 8 lines
	for i := range 8 {
		line := uv.Line{{Content: string(rune('A' + i)), Width: 1}}
		sb.PushLine(line)
	}

	// Reduce max to 5 (should keep last 5: D,E,F,G,H)
	sb.SetMaxLines(5)

	if sb.Len() != 5 {
		t.Errorf("expected 5 lines after resize, got %d", sb.Len())
	}

	if sb.MaxLines() != 5 {
		t.Errorf("expected maxLines=5, got %d", sb.MaxLines())
	}

	// First line should be 'D' (oldest 3 dropped)
	first := sb.Line(0)
	if first == nil || first[0].Content != "D" {
		t.Errorf("expected first line to be 'D', got %v", first)
	}

	// Last line should be 'H'
	last := sb.Line(4)
	if last == nil || last[0].Content != "H" {
		t.Errorf("expected last line to be 'H', got %v", last)
	}
}

func TestScrollbackBoundsChecking(t *testing.T) {
	sb := NewScrollback(5)

	line := uv.Line{{Content: "A", Width: 1}}
	sb.PushLine(line)

	// Out of bounds access should return nil
	if sb.Line(-1) != nil {
		t.Error("expected nil for negative index")
	}
	if sb.Line(100) != nil {
		t.Error("expected nil for out of bounds index")
	}

	// Valid access should work
	if sb.Line(0) == nil {
		t.Error("expected valid line at index 0")
	}
}

func TestScrollbackWidthTracking(t *testing.T) {
	sb := NewScrollback(10)

	if sb.CaptureWidth() != 0 {
		t.Errorf("expected initial width 0, got %d", sb.CaptureWidth())
	}

	sb.SetCaptureWidth(80)
	if sb.CaptureWidth() != 80 {
		t.Errorf("expected width 80, got %d", sb.CaptureWidth())
	}

	// Reflow should update width
	sb.Reflow(100)
	if sb.CaptureWidth() != 100 {
		t.Errorf("expected width 100 after reflow, got %d", sb.CaptureWidth())
	}
}

func TestScrollbackPopNewest(t *testing.T) {
	sb := NewScrollback(10)

	// Empty buffer: PopNewest returns nil.
	if got := sb.PopNewest(3); got != nil {
		t.Errorf("expected nil from empty buffer, got %v", got)
	}

	// Push A..E.
	for i := range 5 {
		line := uv.Line{{Content: string(rune('A' + i)), Width: 1}}
		sb.PushLine(line)
	}

	// Pop the 2 newest lines (D, E) in chronological order.
	got := sb.PopNewest(2)
	if len(got) != 2 {
		t.Fatalf("expected 2 popped lines, got %d", len(got))
	}
	if got[0][0].Content != "D" || got[1][0].Content != "E" {
		t.Errorf("expected [D E], got [%v %v]", got[0][0].Content, got[1][0].Content)
	}
	if sb.Len() != 3 {
		t.Errorf("expected 3 remaining lines, got %d", sb.Len())
	}
	// Remaining lines should be A, B, C in order.
	for i, want := range []string{"A", "B", "C"} {
		if l := sb.Line(i); l == nil || l[0].Content != want {
			t.Errorf("line %d: expected %q, got %v", i, want, l)
		}
	}

	// Popping more than available returns what is available.
	got = sb.PopNewest(10)
	if len(got) != 3 {
		t.Fatalf("expected 3 popped lines, got %d", len(got))
	}
	if sb.Len() != 0 {
		t.Errorf("expected empty buffer, got %d", sb.Len())
	}

	// n<=0 returns nil.
	if got := sb.PopNewest(0); got != nil {
		t.Errorf("expected nil for n=0, got %v", got)
	}
	if got := sb.PopNewest(-1); got != nil {
		t.Errorf("expected nil for negative n, got %v", got)
	}
}

func TestScrollbackPopNewestFullRingBuffer(t *testing.T) {
	sb := NewScrollback(5)

	// Push 7 lines so the ring buffer wraps around (A and B are dropped).
	for i := range 7 {
		line := uv.Line{{Content: string(rune('A' + i)), Width: 1}}
		sb.PushLine(line)
	}
	// Buffer now holds C, D, E, F, G.
	if sb.Len() != 5 {
		t.Fatalf("expected full buffer of 5, got %d", sb.Len())
	}

	got := sb.PopNewest(2)
	if len(got) != 2 || got[0][0].Content != "F" || got[1][0].Content != "G" {
		t.Fatalf("expected [F G], got %v", got)
	}
	if sb.Len() != 3 {
		t.Errorf("expected 3 remaining lines, got %d", sb.Len())
	}
	// Remaining should be C, D, E.
	for i, want := range []string{"C", "D", "E"} {
		if l := sb.Line(i); l == nil || l[0].Content != want {
			t.Errorf("line %d: expected %q, got %v", i, want, l)
		}
	}
	// Subsequent pushes should still work correctly.
	sb.PushLine(uv.Line{{Content: "H", Width: 1}})
	if l := sb.Line(3); l == nil || l[0].Content != "H" {
		t.Errorf("expected H after push, got %v", l)
	}
}

func TestScrollbackEmptyPushIgnored(t *testing.T) {
	sb := NewScrollback(5)

	// Empty line should be ignored
	sb.PushLine(uv.Line{})

	if sb.Len() != 0 {
		t.Errorf("expected empty scrollback after pushing empty line, got %d", sb.Len())
	}
}
