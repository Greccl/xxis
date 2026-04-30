package parser

import "testing"

func TestSplitAndOrPreservesRanges(t *testing.T) {
	tok := &Token{
		Typ: 'C',
		Toks: []*Token{
			{Typ: 'T', Buf: []rune("cmd 1 &&  cmd 2 || cmd 3"), Start: 10, End: 34},
		},
		Start: 10,
		End:   34,
	}

	got := split_and_or(tok)
	if got.Typ != 'B' {
		t.Fatalf("split_and_or() Typ = %q, want 'B'", got.Typ)
	}
	if got.Start != 10 || got.End != 34 {
		t.Fatalf("block range = %d:%d, want 10:34", got.Start, got.End)
	}
	if len(got.Toks) != 3 {
		t.Fatalf("len(block.Toks) = %d, want 3", len(got.Toks))
	}

	want := []struct {
		typ        rune
		start, end int
		text       string
	}{
		{'C', 10, 15, "cmd 1"},
		{'&', 20, 25, "cmd 2"},
		{'/', 29, 34, "cmd 3"},
	}

	for i, want := range want {
		part := got.Toks[i]
		if part.Typ != want.typ || part.Start != want.start || part.End != want.end {
			t.Fatalf("part %d = %q %d:%d, want %q %d:%d", i, part.Typ, part.Start, part.End, want.typ, want.start, want.end)
		}
		if len(part.Toks) != 1 {
			t.Fatalf("part %d child count = %d, want 1", i, len(part.Toks))
		}
		child := part.Toks[0]
		if string(child.Buf) != want.text || child.Start != want.start || child.End != want.end {
			t.Fatalf("part %d child = %q %d:%d, want %q %d:%d", i, string(child.Buf), child.Start, child.End, want.text, want.start, want.end)
		}
	}
}

func TestSplitTokensAtPreservesRangesForSplitAndOr(t *testing.T) {
	tokens := []*Token{
		{Typ: 'T', Buf: []rune("cond && other:  body || fallback"), Start: 4, End: 36},
	}

	cond, body := split_by_colon(tokens)
	condTok := split_and_or(&Token{Typ: 'C', Toks: cond})
	bodyTok := split_and_or(&Token{Typ: 'C', Toks: body})

	if condTok.Start != 4 || condTok.End != 17 {
		t.Fatalf("cond range = %d:%d, want 4:17", condTok.Start, condTok.End)
	}
	if bodyTok.Start != 20 || bodyTok.End != 36 {
		t.Fatalf("body range = %d:%d, want 20:36", bodyTok.Start, bodyTok.End)
	}
	if len(bodyTok.Toks) != 2 {
		t.Fatalf("body child count = %d, want 2", len(bodyTok.Toks))
	}
	if bodyTok.Toks[1].Start != 28 || bodyTok.Toks[1].End != 36 {
		t.Fatalf("body second part range = %d:%d, want 28:36", bodyTok.Toks[1].Start, bodyTok.Toks[1].End)
	}
}
